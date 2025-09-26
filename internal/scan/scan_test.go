package scan_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scan"
	"github.com/andrejacobs/go-aj/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile) // delete if it already exists
	defer os.Remove(tempFile)

	// Create new database
	dbf, err := db.CreateDatabase(tempFile, dataDir)
	require.NoError(t, err)

	// Perform the scan
	s := scan.NewScanner()
	require.NoError(t, s.Scan(context.Background(), dbf))

	// Close database
	require.NoError(t, dbf.Close())

	// Validate
	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	w := file.NewWalker()
	w.DirExcluder = s.DirExcluder
	w.FileExcluder = s.FileExcluder

	count := 0
	err = w.Walk(dataDir, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		require.NoError(t, rcvErr)

		relPath, err := filepath.Rel(dataDir, rcvPath)
		if err != nil {
			return err
		}

		expInfo, err := path.InfoFromWalk(relPath, d)
		require.NoError(t, err)

		info, err := dbf.ReadEntryAtIndex(count)
		require.NoError(t, err)

		assert.True(t, expInfo.Equals(&info))

		count += 1
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, count, dbf.EntriesCount())
}

func TestScanCancelled(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database
	dbf, err := db.CreateDatabase(tempFile, dataDir)
	require.NoError(t, err)

	// Perform the scan
	s := scan.NewScanner()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = s.Scan(ctx, dbf)
	require.ErrorIs(t, err, context.Canceled)
}

//-----------------------------------------------------------------------------

// func TestLocalScan(t *testing.T) {
// 	tempFile := "/Users/andre/temp/test.ajfs"
// 	_ = os.Remove(tempFile)

// 	// Create new database
// 	localDir := "/Users/andre/TODO_SORT_OUT" //+/- 200GB
// 	dbf, err := db.CreateDatabase(tempFile, localDir)
// 	require.NoError(t, err)

// 	// Perform the scan
// 	s := scan.NewScanner()
// 	require.NoError(t, s.Scan(context.Background(), dbf))

// 	// Close database
// 	require.NoError(t, dbf.Close())

// 	// Validate
// 	dbf, err = db.OpenDatabase(tempFile)
// 	require.NoError(t, err)
// 	defer dbf.Close()

// 	w := file.NewWalker()
// 	w.DirExcluder = s.DirExcluder
// 	w.FileExcluder = s.FileExcluder

// 	count := 0
// 	err = w.Walk(localDir, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
// 		require.NoError(t, rcvErr)

// 		relPath, err := filepath.Rel(localDir, rcvPath)
// 		if err != nil {
// 			return err
// 		}

// 		expInfo, err := path.InfoFromWalk(relPath, d)
// 		require.NoError(t, err)

// 		info, err := dbf.ReadEntryAtIndex(count)
// 		require.NoError(t, err)

// 		if !expInfo.Equals(&info) {
// 			fmt.Printf("e: %+v\n", expInfo)
// 			fmt.Printf("a: %+v\n", info)
// 		}
// 		assert.True(t, expInfo.Equals(&info))

// 		count += 1
// 		return nil
// 	})
// 	require.NoError(t, err)

// 	assert.Equal(t, count, dbf.EntriesCount())
// }

const (
	dataDir = "../testdata/scan"
)
