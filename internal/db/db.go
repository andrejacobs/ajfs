// Package db is responsible for creating and managing the underlying data storage required by ajfs as a single file.
package db

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/fs"
	"os"
	"runtime"
	"time"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajio"
	"github.com/andrejacobs/go-aj/ajmath"
)

// The underlying file format for the ajfs database:
//   Uses little endian
// prefix header
// header
// root
// meta
// entries
// entry offset table
// [optional] hash table
// [optional] tree
// [optional] future features (without breaking existing databases)

// DatabaseFile is the underlying data storage used by ajfs as a single file.
//
// NOTE: The order of operations during the creation process is very important:
// - CreateDatabase
// - n * Write
// - Finish
// - Close
type DatabaseFile struct {
	file *os.File
	path string

	prefixHeader prefixHeader
	header       header
	root         rootEntry
	meta         MetaEntry

	entryOffsets []uint32

	// only for reading
	reader ajio.MultiByteReaderSeeker

	// only for creation
	bufWriter *bufio.Writer
	writer    ajio.TrackedOffsetWriter
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

	dbf.bufWriter = bufio.NewWriter(dbf.file)
	dbf.writer = ajio.NewTrackedOffsetWriter(dbf.bufWriter, 0)

	// Write prefix
	dbf.prefixHeader.init()
	if err := dbf.prefixHeader.write(dbf.writer); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs prefix header. path: %q. %w", path, err)
	}

	// Write initial empty header (this should be updated before finishing the file)
	if err := dbf.header.write(dbf.writer); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs header. path: %q. %w", path, err)
	}

	// Root entry
	dbf.root.path = root
	if err := dbf.root.write(dbf.writer); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs root entry. path: %s. %w", path, err)
	}

	// Meta entry
	dbf.meta.init()
	if err := dbf.meta.write(dbf.writer); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs meta entry. path: %q. %w", path, err)
	}

	// Determine the start of the path object entries
	dbf.header.EntriesOffset, err = ajmath.Uint64ToUint32(dbf.currentWriteOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to set the ajfs EntriesOffset. %w", err)
	}

	dbf.entryOffsets = make([]uint32, 0, 4096)

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

	// Read the entry offset table
	if err := dbf.readEntryOffsets(); err != nil {
		return nil, fmt.Errorf("failed to read the ajfs entry offset table. path: %q. %w", path, err)
	}

	return dbf, nil
}

// Sync pending writes and close the file
func (dbf *DatabaseFile) Close() error {
	if dbf.file == nil {
		return nil
	}

	if dbf.bufWriter != nil {
		if err := dbf.Flush(); err != nil {
			return err
		}

		if err := dbf.finishCreation(); err != nil {
			return err
		}

		if err := dbf.file.Sync(); err != nil {
			return err
		}
	}

	if err := dbf.file.Close(); err != nil {
		return err
	}

	dbf.file = nil
	dbf.reader = nil
	dbf.bufWriter = nil
	dbf.writer = nil
	dbf.entryOffsets = nil

	return nil
}

// Ensure unwritten data is written to the file on disk.
func (dbf *DatabaseFile) Flush() error {
	if dbf.bufWriter == nil {
		panic("database was not opened for writing")
	}

	return dbf.bufWriter.Flush()
}

// File format version.
func (dbf *DatabaseFile) Version() int {
	return int(dbf.prefixHeader.Version)
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

// The number of path info entries.
func (dbf *DatabaseFile) EntriesCount() int {
	return int(dbf.header.EntriesCount)
}

// Write the path info to the database.
func (dbf *DatabaseFile) WriteEntry(pi *path.Info) error {
	if dbf.bufWriter == nil {
		panic("database was not opened for writing")
	}

	offset, err := ajmath.Uint64ToUint32(dbf.currentWriteOffset())
	if err != nil {
		return err
	}
	dbf.entryOffsets = append(dbf.entryOffsets, offset)

	entry := pathEntryFromPathInfo(pi)
	if err := entry.write(dbf.writer); err != nil {
		return err
	}

	dbf.header.EntriesCount, err = ajmath.Add32(dbf.header.EntriesCount, 1)
	return err
}

// Read the path info object with the specified index.
func (dbf *DatabaseFile) ReadEntryAtIndex(idx int) (path.Info, error) {
	if dbf.reader == nil {
		panic("database was not opened for reading")
	}

	if idx >= int(dbf.header.EntriesCount) {
		panic(fmt.Sprintf("invalid index %d, EntriesCount = %d", idx, dbf.header.EntriesCount))
	}

	offset := dbf.entryOffsets[idx]
	_, err := dbf.reader.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return path.Info{}, fmt.Errorf("failed to read entry at index %d (offset %d). %w", idx, offset, err)
	}

	entry := pathEntry{}
	if err := entry.read(dbf.reader); err != nil {
		return path.Info{}, fmt.Errorf("failed to read entry at index %d (offset %d). %w", idx, offset, err)
	}

	return pathInfoFromPathEntry(&entry), nil
}

// ReadAllEntriesFn will be called by ReadAllEntries for each entry that was read from the database.
// idx Is the index of the entry.
// pi Is the path info object.
// Return [SkipAll] to stop reading all the entries.
type ReadAllEntriesFn func(idx int, pi path.Info) error

// Read all the path info objects from the database and call the callback function.
// If the callback function returns [SkipAll] then the reading process will be stopped and nil will be returned as the error.
func (dbf *DatabaseFile) ReadAllEntries(fn ReadAllEntriesFn) error {
	if dbf.reader == nil {
		panic("database was not opened for reading")
	}

	_, err := dbf.reader.Seek(int64(dbf.header.EntriesOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read all entries. %w", err)
	}

	for idx := range dbf.header.EntriesCount {
		entry := pathEntry{}
		if err := entry.read(dbf.reader); err != nil {
			offset, _ := dbf.currentReadOffset()
			return fmt.Errorf("failed to read entry at index %d (offset %d). %w", idx, offset, err)
		}

		if err := fn(int(idx), pathInfoFromPathEntry(&entry)); err != nil {
			if err == SkipAll {
				return nil
			}
			return err
		}
	}

	return nil
}

// Write the entries offset table after all path info objects have been written.
func (dbf *DatabaseFile) FinishEntries() error {
	if dbf.header.EntriesCount == 0 {
		return nil
	}

	if err := dbf.Flush(); err != nil {
		return fmt.Errorf("failed to finish writing the entries (flush). %w", err)
	}

	var err error
	dbf.header.EntriesOffsetTableOffset, err = ajmath.Uint64ToUint32(dbf.currentWriteOffset())
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (offset). %w", err)
	}

	if err := dbf.writeEntryOffsets(); err != nil {
		return fmt.Errorf("failed to finish writing the entries (offset table). %w", err)
	}

	return nil
}

//-----------------------------------------------------------------------------

// Update the header
func (dbf *DatabaseFile) finishCreation() error {
	if dbf.bufWriter == nil {
		panic("database was not opened for writing")
	}

	if (dbf.header.EntriesCount > 0) && (dbf.header.EntriesOffsetTableOffset == 0) {
		panic("FinishEntries was never called")
	}

	// Update the header
	_, err := dbf.file.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to finish creating the ajfs database. failed to seek to header offset. %w", err)
	}

	if err := dbf.header.write(dbf.file); err != nil {
		return fmt.Errorf("failed to update the ajfs header. %w", err)
	}

	dbf.bufWriter = nil
	dbf.writer = nil

	return nil
}

// Read the entry offset table
func (dbf *DatabaseFile) readEntryOffsets() error {
	if dbf.header.EntriesCount == 0 {
		return nil
	}

	_, err := dbf.reader.Seek(int64(dbf.header.EntriesOffsetTableOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read the entry offset table. %w", err)
	}

	// Check 1st sentinel
	var s [4]byte
	_, err = dbf.reader.Read(s[:])
	if err != nil {
		return fmt.Errorf("failed to read the entry offset table (1st sentinel). %w", err)
	}
	if s != sentinel {
		return fmt.Errorf("failed to read the entry offset table (1st sentinel %q does not match %q)", s, sentinel)
	}

	dbf.entryOffsets = make([]uint32, dbf.header.EntriesCount)

	var data [4]byte // uint32
	for i := range dbf.header.EntriesCount {
		_, err := dbf.reader.Read(data[:])
		if err != nil {
			return fmt.Errorf("failed to read the entry offset table (near index %d). %w", i, err)
		}

		dbf.entryOffsets[i] = binary.LittleEndian.Uint32(data[:])
	}

	_, err = dbf.reader.Read(s[:])
	if err != nil {
		return fmt.Errorf("failed to read the entry offset table (2nd sentinel). %w", err)
	}
	if s != sentinel {
		return fmt.Errorf("failed to read the entry offset table (2nd sentinel %q does not match %q)", s, sentinel)
	}

	return nil
}

// Write the entry offset table
func (dbf *DatabaseFile) writeEntryOffsets() error {
	if dbf.header.EntriesCount == 0 {
		return nil
	}

	// 1st sentinel
	_, err := dbf.writer.Write(sentinel[:])
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (1st sentinel). %w", err)
	}

	data := make([]byte, 4) // uint32

	for idx, offset := range dbf.entryOffsets {
		binary.LittleEndian.PutUint32(data, offset)
		_, err := dbf.writer.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write the entries offset table (index = %d). %w", idx, err)
		}
	}

	// 2nd sentinel
	_, err = dbf.writer.Write(sentinel[:])
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (2nd sentinel). %w", err)
	}

	if err := dbf.Flush(); err != nil {
		return fmt.Errorf("failed to finish writing the entries (flush). %w", err)
	}

	return nil
}

// Current read position in the file
func (dbf *DatabaseFile) currentReadOffset() (int64, error) {
	return dbf.reader.Seek(0, io.SeekCurrent)
}

// Current write position in the file
func (dbf *DatabaseFile) currentWriteOffset() uint64 {
	return dbf.writer.Offset()
}

//-----------------------------------------------------------------------------
// Prefix Header

// First part of the file to identify the type and version of the format
type prefixHeader struct {
	Signature [4]byte // AJFS
	Version   uint16  // Version of the file format
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
	EntriesCount             uint32 // The number of path objects. Based on inode max limit of 2^32
	EntriesOffset            uint32 // The offset in bytes at which the path objects start. Based on limit of database file being max 4GB
	EntriesOffsetTableOffset uint32 // The offset to the entries offset table

	Features FeatureFlags // Feature flags

	FeatureReserved [8]uint32 // 8x feature offsets reserved for future use without breaking backwards compatibility
}

func (s *header) read(r io.ReadSeeker) error {
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *header) write(w io.Writer) error {
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
// Path info

// Path entry
type pathEntry struct {
	header pathEntryHeader // fixed size struct to make serialization easier

	// The following fields will be written as the size of the data varint followed by the encoded form of the data
	modTime time.Time // Last modification time.
	path    string    // The file system path.
}

type pathEntryHeader struct {
	Id   path.Id // The unique identifier
	Size uint64  // Size in bytes, if it is a file
	Type fs.FileMode
	Mode fs.FileMode
}

func (s *pathEntry) read(r ajio.MultiByteReader) error {
	if err := binary.Read(r, binary.LittleEndian, &s.header); err != nil {
		return fmt.Errorf("failed to read path entry header. %w", err)
	}

	// ModTime
	data, _, err := varData.Read(r, nil)
	if err != nil {
		return fmt.Errorf("failed to read path entry modification time. %w", err)
	}
	if err := s.modTime.UnmarshalBinary(data); err != nil {
		return fmt.Errorf("failed to read path entry modification time (decoding failed). %w", err)
	}

	// Path
	data, _, err = varData.Read(r, nil)
	if err != nil {
		return fmt.Errorf("failed to read path entry's path string. %w", err)
	}

	s.path = string(data)
	return nil
}

func (s *pathEntry) write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, s.header); err != nil {
		return fmt.Errorf("failed to write path entry header. path: %q. %w", s.path, err)
	}

	// ModTime
	data, err := s.modTime.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to write path entry modification time (encoding failed). path: %q. %w", s.path, err)
	}
	if _, err := varData.Write(w, data); err != nil {
		return fmt.Errorf("failed to write path entry modification time. path: %q. %w", s.path, err)
	}

	// Path
	if _, err := varData.WriteString(w, s.path); err != nil {
		return fmt.Errorf("failed to write path entry's path string. path: %q. %w", s.path, err)
	}

	return nil
}

//-----------------------------------------------------------------------------
// Feature flags

type FeatureFlags uint16

const (
	featureHashTable = 1 << iota // Contains the calculated file hash signatures for the path objects.
	featureTree                  // Contains the cached file tree.
)

func (f FeatureFlags) HasHashTable() bool {
	return (f & featureHashTable) != 0
}

func (f FeatureFlags) HasTree() bool {
	return (f & featureTree) != 0
}

//-----------------------------------------------------------------------------
// Helpers

// Convert from path.PathInfo to pathEntry
func pathEntryFromPathInfo(i *path.Info) pathEntry {
	result := pathEntry{
		header: pathEntryHeader{
			Id:   i.Id,
			Size: i.Size,
			Mode: i.Mode,
		},
		modTime: i.ModTime,
		path:    i.Path,
	}
	return result
}

// Convert from pathEntry to path.PathInfo
func pathInfoFromPathEntry(e *pathEntry) path.Info {
	result := path.Info{
		Id:      e.header.Id,
		Size:    e.header.Size,
		Mode:    e.header.Mode,
		ModTime: e.modTime,
		Path:    e.path,
	}
	return result
}

//-----------------------------------------------------------------------------
// Constants and Misc

var (
	SkipAll = fs.SkipAll
)

var (
	signature = [4]byte{0x41, 0x4A, 0x46, 0x53} // AJFS
	sentinel  = [4]byte{0x41, 0x4A, 0x43, 0x43} // AJCC (as in interupt 3 0xCC :-)
	varData   = ajio.NewVariableData()
)

const (
	currentVersion = uint16(1)
)
