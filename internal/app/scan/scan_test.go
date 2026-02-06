// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package scan_test

import (
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/testshared"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExistingFile(t *testing.T) {
	tempFile, err := random.CreateTempFile("", "unit-testing", 1)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile

	err = scan.Run(cfg)
	assert.ErrorContains(t, err, "file already exists at")
}

func TestOverrideExistingFile(t *testing.T) {
	tempFile, err := random.CreateTempFile("", "unit-testing", 1)
	require.NoError(t, err)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.ForceOverride = true

	err = scan.Run(cfg)
	assert.NoError(t, err)
}

func TestScan(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile

	err := scan.Run(cfg)
	require.NoError(t, err)

	// Validate
	paths, err := testshared.DatabasePaths(cfg.DbPath)
	require.NoError(t, err)

	expPaths, err := testshared.ExpectedPaths(cfg.Root, nil)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, paths)
}

func TestScanEmptyDir(t *testing.T) {
	scanDir, err := os.MkdirTemp("", "test-empty")
	require.NoError(t, err)

	tempFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.Root = scanDir

	err = scan.Run(cfg)
	require.NoError(t, err)

	paths, err := testshared.DatabasePaths(cfg.DbPath)
	require.NoError(t, err)
	// Expect the root dir to be in the database and which is relative to itself "."
	require.Len(t, paths, 1)
	assert.Equal(t, ".", paths[0].Path)
}

func TestScanWithHashes(t *testing.T) {
	testCases := []struct {
		algo         ajhash.Algo
		hashDeepFile string
	}{
		{
			algo:         ajhash.AlgoSHA1,
			hashDeepFile: "../../testdata/expected/scan.sha1",
		},
		{
			algo:         ajhash.AlgoSHA256,
			hashDeepFile: "../../testdata/expected/scan.sha256",
		},
		// Can't test SHA-512 atm because hashdeep doesn't support it
	}
	for _, tC := range testCases {
		t.Run(tC.algo.String(), func(t *testing.T) {
			algo := tC.algo

			tempFile := filepath.Join(t.TempDir(), "unit-testing")
			_ = os.Remove(tempFile)
			defer os.Remove(tempFile)

			cfg := initialConfig()
			cfg.DbPath = tempFile
			cfg.CalculateHashes = true
			cfg.Algo = algo

			err := scan.Run(cfg)
			require.NoError(t, err)

			// Validate
			paths, err := testshared.DatabasePaths(cfg.DbPath)
			require.NoError(t, err)

			expPaths, err := testshared.ExpectedPaths(cfg.Root, nil)
			require.NoError(t, err)

			assert.ElementsMatch(t, expPaths, paths)

			expectedHashDeep, err := testshared.ReadHashDeepFile(tC.hashDeepFile)
			require.NoError(t, err)

			// Map from path to hash string
			exp := make(map[string]string, len(expectedHashDeep))
			for _, hd := range expectedHashDeep {
				exp[hd.Path] = hd.Hash
			}

			dbf, err := db.OpenDatabase(cfg.DbPath)
			require.NoError(t, err)
			defer dbf.Close()

			ht, err := dbf.ReadHashTable()
			require.NoError(t, err)

			result := make(map[string]string, len(ht))
			for k, v := range ht {
				pi, err := dbf.ReadEntryAtIndex(k)
				require.NoError(t, err)
				hash := hex.EncodeToString(v)
				result[pi.Path] = hash
			}

			assert.Equal(t, exp, result)
		})

	}
}

func TestScanInitOnly(t *testing.T) {
	testCases := []struct {
		algo ajhash.Algo
	}{
		{
			algo: ajhash.AlgoSHA1,
		},
		{
			algo: ajhash.AlgoSHA256,
		},
		{
			algo: ajhash.AlgoSHA512,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.algo.String(), func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), "unit-testing")
			_ = os.Remove(tempFile)
			defer os.Remove(tempFile)

			cfg := initialConfig()
			cfg.DbPath = tempFile
			cfg.CalculateHashes = true
			cfg.Algo = tC.algo
			cfg.InitOnly = true

			err := scan.Run(cfg)
			require.NoError(t, err)

			// Verify
			dbf, err := db.OpenDatabase(tempFile)
			require.NoError(t, err)
			defer dbf.Close()

			require.True(t, dbf.Features().HasHashTable())
			algo, err := dbf.HashTableAlgo()
			require.NoError(t, err)
			assert.Equal(t, tC.algo, algo)

			ht, err := dbf.ReadHashTable()
			require.NoError(t, err)
			assert.Empty(t, ht)
		})
	}
}

//-----------------------------------------------------------------------------

func initialConfig() scan.Config {
	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Root: "../../testdata/scan",
	}
	return cfg
}
