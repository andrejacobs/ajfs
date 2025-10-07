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
	dbFile := filepath.Join(os.TempDir(), "unit-testing")
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
	dbFile := filepath.Join(os.TempDir(), "unit-testing")
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
	tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.hashdeep")
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
