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

package resume

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
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
	"github.com/stretchr/testify/require"
)

func TestResumeWithHashingErrors(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create initial database
	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			DbPath: tempFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Root:            "../../testdata/scan",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
		InitOnly:        true,
	}

	err := scan.Run(cfg)
	require.NoError(t, err)

	// Resume calculating hashes
	resumeCfg := Config{
		CommonConfig: cfg.CommonConfig,
	}

	// Cause an error while hashing
	const expErrMsg = "simulating a file hashing that failed"
	count := 0
	resumeCfg.hashFn = func(ctx context.Context, path string, hasher hash.Hash, w io.Writer) ([]byte, uint64, error) {
		count++
		if count == 3 || count == 7 {
			return nil, 0, fmt.Errorf(expErrMsg)
		}
		return file.Hash(ctx, path, hasher, w)
	}

	// Resume
	var errOutput bytes.Buffer
	resumeCfg.Stderr = &errOutput
	err = Run(resumeCfg)
	require.NoError(t, err)
	require.Contains(t, errOutput.String(), expErrMsg)

	// Check incomplete hashes
	dbf, err := db.OpenDatabase(cfg.DbPath)
	require.NoError(t, err)

	count = 0
	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		count++
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, count)
	require.NoError(t, dbf.Close())

	// Resume without errors
	resumeCfg.hashFn = nil
	err = Run(resumeCfg)
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
