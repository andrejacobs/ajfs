package db_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteInitialHashTable(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database and write 2 path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test/")
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
		Id:      path.IdFromPath("some/dir/b.txt"),
		Path:    "some/dir/b.txt",
		Size:    uint64(142),
		Mode:    0644,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p2))
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
	require.Equal(t, 2, dbf.EntriesCount())

	assert.True(t, dbf.Features().HasHashTable())

	count := 0
	fn := func(idx int, hash []byte) error {
		count++
		assert.Equal(t, algo.Size(), len(hash))
		assert.Equal(t, algo.ZeroValue(), hash)
		return nil
	}
	assert.NoError(t, dbf.ReadHashTableEntries(fn))
	assert.Equal(t, dbf.EntriesCount(), count)
}

//TODO: need to check that you can create and "empty" database with hash table and have no issue opening
