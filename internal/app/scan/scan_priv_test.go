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

package scan

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/resume"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/testshared"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanWithErrors(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.CalculateHashes = true
	cfg.Algo = ajhash.AlgoSHA1

	// Cause an error while scanning
	cfg.simulateScanningError = true

	var err error
	require.NotPanics(t, func() {
		err = Run(cfg)
	})

	require.ErrorContains(t, err, "simulating an error while scanning")
}

func TestScanWithHashingErrors(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.CalculateHashes = true
	cfg.Algo = ajhash.AlgoSHA1

	// Cause an error while hashing
	cfg.simulateHashingError = true

	err := Run(cfg)
	require.Error(t, err)

	// Validate: Expect the database to still be valid
	paths, err := testshared.DatabasePaths(cfg.DbPath)
	require.NoError(t, err)

	expPaths, err := testshared.ExpectedPaths(cfg.Root, nil)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, paths)
}

func TestScanWithHashingErrorsShouldBeAbleToContinue(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.CalculateHashes = true
	cfg.Algo = ajhash.AlgoSHA1

	// Cause an error while hashing
	const expErrMsg = "simulating a file hashing that failed"
	count := 0
	cfg.hashFn = func(ctx context.Context, path string, hasher hash.Hash, w io.Writer) ([]byte, uint64, error) {
		count++
		if count == 3 || count == 7 {
			return nil, 0, fmt.Errorf(expErrMsg)
		}
		return file.Hash(ctx, path, hasher, w)
	}

	var errOutput bytes.Buffer
	cfg.Stderr = &errOutput

	err := Run(cfg)
	require.NoError(t, err)

	require.Contains(t, errOutput.String(), expErrMsg)

	// Validate: Expect the database to still be valid
	paths, err := testshared.DatabasePaths(cfg.DbPath)
	require.NoError(t, err)

	expPaths, err := testshared.ExpectedPaths(cfg.Root, nil)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, paths)

	// Count incomplete hashes
	dbf, err := db.OpenDatabase(cfg.DbPath)
	require.NoError(t, err)

	count = 0
	err = dbf.ReadHashTableEntries(func(idx int, hash []byte) error {
		if ajhash.AllZeroBytes(hash) {
			count++
		}
		return nil
	})
	require.NoError(t, dbf.Close())
	require.NoError(t, err)
	require.Equal(t, 2, count)

	// Resume
	cfg.Stderr = io.Discard
	err = resume.Run(resume.Config{CommonConfig: cfg.CommonConfig})
	require.NoError(t, err)

	dbf, err = db.OpenDatabase(cfg.DbPath)
	require.NoError(t, err)
	defer dbf.Close()

	count = 0
	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		count++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 0, count)
}

func initialConfig() Config {
	cfg := Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Root: "../../testdata/scan",
	}
	return cfg
}
