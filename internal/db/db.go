// Package db is responsible for creating and managing the underlying data storage required by ajfs as a single file.
package db

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/andrejacobs/go-aj/ajio"
)

// The underlying file format for the ajfs database:
//   Uses little endian
// prefix header
// header
// root
// meta
// entries
// entry offset table
// [optional] checksum table
// [optional] tree
// [optional] future features (without breaking existing databases)

// DatabaseFile is the underlying data storage used by ajfs as a single file.
type DatabaseFile struct {
	file *os.File
	path string

	prefixHeader prefixHeader
	header       header
	root         rootEntry
	meta         MetaEntry

	reader ajio.MultiByteReaderSeeker
}

// Create a new file
// If the file already exists then an error will be returned.
// path is the file path at which the database file will be created.
// root is the file path that the database will represents and that will be used to scan the file hierarchy.
func CreateDatabase(path string, root string) (*DatabaseFile, error) {
	dbf := &DatabaseFile{
		path: path,
	}

	var err error
	dbf.file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create the ajfs database file. path: %q. %w", path, err)
	}

	dbf.reader = ajio.NewMultiByteReaderSeeker(dbf.file)

	// Write prefix
	dbf.prefixHeader.init()
	if err := dbf.prefixHeader.write(dbf.file); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs prefix header. path: %q. %w", path, err)
	}

	// Write initial empty header (this should be updated before finishing the file)
	if err := dbf.header.write(dbf.file); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs header. path: %q. %w", path, err)
	}

	// Root entry
	dbf.root.path = root
	if err := dbf.root.write(dbf.file); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs root entry. path: %s. %w", path, err)
	}

	// Meta entry
	dbf.meta.init()
	if err := dbf.meta.write(dbf.file); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs meta entry. path: %q. %w", path, err)
	}

	// Determine the start of the path object entries
	offset, err := currentOffset(dbf.file)
	if err != nil {
		return nil, fmt.Errorf("failed to get the current file offset. path: %q. %w", path, err)
	}
	dbf.header.EntriesOffset = uint64(offset)

	return dbf, nil
}

// Open an existing database file and check the signature is valid and the version is supported.
func OpenDatabase(path string) (*DatabaseFile, error) {
	dbf := &DatabaseFile{
		path: path,
	}

	var err error
	dbf.file, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the ajfs database file. path: %q. %w", path, err)
	}

	dbf.reader = ajio.NewMultiByteReaderSeeker(dbf.file)

	// Check the signature and version
	if err := dbf.prefixHeader.read(dbf.reader); err != nil {
		return nil, fmt.Errorf("error reading the ajfs prefix header. path: %q. %w", path, err)
	}
	if dbf.prefixHeader.Signature != signature {
		return nil, fmt.Errorf("not a valid ajfs file (invalid signature %q, expected %q). path: %q", dbf.prefixHeader.Signature, signature, path)
	}
	if dbf.prefixHeader.Version > currentVersion {
		return nil, fmt.Errorf("not a supported ajfs file (invalid version %d, expected <= %d). path: %q", dbf.prefixHeader.Version, currentVersion, path)
	}

	// Read the header
	if err := dbf.header.read(dbf.reader); err != nil {
		return nil, fmt.Errorf("failed to read the ajfs header. path: %q. %w", path, err)
	}

	// Read the root info
	if err := dbf.root.read(dbf.reader); err != nil {
		return nil, fmt.Errorf("failed to read the ajfs root entry. path: %q. %w", path, err)
	}

	// Read the meta info
	if err := dbf.meta.read(dbf.reader); err != nil {
		return nil, fmt.Errorf("failed to read the ajfs meta entry. path: %q. %w", path, err)
	}

	return dbf, nil
}

// Sync pending writes and close the file
func (dbf *DatabaseFile) Close() error {
	if dbf.file == nil {
		return nil
	}

	if err := dbf.file.Sync(); err != nil {
		return err
	}

	if err := dbf.file.Close(); err != nil {
		return err
	}

	dbf.file = nil
	return nil
}

// File format version.
func (dbf *DatabaseFile) Version() uint8 {
	return dbf.prefixHeader.Version
}

// File path to the database.
func (dbf *DatabaseFile) Path() string {
	return dbf.path
}

// Features present in the database.
func (dbf *DatabaseFile) Features() FeatureFlags {
	return dbf.header.Features
}

// The file path that the database represents and that was used to scan the file hierarchy.
func (dbf *DatabaseFile) RootPath() string {
	return dbf.root.path
}

// Meta data about the database.
func (dbf *DatabaseFile) Meta() MetaEntry {
	return dbf.meta
}

//-----------------------------------------------------------------------------
// Prefix Header

// First part of the file to identify the type and version of the format
type prefixHeader struct {
	Signature [4]byte // AJFS
	Version   uint8   // Version of the file format
}

func (s *prefixHeader) init() {
	s.Signature = signature
	s.Version = currentVersion
}

func (s *prefixHeader) read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *prefixHeader) write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, s)
}

//-----------------------------------------------------------------------------
// Header (version 1)

type header struct {
	EntriesCount             uint64 // The number of path objects
	EntriesOffset            uint64 // The offset in bytes at which the path objects start
	EntriesOffsetTableOffset uint64 // The offset to the entries offset table

	Features FeatureFlags // Feature flags

	FeatureReserved [8]uint64 // 8x feature offsets reserved for future use without breaking backwards compatibility
}

func (s *header) read(r io.ReadSeeker) error {
	_, err := r.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return err
	}
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *header) write(w io.WriteSeeker) error {
	var err error
	_, err = w.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return err
	}

	return binary.Write(w, binary.LittleEndian, s)
}

func headerOffset() int64 {
	return int64(binary.Size(prefixHeader{}))
}

//-----------------------------------------------------------------------------
// Root entry

type rootEntry struct {
	// The following fields will be written as the size of the data varint followed by the encoded form of the data
	path string
}

func (s *rootEntry) read(r ajio.MultiByteReader) error {
	path, _, err := varData.ReadString(r)
	if err != nil {
		return fmt.Errorf("failed to read the root path. %w", err)
	}
	s.path = path
	return nil
}

func (s *rootEntry) write(w io.Writer) error {
	_, err := varData.WriteString(w, s.path)
	if err != nil {
		return fmt.Errorf("failed to write the root path. %w", err)
	}
	return nil
}

//-----------------------------------------------------------------------------
// Meta entry

// Meta info about how the database was created
type MetaEntry struct {
	// The following fields will be written as the size of the data varint followed by the encoded form of the data
	OS        string    `json:"os"`        // The operating system (e.g. darwin, linux, windows etc.)
	Arch      string    `json:"arch"`      // The architecture (e.g. arm64 etc.)
	CreatedAt time.Time `json:"createdAt"` // Time of database creation (this is captured instead of relying on the file system time)

	// NOTE: You can see the list of GOOS values at: https://github.com/golang/go/blob/master/src/go/build/syslist.go
}

func (s *MetaEntry) init() {
	s.OS = runtime.GOOS
	s.Arch = runtime.GOARCH
	s.CreatedAt = time.Now()
}

func (s *MetaEntry) read(r ajio.MultiByteReader) error {
	os, _, err := varData.ReadString(r)
	if err != nil {
		return fmt.Errorf("failed to read the operating system. %w", err)
	}
	s.OS = os

	arch, _, err := varData.ReadString(r)
	if err != nil {
		return fmt.Errorf("failed to read the architecture. %w", err)
	}
	s.Arch = arch

	data, _, err := varData.Read(r, nil)
	if err != nil {
		return fmt.Errorf("failed to read creation time. %w", err)
	}
	if err := s.CreatedAt.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("failed to read creation time (decoding failed). %w", err)
	}

	return nil
}

func (s *MetaEntry) write(w io.Writer) error {
	_, err := varData.WriteString(w, s.OS)
	if err != nil {
		return fmt.Errorf("failed to write the operating system. %w", err)
	}

	_, err = varData.WriteString(w, s.Arch)
	if err != nil {
		return fmt.Errorf("failed to write the architecture. %w", err)
	}

	data, err := s.CreatedAt.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to write creation time (encoding failed). %w", err)
	}
	if _, err := varData.Write(w, data); err != nil {
		return fmt.Errorf("failed to write creation time. %w", err)
	}

	return nil
}

//-----------------------------------------------------------------------------
// Feature flags

type FeatureFlags uint16

const (
	featureChecksums = 1 << iota // Contains the calculated file checksums (hashes) for the path objects.
	featureTree                  // Contains the cached file tree.
)

func (f FeatureFlags) HasChecksums() bool {
	return (f & featureChecksums) != 0
}

func (f FeatureFlags) HasTree() bool {
	return (f & featureTree) != 0
}

//-----------------------------------------------------------------------------
// Helpers

// Current position in the file
func currentOffset(w io.WriteSeeker) (int64, error) {
	return w.Seek(0, io.SeekCurrent)
}

//-----------------------------------------------------------------------------
// Constants and Misc

var (
	signature = [4]byte{0x41, 0x4A, 0x46, 0x53} // AJFS
)

const (
	currentVersion = uint8(1)
)

var (
	varData = ajio.NewVariableData()
)
