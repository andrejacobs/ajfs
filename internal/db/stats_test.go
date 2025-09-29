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
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
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
