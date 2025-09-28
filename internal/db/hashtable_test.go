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
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO: empty database, database with only initial hash table

func TestWriteInitialHashTable(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database and write path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureHashTable)
	require.NoError(t, err)

	p1 := path.Info{
		Id:      path.IdFromPath("a.txt"),
		Path:    "a.txt",
		Size:    uint64(42),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p1))

	p2 := path.Info{
		Id:      path.IdFromPath("some/dir"),
		Path:    "some/dir",
		Size:    uint64(142),
		Mode:    0644 | fs.ModeDir,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p2))

	p3 := path.Info{
		Id:      path.IdFromPath("c.txt"),
		Path:    "c.txt",
		Size:    uint64(442),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p3))

	require.NoError(t, dbf.FinishEntries())

	algo := ajhash.AlgoSHA1
	// Not writing any hash values, just the empty hash table
	assert.NoError(t, dbf.StartHashTable(algo))
	assert.NoError(t, dbf.FinishHashTable())
	assert.NoError(t, dbf.Close())

	// Open and validate
	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()
	require.Equal(t, 3, dbf.EntriesCount())
	require.Equal(t, 2, dbf.FileEntriesCount())

	assert.True(t, dbf.Features().HasHashTable())

	count := 0
	zeroValue := algo.ZeroValue()
	fn := func(idx int, hash []byte) error {
		count++
		assert.Equal(t, algo.Size(), len(hash))
		assert.Equal(t, zeroValue, hash)
		return nil
	}
	assert.NoError(t, dbf.ReadHashTableEntries(fn))
	assert.Equal(t, dbf.FileEntriesCount(), count)
}

func TestWriteHashTable(t *testing.T) {
	tempFile := "/Users/andre/temp/test.ajfs"
	// tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	// defer os.Remove(tempFile)

	// Create new database and write path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureHashTable)
	require.NoError(t, err)

	p1 := path.Info{
		Id:      path.IdFromPath("a.txt"),
		Path:    "a.txt",
		Size:    uint64(42),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p1))

	p2 := path.Info{
		Id:      path.IdFromPath("some/dir"),
		Path:    "some/dir",
		Size:    uint64(142),
		Mode:    0644 | fs.ModeDir,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p2))

	p3 := path.Info{
		Id:      path.IdFromPath("c.txt"),
		Path:    "c.txt",
		Size:    uint64(442),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p3))

	require.NoError(t, dbf.FinishEntries())

	algo := ajhash.AlgoSHA1
	assert.NoError(t, dbf.StartHashTable(algo))

	h1 := make([]byte, algo.Size())
	require.NoError(t, random.SecureBytes(h1))
	require.NoError(t, dbf.WriteHashEntry(0, h1))

	h2 := make([]byte, algo.Size())
	require.NoError(t, random.SecureBytes(h2))
	require.NoError(t, dbf.WriteHashEntry(2, h2))

	assert.NoError(t, dbf.FinishHashTable())
	assert.NoError(t, dbf.Close())

	// Open and validate
	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()
	require.Equal(t, 3, dbf.EntriesCount())
	require.Equal(t, 2, dbf.FileEntriesCount())

	assert.True(t, dbf.Features().HasHashTable())

	count := 0
	fn := func(idx int, hash []byte) error {
		count++
		assert.Equal(t, algo.Size(), len(hash))

		switch idx {
		case 0:
			assert.Equal(t, h1, hash)
		case 2:
			assert.Equal(t, h2, hash)
		default:
			assert.Fail(t, fmt.Sprintf("did not expect the index %d to be read", idx))
		}

		return nil
	}
	assert.NoError(t, dbf.ReadHashTableEntries(fn))
	assert.Equal(t, dbf.FileEntriesCount(), count)
}

//TODO: need to check that you can create and "empty" database with hash table and have no issue opening
