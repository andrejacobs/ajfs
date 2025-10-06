// Package db is responsible for creating and managing the underlying data storage required by ajfs as a single file.
package db

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajio/trackedoffset"
	"github.com/andrejacobs/go-aj/ajio/vardata"
	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// The underlying file format for the ajfs database:
//   Uses little endian
//   [c] means will be included in the file checksum
//   Feature will be checksummed individually
//
// prefix header
// header
// root [c]
// meta [c]
// entries [c]
// entry offset table [c]
// [optional] hash table
// [optional] tree
// [optional] future features (without breaking existing databases)

// DatabaseFile is the underlying data storage used by ajfs as a single file.
//
// NOTE: The order of operations during the creation process is very important:
// - CreateDatabase
// - n * Write
// - [features]
// - Finish
// - Close
type DatabaseFile struct {
	file *trackedoffset.File
	path string

	prefixHeader prefixHeader
	header       header
	root         rootEntry
	meta         MetaEntry

	entryOffsets []uint32 // offset to where each path info entry is stored

	// only for creation
	creating       bool
	createFeatures FeatureFlags
	fileIndices    []uint32 // indices of path info entries that are files

	checksumHasher hash.Hash32
	checksumWriter io.Writer

	createHashTable createHashTable
	resuming        bool
}

// Create a new file
// If the file already exists then an error will be returned.
// path is the file path at which the database file will be created.
// root is the file path that the database will represents and that will be used to scan the file hierarchy.
// features indicate the expected features that will be present in the database.
func CreateDatabase(path string, root string, features FeatureFlags) (*DatabaseFile, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to get the absolute root path from %q. %w", root, err)
	}

	dbf := &DatabaseFile{
		path:           path,
		creating:       true,
		createFeatures: features,
	}

	dbf.file, err = trackedoffset.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to create the ajfs database file. path: %q. %w", path, err)
	}

	dbf.checksumHasher = crc32.NewIEEE()
	dbf.checksumWriter = io.MultiWriter(dbf.file, dbf.checksumHasher)

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
	dbf.root.path = absRoot
	if err := dbf.root.write(dbf.checksumWriter); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs root entry. path: %s. %w", path, err)
	}

	// Meta entry
	dbf.meta.init()
	if err := dbf.meta.write(dbf.checksumWriter); err != nil {
		return nil, fmt.Errorf("failed to write the ajfs meta entry. path: %q. %w", path, err)
	}

	if err := dbf.file.Flush(); err != nil {
		return nil, fmt.Errorf("failed to create the ajfs database. path: %q. %w", path, err)
	}

	// Determine the start of the path object entries
	dbf.header.EntriesOffset, err = safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return nil, fmt.Errorf("failed to set the ajfs EntriesOffset. %w", err)
	}

	dbf.entryOffsets = make([]uint32, 0, 4096)

	if dbf.createFeatures.HasHashTable() {
		dbf.fileIndices = make([]uint32, 0, 4096)
	}

	return dbf, nil
}

// Open an existing database file (as read-only) and check the signature is valid and the version is supported.
func OpenDatabase(path string) (*DatabaseFile, error) {
	dbf := &DatabaseFile{
		path: path,
	}

	var err error
	dbf.file, err = trackedoffset.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the ajfs database file. path: %q. %w", path, err)
	}

	if err = dbf.readHeadersAndVerify(); err != nil {
		return nil, err
	}

	return dbf, nil
}

// Open an existing database file (read-write) to resume processing of extra features.
func ResumeDatabase(path string) (*DatabaseFile, error) {
	dbf := &DatabaseFile{
		path:     path,
		resuming: true,
	}

	var err error
	dbf.file, err = trackedoffset.OpenFile(path, os.O_RDWR|os.O_EXCL, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open the ajfs database file. path: %q. %w", path, err)
	}

	if err = dbf.readHeadersAndVerify(); err != nil {
		return nil, err
	}

	dbf.creating = true

	if dbf.Features().HasHashTable() {
		if err = dbf.resumeHashTable(); err != nil {
			return nil, err
		}
	}

	return dbf, nil
}

func (dbf *DatabaseFile) readHeadersAndVerify() error {
	// Check the signature and version
	if err := dbf.prefixHeader.read(dbf.file); err != nil {
		return fmt.Errorf("error reading the ajfs prefix header. path: %q. %w", dbf.path, err)
	}
	if dbf.prefixHeader.Signature != signature {
		return fmt.Errorf("not a valid ajfs file (invalid signature %q, expected %q). path: %q", dbf.prefixHeader.Signature, signature, dbf.path)
	}
	if dbf.prefixHeader.Version > currentVersion {
		return fmt.Errorf("not a supported ajfs file (invalid version %d, expected <= %d). path: %q", dbf.prefixHeader.Version, currentVersion, dbf.path)
	}

	// Read the header
	if err := dbf.header.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs header. path: %q. %w", dbf.path, err)
	}

	// Read the root info
	if err := dbf.root.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs root entry. path: %q. %w", dbf.path, err)
	}

	// Read the meta info
	if err := dbf.meta.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs meta entry. path: %q. %w", dbf.path, err)
	}

	// Read the entry offset table
	if err := dbf.readEntryOffsets(); err != nil {
		return fmt.Errorf("failed to read the ajfs entry offset table. path: %q. %w", dbf.path, err)
	}

	return nil
}

// Sync pending writes and close the file
func (dbf *DatabaseFile) Close() error {
	if dbf.file == nil {
		return nil
	}

	if dbf.creating {
		if err := dbf.Flush(); err != nil {
			return err
		}

		if !dbf.resuming {
			if err := dbf.finishCreation(); err != nil {
				return err
			}
		}

		if err := dbf.file.Sync(); err != nil {
			return err
		}
	}

	if err := dbf.file.Close(); err != nil {
		return err
	}

	dbf.file = nil
	dbf.entryOffsets = nil
	dbf.fileIndices = nil

	return nil
}

// Called when the app has to shutdown before the database could be created.
// This will remove the database file.
func (dbf *DatabaseFile) Interrupted() error {
	if dbf.file == nil {
		return nil
	}

	if err := dbf.file.Close(); err != nil {
		return err
	}

	if err := os.Remove(dbf.path); err != nil {
		return err
	}

	dbf.file = nil
	dbf.entryOffsets = nil
	dbf.fileIndices = nil
	return nil
}

// Ensure unwritten data is written to the file on disk.
func (dbf *DatabaseFile) Flush() error {
	dbf.panicIfNotWriting()
	return dbf.file.Flush()
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

// The number of path info entries that are files.
func (dbf *DatabaseFile) FileEntriesCount() int {
	return int(dbf.header.FileEntriesCount)
}

// Write the path info to the database.
func (dbf *DatabaseFile) WriteEntry(pi *path.Info) error {
	dbf.panicIfNotWriting()

	offset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}
	dbf.entryOffsets = append(dbf.entryOffsets, offset)

	index := dbf.header.EntriesCount

	entry := pathEntryFromPathInfo(pi)
	if err := entry.write(dbf.checksumWriter); err != nil {
		return err
	}

	dbf.header.EntriesCount, err = safe.Add32(dbf.header.EntriesCount, 1)
	if err != nil {
		return err
	}

	if pi.IsFile() {
		dbf.header.FileEntriesCount, err = safe.Add32(dbf.header.FileEntriesCount, 1)
		if err != nil {
			return err
		}

		if dbf.fileIndices != nil {
			dbf.fileIndices = append(dbf.fileIndices, index)
		}
	}

	return nil
}

// Read the path info object with the specified index.
func (dbf *DatabaseFile) ReadEntryAtIndex(idx int) (path.Info, error) {
	if idx >= int(dbf.header.EntriesCount) {
		panic(fmt.Sprintf("invalid index %d, EntriesCount = %d", idx, dbf.header.EntriesCount))
	}

	offset := dbf.entryOffsets[idx]
	_, err := dbf.file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return path.Info{}, fmt.Errorf("failed to read entry at index %d (offset %d). %w", idx, offset, err)
	}
	dbf.file.ResetReadBuffer()

	entry := pathEntry{}
	if err := entry.read(dbf.file); err != nil {
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
	_, err := dbf.file.Seek(int64(dbf.header.EntriesOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read all entries. %w", err)
	}
	dbf.file.ResetReadBuffer()

	for idx := range dbf.header.EntriesCount {
		entry := pathEntry{}
		if err := entry.read(dbf.file); err != nil {
			offset := dbf.file.Offset()
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
	dbf.header.EntriesOffsetTableOffset, err = safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (offset). %w", err)
	}

	if err := dbf.writeEntryOffsets(); err != nil {
		return fmt.Errorf("failed to finish writing the entries (offset table). %w", err)
	}

	dbf.header.FeaturesOffset, err = safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (features offset). %w", err)
	}

	return nil
}

var ErrInvalidChecksum = errors.New("ajfs database file does not match the stored checksum")

// Check the database file integrity and return [ErrInvalidChecksum] if the checksum does not match.
func (dbf *DatabaseFile) VerifyChecksums() error {
	offset := headerOffset() + headerSize()
	_, err := dbf.file.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	dbf.file.ResetReadBuffer()

	count := int64(dbf.header.FeaturesOffset) - offset

	hasher := crc32.NewIEEE()
	_, err = io.CopyN(hasher, dbf.file, count)
	if err != nil {
		return fmt.Errorf("failed to verify checksum. %w", err)
	}

	if hasher.Sum32() != dbf.header.Checksum {
		return ErrInvalidChecksum
	}

	return nil
}

//-----------------------------------------------------------------------------

// Update the header
func (dbf *DatabaseFile) finishCreation() error {
	dbf.panicIfNotWriting()

	if (dbf.header.EntriesCount > 0) && (dbf.header.EntriesOffsetTableOffset == 0) {
		panic("FinishEntries was never called")
	}

	if dbf.createFeatures != dbf.header.Features {
		panic(fmt.Sprintf("not all the expected features were created. expected = %d, actual = %d", dbf.createFeatures, dbf.header.Features))
	}

	if dbf.header.Features.HasHashTable() && (dbf.header.HashTableOffset == 0) {
		panic("hash table was not written")
	}

	dbf.header.Checksum = dbf.checksumHasher.Sum32()

	// Update the header
	_, err := dbf.file.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to finish creating the ajfs database. failed to seek to header offset. %w", err)
	}
	dbf.file.ResetWriteBuffer()

	if err := dbf.header.write(dbf.file); err != nil {
		return fmt.Errorf("failed to update the ajfs header. %w", err)
	}

	if err := dbf.Flush(); err != nil {
		return err
	}

	return nil
}

// Read the entry offset table
func (dbf *DatabaseFile) readEntryOffsets() error {
	if dbf.header.EntriesCount == 0 {
		return nil
	}

	_, err := dbf.file.Seek(int64(dbf.header.EntriesOffsetTableOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read the entry offset table. %w", err)
	}
	dbf.file.ResetReadBuffer()

	// Check 1st sentinel
	var s [4]byte
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		return fmt.Errorf("failed to read the entry offset table (1st sentinel). %w", err)
	}
	if s != sentinel {
		return fmt.Errorf("failed to read the entry offset table (1st sentinel %q does not match %q)", s, sentinel)
	}

	dbf.entryOffsets = make([]uint32, dbf.header.EntriesCount)

	var data [4]byte // uint32
	for i := range dbf.header.EntriesCount {
		_, err := io.ReadFull(dbf.file, data[:])
		if err != nil {
			return fmt.Errorf("failed to read the entry offset table (near index %d). %w", i, err)
		}

		dbf.entryOffsets[i] = binary.LittleEndian.Uint32(data[:])
	}

	// Check 2nd sentinel
	_, err = io.ReadFull(dbf.file, s[:])
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
	_, err := dbf.checksumWriter.Write(sentinel[:])
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (1st sentinel). %w", err)
	}

	data := make([]byte, 4) // uint32

	for idx, offset := range dbf.entryOffsets {
		binary.LittleEndian.PutUint32(data, offset)
		_, err := dbf.checksumWriter.Write(data)
		if err != nil {
			return fmt.Errorf("failed to write the entries offset table (index = %d). %w", idx, err)
		}
	}

	// 2nd sentinel
	_, err = dbf.checksumWriter.Write(sentinel[:])
	if err != nil {
		return fmt.Errorf("failed to finish writing the entries (2nd sentinel). %w", err)
	}

	if err := dbf.Flush(); err != nil {
		return fmt.Errorf("failed to finish writing the entries (flush). %w", err)
	}

	return nil
}

// Panic if the database was not opened for creation (as in file writing)
func (dbf *DatabaseFile) panicIfNotWriting() {
	if !dbf.creating {
		panic("database was not opened for writing")
	}
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
	Checksum                 uint32 // Checksum used to check file integrity.
	EntriesCount             uint32 // The number of path objects. Based on inode max limit of 2^32
	FileEntriesCount         uint32 // The number of path objects that are just files.
	EntriesOffset            uint32 // The offset in bytes at which the path objects start. Based on limit of database file being max 4GB
	EntriesOffsetTableOffset uint32 // The offset to the entries offset table

	Features       FeatureFlags // Feature flags
	FeaturesOffset uint32       // Start of features

	HashTableOffset uint32 // The start of the hash table
	TreeOffset      uint32 // The start of the tree

	FeatureReserved [8]uint32 // 8x feature offsets reserved for future use without breaking backwards compatibility
}

func (s *header) read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *header) write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, s)
}

func headerOffset() int64 {
	return int64(binary.Size(prefixHeader{}))
}

func headerSize() int64 {
	return int64(binary.Size(header{}))
}

//-----------------------------------------------------------------------------
// Root entry

type rootEntry struct {
	// The following fields will be written as the size of the data varint followed by the encoded form of the data
	path string
}

func (s *rootEntry) read(r vardata.Reader) error {
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

func (s *MetaEntry) read(r vardata.Reader) error {
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

func (s *pathEntry) read(r vardata.Reader) error {
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
	FeatureJustEntries = 0         // Contains no extra features. Only path info entries.
	FeatureHashTable   = 1 << iota // Contains the calculated file hash signatures for the path objects.
	FeatureTree                    // Contains the cached file tree.
)

func (f FeatureFlags) HasHashTable() bool {
	return (f & FeatureHashTable) != 0
}

func (f FeatureFlags) HasTree() bool {
	return (f & FeatureTree) != 0
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
	varData   = vardata.NewVariableData()
)

const (
	currentVersion = uint16(1)
)
