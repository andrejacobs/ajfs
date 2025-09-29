package info_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/info"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/file"
	"github.com/andrejacobs/go-aj/human"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfo(t *testing.T) {
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

	cfg := info.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
	}

	err = info.Run(cfg)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(cfg.DbPath)
	require.NoError(t, err)

	exp, err := expected(scanCfg.Root)
	require.NoError(t, err)

	expOut1 := fmt.Sprintf(`Database path: %s
Version:       %d
Root path:     %s
OS:            %s
Architecture:  %s`,
		tempFile,
		1,
		scanCfg.Root,
		runtime.GOOS,
		runtime.GOARCH)

	expOut2 := fmt.Sprintf(`Entries:       %d
File size:     %s`,
		exp.entries,
		human.Bytes(uint64(fileInfo.Size())))

	expOut3 := fmt.Sprintf(`Calculating statistics...
File count:    %d
Dir count:     %d
Total size:    %s [all files toghether]
Max file size: %s [single biggest file]
Avg file size: %s`,
		exp.fileCount,
		exp.dirCount,
		human.Bytes(exp.totalSize),
		human.Bytes(exp.maxFileSize),
		human.Bytes(exp.avgFileSize))

	outStr := outBuffer.String()

	assert.Contains(t, outStr, expOut1)
	assert.Contains(t, outStr, expOut2)
	assert.Contains(t, outStr, expOut3)
}

//-----------------------------------------------------------------------------

type expectedResults struct {
	entries     int
	fileCount   int
	dirCount    int
	totalSize   uint64
	maxFileSize uint64
	avgFileSize uint64
}

func expected(scanDir string) (expectedResults, error) {
	w := file.NewWalker()
	w.FileExcluder = scanner.DefaultFileExcluder()

	result := expectedResults{}

	err := w.Walk(scanDir, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		result.entries++

		if d.IsDir() {
			result.dirCount++
		} else if d.Type().IsRegular() {
			result.fileCount++
			fileInfo, err := d.Info()
			if err != nil {
				return err
			}
			result.totalSize += uint64(fileInfo.Size())
			result.maxFileSize = max(result.maxFileSize, uint64(fileInfo.Size()))
		}

		return nil
	})

	result.avgFileSize = result.totalSize / uint64(result.fileCount)
	return result, err
}
