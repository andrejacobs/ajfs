package db_test

import (
	"encoding/hex"
	"errors"
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
	testCases := []struct {
		algo ajhash.Algo
	}{
		{
			algo: ajhash.AlgoSHA1,
		},
		{
			algo: ajhash.AlgoSHA256,
		},
		{
			algo: ajhash.AlgoSHA512,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.algo.String(), func(t *testing.T) {
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

			algo := tC.algo
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
		})
	}
}

func TestWriteHashTable(t *testing.T) {
	testCases := []struct {
		algo ajhash.Algo
	}{
		{
			algo: ajhash.AlgoSHA1,
		},
		{
			algo: ajhash.AlgoSHA256,
		},
		{
			algo: ajhash.AlgoSHA512,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.algo.String(), func(t *testing.T) {
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

			algo := tC.algo
			assert.NoError(t, dbf.StartHashTable(algo))

			h1 := make([]byte, algo.Size())
			require.NoError(t, random.SecureBytes(h1))
			require.NoError(t, dbf.WriteHashEntry(0, h1))

			h2 := make([]byte, algo.Size())
			require.NoError(t, random.SecureBytes(h2))
			require.NoError(t, dbf.WriteHashEntry(2, h2))

			assert.Panics(t, func() {
				buf := make([]byte, algo.Size()+1)
				dbf.WriteHashEntry(1, buf)
			})

			assert.NoError(t, dbf.FinishHashTable())
			assert.NoError(t, dbf.Close())

			// Open and validate
			dbf, err = db.OpenDatabase(tempFile)
			require.NoError(t, err)
			defer dbf.Close()
			require.Equal(t, 3, dbf.EntriesCount())
			require.Equal(t, 2, dbf.FileEntriesCount())

			assert.True(t, dbf.Features().HasHashTable())

			ht, err := dbf.ReadHashTable()
			require.NoError(t, err)
			assert.Len(t, ht, dbf.FileEntriesCount())

			hash, ok := ht[0]
			assert.True(t, ok)
			assert.Equal(t, h1, hash)

			hash, ok = ht[2]
			assert.True(t, ok)
			assert.Equal(t, h2, hash)

			_, ok = ht[1]
			assert.False(t, ok)
		})
	}
}

func TestEntriesNeedHashing(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	algo := ajhash.AlgoSHA1

	// Create new database and write path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureHashTable)
	require.NoError(t, err)
	defer dbf.Close()

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

	// Start initial hash table with empty hashes
	assert.NoError(t, dbf.StartHashTable(algo))
	assert.NoError(t, dbf.FinishHashTable())

	//----------

	// Cause an error
	expErr := errors.New("unit-testing err")
	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		return expErr
	})
	require.ErrorIs(t, err, expErr)

	// Skip
	rcvIdx := make([]int, 0, 4)
	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		rcvIdx = append(rcvIdx, idx)
		return db.SkipAll
	})
	require.NoError(t, err)
	assert.Len(t, rcvIdx, 1)

	// Check which ones still need to be calculated
	rcvIdx = make([]int, 0, 4)
	rcvPi := make([]path.Info, 0, 4)

	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		rcvIdx = append(rcvIdx, idx)
		rcvPi = append(rcvPi, pi)
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, []int{0, 2}, rcvIdx)
	assert.True(t, p1.Equals(&rcvPi[0]))
	assert.True(t, p3.Equals(&rcvPi[1]))

	// Write p1's hash
	h1 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h1))
	dbf.WriteHashEntry(0, h1)

	// Check again
	rcvIdx = make([]int, 0, 4)
	rcvPi = make([]path.Info, 0, 4)

	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		rcvIdx = append(rcvIdx, idx)
		rcvPi = append(rcvPi, pi)
		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, []int{2}, rcvIdx)
	assert.True(t, p3.Equals(&rcvPi[0]))

	// Write p3's hash
	h3 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h3))
	dbf.WriteHashEntry(2, h3)

	// Check again
	rcvIdx = make([]int, 0, 4)
	rcvPi = make([]path.Info, 0, 4)

	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		rcvIdx = append(rcvIdx, idx)
		rcvPi = append(rcvPi, pi)
		return nil
	})
	require.NoError(t, err)

	assert.Len(t, rcvIdx, 0)
	assert.Len(t, rcvPi, 0)
}

func TestFindDuplicatesPanics(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Empty database
	dbf, err := db.CreateDatabase(tempFile, "/test/", db.FeatureJustEntries)
	require.NoError(t, err)
	assert.NoError(t, dbf.Close())

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	assert.Panics(t, func() { _, _ = dbf.FindDuplicateHashes() })
	assert.Panics(t, func() {
		_ = dbf.FindDuplicates(func(group int, idx int, pi path.Info, hash string) error { return nil })
	})
}

func TestFindDuplicates(t *testing.T) {
	algo := ajhash.AlgoSHA1

	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

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

	// Write a duplicate
	require.NoError(t, dbf.WriteEntry(&p1))

	require.NoError(t, dbf.FinishEntries())

	// Start initial hash table with empty hashes
	assert.NoError(t, dbf.StartHashTable(algo))
	assert.NoError(t, dbf.FinishHashTable())

	// Hashes
	h1 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h1))
	dbf.WriteHashEntry(0, h1)

	h3 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h3))
	dbf.WriteHashEntry(2, h3)

	dbf.WriteHashEntry(3, h1)

	assert.NoError(t, dbf.Close())

	// Find duplicates

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	dupes, err := dbf.FindDuplicateHashes()
	require.NoError(t, err)

	assert.Len(t, dupes, 1)
	indices, ok := dupes[hex.EncodeToString(h1)]
	assert.True(t, ok)
	expIndices := []uint32{0, 3}
	assert.ElementsMatch(t, expIndices, indices)

	err = dbf.FindDuplicates(func(group int, idx int, pi path.Info, hash string) error {
		assert.Equal(t, 0, group)
		switch idx {
		case 0:
			assert.True(t, p1.Equals(&pi))
		case 3:
			assert.True(t, p1.Equals(&pi))
		default:
			assert.Fail(t, "not a duplicate!")
		}
		return nil
	})
	require.NoError(t, err)
}

func TestReadAllEntriesWithHashes(t *testing.T) {
	algo := ajhash.AlgoSHA1

	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

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
	assert.NoError(t, dbf.StartHashTable(algo))
	assert.NoError(t, dbf.FinishHashTable())

	// Hashes
	h1 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h1))
	dbf.WriteHashEntry(0, h1)

	h3 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h3))
	dbf.WriteHashEntry(2, h3)

	assert.NoError(t, dbf.Close())

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	result := make([][]byte, 0, 2)
	err = dbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
		result = append(result, hash)
		return nil
	})
	require.NoError(t, err)

	expected := [][]byte{h1, h3}
	assert.Equal(t, expected, result)
}

func TestBuildIdToHashMap(t *testing.T) {
	algo := ajhash.AlgoSHA1

	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

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

	p3 := path.Info{
		Id:      path.IdFromPath("c.txt"),
		Path:    "c.txt",
		Size:    uint64(442),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p3))

	require.NoError(t, dbf.FinishEntries())
	assert.NoError(t, dbf.StartHashTable(algo))
	assert.NoError(t, dbf.FinishHashTable())

	// Hashes
	h1 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h1))
	require.NoError(t, dbf.WriteHashEntry(0, h1))

	h3 := algo.Buffer()
	require.NoError(t, random.SecureBytes(h3))
	require.NoError(t, dbf.WriteHashEntry(1, h3))

	assert.NoError(t, dbf.Close())

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	hm, err := dbf.BuildIdToHashMap()
	require.NoError(t, err)
	assert.Len(t, hm, 2)

	v, ok := hm[p1.Id]
	assert.True(t, ok)
	assert.Equal(t, h1, v)

	v, ok = hm[p3.Id]
	assert.True(t, ok)
	assert.Equal(t, h3, v)
}

func TestBuildHashStrToIndexMap(t *testing.T) {
	assert.Fail(t, "TODO!")
}
