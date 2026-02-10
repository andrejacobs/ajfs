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

package db_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateStats(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database and write N path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureJustEntries)
	require.NoError(t, err)

	expCount := 10
	expTime := time.Now().Add(-10 * time.Minute)

	expStats := db.Stats{}

	for i := range expCount {
		filePath := fmt.Sprintf("/some/path/%d.txt", i)
		p := path.Info{
			Id:      path.IdFromPath(filePath),
			Path:    filePath,
			Size:    uint64(random.Int(10, 4242)),
			Mode:    0740,
			ModTime: expTime,
		}
		if i == 3 || i == 7 {
			p.Mode |= fs.ModeDir
			expStats.DirCount++
		} else {
			expStats.FileCount++
			expStats.TotalFileSize += p.Size
			expStats.MaxFileSize = max(expStats.MaxFileSize, p.Size)
		}
		require.NoError(t, dbf.WriteEntry(&p))
	}

	expStats.AvgFileSize = expStats.TotalFileSize / expStats.FileCount

	require.NoError(t, dbf.FinishEntries())
	require.NoError(t, dbf.Close())

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	stats, err := dbf.CalculateStats()
	require.NoError(t, err)

	assert.Equal(t, expStats, stats)
}

func TestCalculateStatsWhenEmpty(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create a new empty database
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureJustEntries)
	require.NoError(t, err)
	require.NoError(t, dbf.FinishEntries())
	require.NoError(t, dbf.Close())

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	stats, err := dbf.CalculateStats()
	require.NoError(t, err)
	require.Equal(t, db.Stats{}, stats)
}
