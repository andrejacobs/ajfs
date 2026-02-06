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

package update_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/export"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/app/update"
	"github.com/andrejacobs/ajfs/internal/filter"
	"github.com/andrejacobs/ajfs/internal/testshared"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(dbFile)
	defer os.Remove(dbFile)

	// Create database
	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			DbPath: dbFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Root: "../../testdata/scan",
	}

	// Filter out some files
	exclF, _, err := filter.ParsePathRegexToMatchPathFn([]string{"f:blank\\.txt$"}, false)
	require.NoError(t, err)
	scanCfg.FileExcluder = file.MatchAppleDSStore(exclF)
	require.NoError(t, scan.Run(scanCfg))

	// Update (without filtering)
	updateCfg := update.Config{
		CommonConfig: scanCfg.CommonConfig,
	}
	require.NoError(t, update.Run(updateCfg))

	expPaths, err := testshared.ExpectedPaths(scanCfg.Root, nil)
	require.NoError(t, err)

	dbPaths, err := testshared.DatabasePaths(scanCfg.DbPath)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, dbPaths)
}

func TestUpdateWithHashes(t *testing.T) {
	dbFile := filepath.Join(t.TempDir(), "unit-testing")
	_ = os.Remove(dbFile)
	defer os.Remove(dbFile)

	// Create database
	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			DbPath: dbFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Root:            "../../testdata/scan",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}

	// Filter out some files
	exclF, _, err := filter.ParsePathRegexToMatchPathFn([]string{"f:blank\\.txt$"}, false)
	require.NoError(t, err)
	scanCfg.FileExcluder = file.MatchAppleDSStore(exclF)
	require.NoError(t, scan.Run(scanCfg))

	// Update (without filtering)
	updateCfg := update.Config{
		CommonConfig: scanCfg.CommonConfig,
	}
	require.NoError(t, update.Run(updateCfg))

	expPaths, err := testshared.ExpectedPaths(scanCfg.Root, nil)
	require.NoError(t, err)

	dbPaths, err := testshared.DatabasePaths(scanCfg.DbPath)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, dbPaths)

	// Export and check hashes
	tempExportFile := filepath.Join(t.TempDir(), "unit-test.ajfs.hashdeep")
	_ = os.Remove(tempExportFile)
	defer os.Remove(tempExportFile)

	exportCfg := export.Config{
		CommonConfig: config.CommonConfig{
			DbPath: dbFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Format:     export.FormatHashdeep,
		ExportPath: tempExportFile,
	}

	require.NoError(t, export.Run(exportCfg))

	// Validate
	expectedHashDeep, err := testshared.ReadHashDeepFile("../../testdata/expected/scan.sha1")
	require.NoError(t, err)

	exportedHashDeep, err := testshared.ReadHashDeepFile(tempExportFile)
	require.NoError(t, err)

	assert.ElementsMatch(t, expectedHashDeep, exportedHashDeep)
}
