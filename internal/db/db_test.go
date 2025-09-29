package db_test

import (
	"encoding/binary"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type prefixHeader struct {
	Signature [4]byte
	Version   uint16
}

type header struct {
	EntryCount       uint32
	FileEntriesCount uint32
	EntryOffset      uint32
	Features         db.FeatureFlags
	FeatureReserved  [8]uint32
}

func TestCreateDatabase(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)

	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	defer os.Remove(tempFile)
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	f, err := os.Open(tempFile)
	require.NoError(t, err)
	defer f.Close()

	prefix := prefixHeader{}
	err = binary.Read(f, binary.LittleEndian, &prefix)
	require.NoError(t, err)
	expSignature := [4]byte{0x41, 0x4A, 0x46, 0x53} // AJFS
	assert.Equal(t, expSignature, prefix.Signature)
	assert.Equal(t, uint16(1), prefix.Version)

	header := header{}
	err = binary.Read(f, binary.LittleEndian, &header)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), header.EntryCount)
	assert.Greater(t, header.EntryOffset, uint32(0))
	assert.Equal(t, db.FeatureFlags(0), header.Features)
	assert.Equal(t, [8]uint32{}, header.FeatureReserved)
}

func TestCreateDatabaseWhenExistingFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "unit-testing")
	require.NoError(t, err)
	_ = f.Close()
	defer os.Remove(f.Name())

	_, err = db.CreateDatabase(f.Name(), "/test", db.FeatureJustEntries)
	var expErr *fs.PathError
	require.ErrorAs(t, err, &expErr)
}

func TestOpenDatabaseForNonExistentFile(t *testing.T) {
	path := "./does-not-exist"
	require.NoFileExists(t, path)

	_, err := db.OpenDatabase(path)
	var expErr *fs.PathError
	require.ErrorAs(t, err, &expErr)
}

func TestOpenDatabaseWhenInvalidFile(t *testing.T) {
	// Create an invalid file
	f, err := os.CreateTemp("", "unit-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	magic := [6]byte{0x12, 0x4A, 0x46, 0x53, 0x41, 0xAB}
	require.NoError(t, binary.Write(f, binary.LittleEndian, magic))
	_ = f.Close()

	_, err = db.OpenDatabase(f.Name())
	assert.ErrorContains(t, err, "not a valid ajfs file (invalid signature")

	// Create a valid signature but wrong version
	f, err = os.CreateTemp("", "unit-test")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	prefix := prefixHeader{
		Signature: [4]byte{0x41, 0x4A, 0x46, 0x53},
		Version:   42,
	}
	require.NoError(t, binary.Write(f, binary.LittleEndian, &prefix))
	_ = f.Close()

	_, err = db.OpenDatabase(f.Name())
	assert.ErrorContains(t, err, "not a supported ajfs file (invalid version")
}

func TestCreateDatabaseAbsRoot(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)

	dbf, err := db.CreateDatabase(tempFile, "../", db.FeatureJustEntries)
	defer os.Remove(tempFile)
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	absPath, err := filepath.Abs("../")
	require.NoError(t, err)
	assert.Equal(t, absPath, dbf.RootPath())
}

func TestOpenDatabase(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)

	// Create a valid "empty" database
	expRoot := "/test"
	dbf, err := db.CreateDatabase(tempFile, expRoot, db.FeatureJustEntries)
	defer os.Remove(tempFile)
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	// Open and validate
	f, err := db.OpenDatabase(tempFile)
	assert.NoError(t, err)
	defer f.Close()

	assert.Equal(t, tempFile, f.Path())
	assert.Equal(t, 1, f.Version())
	assert.Equal(t, db.FeatureFlags(0), f.Features())
	assert.Equal(t, expRoot, f.RootPath())

	meta := f.Meta()
	assert.Equal(t, runtime.GOOS, meta.OS)
	assert.Equal(t, runtime.GOARCH, meta.Arch)
	assert.True(t, time.Now().After(meta.CreatedAt))
}

func TestWritePathInfo(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database and write 2 path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	require.NoError(t, err)

	p1 := path.Info{
		Id:      path.IdFromPath("a.txt"),
		Path:    "a.txt",
		Size:    uint64(42),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	assert.NoError(t, dbf.WriteEntry(&p1))

	p2 := path.Info{
		Id:      path.IdFromPath("some/dir"),
		Path:    "some/dir",
		Size:    uint64(142),
		Mode:    0644 | fs.ModeDir,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	assert.NoError(t, dbf.WriteEntry(&p2))

	assert.NoError(t, dbf.FinishEntries())
	assert.NoError(t, dbf.Close())

	// Open and validate
	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	assert.Equal(t, 2, dbf.EntriesCount())
	assert.Equal(t, 1, dbf.FileEntriesCount())

	c2, err := dbf.ReadEntryAtIndex(1)
	require.NoError(t, err)
	assert.True(t, p2.Equals(&c2))

	c1, err := dbf.ReadEntryAtIndex(0)
	require.NoError(t, err)
	assert.True(t, p1.Equals(&c1))
}

func TestReadAll(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database and write N path info objects
	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	require.NoError(t, err)

	expCount := 10
	expTime := time.Now().Add(-10 * time.Minute)

	for i := range expCount {
		filePath := fmt.Sprintf("/some/path/%d.txt", i)
		p := path.Info{
			Id:      path.IdFromPath(filePath),
			Path:    filePath,
			Size:    uint64(i),
			Mode:    0740,
			ModTime: expTime,
		}
		require.NoError(t, dbf.WriteEntry(&p))
	}

	require.NoError(t, dbf.FinishEntries())
	require.NoError(t, dbf.Close())

	// Open, read and validate

	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()
	assert.Equal(t, expCount, dbf.EntriesCount())

	rcvCount := 0
	fn := func(idx int, pi path.Info) error {
		rcvCount += 1
		filePath := fmt.Sprintf("/some/path/%d.txt", idx)
		assert.Equal(t, path.IdFromPath(filePath), pi.Id)
		assert.Equal(t, filePath, pi.Path)
		assert.Equal(t, uint64(idx), pi.Size)
		assert.Equal(t, fs.FileMode(0740), pi.Mode)
		assert.True(t, expTime.Equal(pi.ModTime))
		return nil
	}

	assert.NoError(t, dbf.ReadAllEntries(fn))
	assert.Equal(t, expCount, rcvCount)

	// Search for an entry and then stop
	rcvCount = 0
	fnSearch := func(idx int, pi path.Info) error {
		rcvCount += 1
		if idx == 5 {
			return db.SkipAll
		}
		return nil
	}

	assert.NoError(t, dbf.ReadAllEntries(fnSearch))
	assert.Equal(t, 6, rcvCount)
}

func TestReadWritePanicConditions(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	// Create new database
	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	require.NoError(t, err)

	// Not allowed to read
	assert.Panics(t, func() { _, _ = dbf.ReadEntryAtIndex(0) })
	assert.Panics(t, func() { _ = dbf.ReadAllEntries(func(idx int, pi path.Info) error { return nil }) })

	// Write 1 entry
	p := path.Info{
		Id:      path.IdFromPath("some/dir/b.txt"),
		Path:    "some/dir/b.txt",
		Size:    uint64(142),
		Mode:    0644,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	assert.NoError(t, dbf.WriteEntry(&p))

	// Not allowed to Close before you called FinishEntries
	assert.Panics(t, func() { dbf.Close() })

	// Finish creation
	assert.NoError(t, dbf.FinishEntries())
	assert.NoError(t, dbf.Close())

	// Open tests
	dbf, err = db.OpenDatabase(tempFile)
	require.NoError(t, err)
	defer dbf.Close()

	// Not allowed to write
	assert.Panics(t, func() { _ = dbf.WriteEntry(&p) })
	assert.Panics(t, func() { _ = dbf.Flush() })
	assert.Panics(t, func() { _ = dbf.FinishEntries() })

	// Not allowed to read out of index bounds
	assert.Panics(t, func() { _, _ = dbf.ReadEntryAtIndex(1) })
}

//TODO: Need to check if the vardata stuff actually respects endianess
