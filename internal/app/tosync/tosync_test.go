package tosync_test

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/diff"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/app/tosync"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToSync(t *testing.T) {
	aPath := filepath.Join("testdata", "../../../testdata/need-sync/a")
	bPath := filepath.Join("testdata", "../../../testdata/need-sync/b")

	lhsPath, rhsPath, err := makeTwoDatabases(aPath, bPath, false, false)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(lhsPath)
		_ = os.Remove(rhsPath)
	}()

	cfg := tosync.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		RhsPath: rhsPath,
	}

	result := make([]string, 0, 2)

	cfg.Fn = func(d diff.Diff) error {
		result = append(result, d.Path)
		return nil
	}

	require.NoError(t, tosync.Run(cfg))

	expected := []string{
		"blank.txt",
		"cached/2.txt",
	}

	slices.Sort(result)
	slices.Sort(expected)

	assert.Equal(t, expected, result)
}

func TestToSyncNothing(t *testing.T) {
	aPath := filepath.Join("testdata", "../../../testdata/need-sync/a")

	lhsPath, rhsPath, err := makeTwoDatabases(aPath, aPath, false, false)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(lhsPath)
		_ = os.Remove(rhsPath)
	}()

	cfg := tosync.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		RhsPath: rhsPath,
	}

	cfg.Fn = func(d diff.Diff) error {
		require.Fail(t, "there should be nothing to sync")
		return nil
	}

	require.NoError(t, tosync.Run(cfg))
}

func TestToSyncOnlyHashes(t *testing.T) {
	aPath := filepath.Join("testdata", "../../../testdata/need-sync/a")
	bPath := filepath.Join("testdata", "../../../testdata/need-sync/c")

	lhsPath, rhsPath, err := makeTwoDatabases(aPath, bPath, true, false)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(lhsPath)
		_ = os.Remove(rhsPath)
	}()

	cfg := tosync.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath:    lhsPath,
		RhsPath:    rhsPath,
		OnlyHashes: true,
	}

	result := make([]string, 0, 2)

	cfg.Fn = func(d diff.Diff) error {
		result = append(result, d.Path)
		return nil
	}

	require.NoError(t, tosync.Run(cfg))

	expected := []string{
		"blank.txt",
	}

	slices.Sort(result)
	slices.Sort(expected)

	assert.Equal(t, expected, result)
}

func TestToSyncOnlyHashesWithDifferentAlgos(t *testing.T) {
	aPath := filepath.Join("testdata", "../../../testdata/need-sync/a")
	bPath := filepath.Join("testdata", "../../../testdata/need-sync/b")

	lhsPath, rhsPath, err := makeTwoDatabases(aPath, bPath, true, true)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(lhsPath)
		_ = os.Remove(rhsPath)
	}()

	cfg := tosync.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath:    lhsPath,
		RhsPath:    rhsPath,
		OnlyHashes: true,
	}

	cfg.Fn = func(d diff.Diff) error {
		return nil
	}

	require.ErrorContains(t, tosync.Run(cfg), "can't compare the two databases")
}

//-----------------------------------------------------------------------------

func makeTwoDatabases(scanA string, scanB string, hashes bool, differentAlgos bool) (string, string, error) {
	lhsPath := filepath.Join(os.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: scanA,
	}

	if hashes {
		cfg.CalculateHashes = true
		cfg.Algo = ajhash.AlgoSHA1
	}

	if err := scan.Run(cfg); err != nil {
		_ = os.Remove(lhsPath)
		return "", "", err
	}

	rhsPath := filepath.Join(os.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = scanB

	if differentAlgos {
		cfg.Algo = ajhash.AlgoSHA256
	}

	if err := scan.Run(cfg); err != nil {
		_ = os.Remove(lhsPath)
		_ = os.Remove(rhsPath)
		return "", "", err
	}

	return lhsPath, rhsPath, nil
}
