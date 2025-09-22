package db_test

import (
	"encoding/binary"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type prefixHeader struct {
	Signature [4]byte
	Version   uint8
}

type header struct {
	EntryCount      uint64
	EntryOffset     uint64
	Features        db.FeatureFlags
	FeatureReserved [8]uint64
}

func TestCreateDatabase(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile) // delete if it already exists

	dbf, err := db.CreateDatabase(tempFile, "/test/")
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
	assert.Equal(t, uint8(1), prefix.Version)

	header := header{}
	err = binary.Read(f, binary.LittleEndian, &header)
	require.NoError(t, err)
	assert.Equal(t, uint64(0), header.EntryCount)
	assert.Equal(t, uint64(0), header.EntryOffset)
	assert.Equal(t, db.FeatureFlags(0), header.Features)
	assert.Equal(t, [8]uint64{}, header.FeatureReserved)
}

func TestCreateDatabaseWhenExistingFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "unit-testing")
	require.NoError(t, err)
	_ = f.Close()
	defer os.Remove(f.Name())

	_, err = db.CreateDatabase(f.Name(), "/test/")
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
	magic := [5]byte{0x12, 0x4A, 0x46, 0x53, 0x41}
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

func TestOpenDatabase(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile) // delete if it already exists

	// Create a valid "empty" database
	expRoot := "/test/"
	dbf, err := db.CreateDatabase(tempFile, expRoot)
	defer os.Remove(tempFile)
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	// Open and validate
	f, err := db.OpenDatabase(tempFile)
	assert.NoError(t, err)
	defer f.Close()

	assert.Equal(t, tempFile, f.Path())
	assert.Equal(t, uint8(1), f.Version())
	assert.Equal(t, db.FeatureFlags(0), f.Features())
	assert.Equal(t, expRoot, f.RootPath())

	meta := f.Meta()
	assert.Equal(t, runtime.GOOS, meta.OS)
	assert.Equal(t, runtime.GOARCH, meta.Arch)
	assert.True(t, time.Now().After(meta.CreatedAt))
}
