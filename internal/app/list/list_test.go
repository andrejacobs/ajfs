package list_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/list"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/scan",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cfg := list.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
	}

	err = list.Run(cfg)
	assert.NoError(t, err)

	exp, err := expected(scanCfg.Root, cfg.DisplayFullPaths)
	require.NoError(t, err)

	assert.Equal(t, exp, outBuffer.String())
	assert.Equal(t, "", errBuffer.String())

	// Full paths
	outBuffer.Reset()
	cfg.DisplayFullPaths = true
	err = list.Run(cfg)
	assert.NoError(t, err)

	exp, err = expected(scanCfg.Root, cfg.DisplayFullPaths)
	require.NoError(t, err)

	assert.Equal(t, exp, outBuffer.String())
	assert.Equal(t, "", errBuffer.String())

	// Verbose
	outBuffer.Reset()
	cfg.CommonConfig.Verbose = true

	err = list.Run(cfg)
	assert.NoError(t, err)
	assert.Contains(t, outBuffer.String(), path.Header())
}

func TestListWithHashes(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root:            "../../testdata/scan",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cfg := list.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
		DisplayHashes: true,
	}

	err = list.Run(cfg)
	assert.NoError(t, err)

	scanner := bufio.NewScanner(&outBuffer)
	for scanner.Scan() {
		assert.Len(t, strings.Split(scanner.Text(), ","), 6)
	}

	assert.Equal(t, "", errBuffer.String())

	// Verbose
	outBuffer.Reset()
	cfg.CommonConfig.Verbose = true

	err = list.Run(cfg)
	assert.NoError(t, err)
	assert.Contains(t, outBuffer.String(), path.HeaderWithHash())
}

func expected(scanDir string, fullPaths bool) (string, error) {
	w := file.NewWalker()
	w.FileExcluder = scanner.DefaultFileExcluder()

	scanDir, err := filepath.Abs(scanDir)
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer

	err = w.Walk(scanDir, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		relPath, err := filepath.Rel(scanDir, rcvPath)
		if err != nil {
			return err
		}

		expInfo, err := path.InfoFromWalk(relPath, d)
		if err != nil {
			return err
		}

		if fullPaths {
			expInfo.Path = filepath.Join(scanDir, expInfo.Path)
		}

		fmt.Fprintln(&buffer, expInfo)

		return nil
	})

	return buffer.String(), err
}
