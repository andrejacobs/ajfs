package scan_test

import (
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/ajfs/internal/testshared"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
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
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile

	err := scan.Run(cfg)
	require.NoError(t, err)

	// Validate
	paths, err := databasePaths(cfg)
	require.NoError(t, err)

	expPaths, err := expectedPaths(cfg)
	require.NoError(t, err)

	assert.ElementsMatch(t, expPaths, paths)
}

func TestScanEmptyDir(t *testing.T) {
	scanDir, err := os.MkdirTemp("", "test-empty")
	require.NoError(t, err)

	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	cfg := initialConfig()
	cfg.DbPath = tempFile
	cfg.Root = scanDir

	err = scan.Run(cfg)
	require.NoError(t, err)

	paths, err := databasePaths(cfg)
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

			tempFile := filepath.Join(os.TempDir(), "unit-testing")
			_ = os.Remove(tempFile)
			defer os.Remove(tempFile)

			cfg := initialConfig()
			cfg.DbPath = tempFile
			cfg.CalculateHashes = true
			cfg.Algo = algo

			err := scan.Run(cfg)
			require.NoError(t, err)

			// Validate
			paths, err := databasePaths(cfg)
			require.NoError(t, err)

			expPaths, err := expectedPaths(cfg)
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
			tempFile := filepath.Join(os.TempDir(), "unit-testing")
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

func expectedPaths(cfg scan.Config) ([]path.Info, error) {
	w := file.NewWalker()
	w.FileExcluder = scanner.DefaultFileExcluder()

	result := make([]path.Info, 0, 32)

	err := w.Walk(cfg.Root, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		relPath, err := filepath.Rel(cfg.Root, rcvPath)
		if err != nil {
			return err
		}

		expInfo, err := path.InfoFromWalk(relPath, d)
		if err != nil {
			return err
		}

		result = append(result, expInfo)

		return nil
	})

	return result, err
}

func databasePaths(cfg scan.Config) ([]path.Info, error) {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return nil, err
	}
	defer dbf.Close()

	result := make([]path.Info, 0, 32)

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		result = append(result, pi)
		return nil
	})

	return result, err
}
