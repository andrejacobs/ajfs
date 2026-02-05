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

	absRoot, err := filepath.Abs(scanCfg.Root)
	require.NoError(t, err)

	expOut1 := fmt.Sprintf(`Database path: %s
Version:       %d
Root path:     %s
Tool:          %s
OS:            %s
Architecture:  %s`,
		tempFile,
		1,
		absRoot,
		"ajfs: v0.0.0 ",
		runtime.GOOS,
		runtime.GOARCH)

	expOut2 := fmt.Sprintf(`Entries:       %d
File size:     %s`,
		exp.entries,
		human.Bytes(uint64(fileInfo.Size())))

	expOut3 := fmt.Sprintf(`Calculating statistics...
File count:    %d
Dir count:     %d
Total size:    %s [all files together]
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

	assert.Equal(t, "", errBuffer.String())
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
