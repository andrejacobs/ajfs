package db

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// file format
// ... <entries and entries offset table>
// sentinel
// header
// n * hashEntry, where n == number of file path entries
// sentinel

//TODO: need a util dbf.HashTableAlgo etc.

//-----------------------------------------------------------------------------
// DatabaseFile

type createHashTable struct {
	header hashTableHeader

	offsets map[uint32]uint32 // map from path entry index to the hash offset
}

// Start writing the initial hash table.
func (dbf *DatabaseFile) StartHashTable(algo ajhash.Algo) error {
	dbf.panicIfNotWriting()

	if !dbf.createFeatures.HasHashTable() {
		panic("database is not expected to have a hash table")
	}

	// Determine the offset
	var err error
	dbf.header.HashTableOffset, err = safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return fmt.Errorf("failed to set the ajfs hash table offset. %w", err)
	}

	// Enable feature
	dbf.header.Features |= FeatureHashTable

	// 1st sentinel
	_, err = dbf.file.Write(hashTableSentinel[:])
	if err != nil {
		return fmt.Errorf("failed to write the hash table (1st sentinel). %w", err)
	}

	// Write header
	dbf.createHashTable = createHashTable{
		header: hashTableHeader{
			Algo:         algo,
			EntriesCount: dbf.header.FileEntriesCount,
		},
		offsets: make(map[uint32]uint32, dbf.header.FileEntriesCount),
	}

	if err := dbf.createHashTable.header.write(dbf.file); err != nil {
		return fmt.Errorf("failed to write the hash table header. %w", err)
	}

	// Write inital empty entries
	zeroHash := algo.ZeroValue()
	for _, idx := range dbf.fileIndices {
		entry := hashEntry{
			Index: idx,
			Hash:  zeroHash,
		}

		offset, err := safe.Uint64ToUint32(dbf.file.Offset())
		if err != nil {
			return fmt.Errorf("failed to write the initial hash table entries (index %d). %w", idx, err)
		}
		dbf.createHashTable.offsets[idx] = offset

		if err := entry.write(dbf.file); err != nil {
			return fmt.Errorf("failed to write the initial hash table entries (index %d). %w", idx, err)
		}
	}

	// 2nd sentinel
	_, err = dbf.file.Write(hashTableSentinel[:])
	if err != nil {
		return fmt.Errorf("failed to write the hash table (1st sentinel). %w", err)
	}

	if err := dbf.file.Flush(); err != nil {
		return fmt.Errorf("failed to write the hash table. %w", err)
	}

	return nil
}

// Write the file hash signature for the path info object with the specified index in the database.
// idx Index of the path info object.
// hash The file hash signature.
func (dbf *DatabaseFile) WriteHashEntry(idx int, hash []byte) error {
	dbf.panicIfNotWriting()

	safeIdx, err := safe.IntToUint32(idx)
	if err != nil {
		return fmt.Errorf("failed to write hash entry for index %d. %w", idx, err)
	}

	offset, ok := dbf.createHashTable.offsets[safeIdx]
	if !ok {
		return fmt.Errorf("failed to write hash entry for index %d, no offset found", idx)
	}

	_, err = dbf.file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to write hash entry for index %d (file seek). %w", idx, err)
	}
	dbf.file.ResetWriteBuffer()

	entry := hashEntry{
		Index: safeIdx,
		Hash:  hash,
	}

	if err := entry.write(dbf.file); err != nil {
		return fmt.Errorf("failed to write hash entry for index %d. %w", idx, err)
	}

	if err := dbf.file.Flush(); err != nil {
		return fmt.Errorf("failed to write hash entry for index %d. %w", idx, err)
	}

	return nil
}

// Finish writing the hash table.
func (dbf *DatabaseFile) FinishHashTable() error {
	dbf.panicIfNotWriting()
	//TODO: need a way for the finishCreation to check this was called or panic

	if err := dbf.Flush(); err != nil {
		return fmt.Errorf("failed to finish writing the hash table (flush). %w", err)
	}

	return nil
}

// ReadHashTableEntryFn will be called by ReadHashTableEntries for each hash table entry that was read from the database.
// idx Is the index of the hash table entry which also maps 1:1 to the path entry index.
// hash Is the file hash signature.
// Return [SkipAll] to stop reading further entries.
type ReadHashTableEntryFn func(idx int, hash []byte) error

// Read all hash table entries from the database and call the callback function.
// If the callback function returns [SkipAll] then the reading process will be stopped and nil will be returned as the error.
func (dbf *DatabaseFile) ReadHashTableEntries(fn ReadHashTableEntryFn) error {
	dbf.panicIfNotReading()

	if !dbf.header.Features.HasHashTable() || (dbf.header.HashTableOffset == 0) {
		panic("database contains no hash table")
	}

	_, err := dbf.file.Seek(int64(dbf.header.HashTableOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read hash table entries. %w", err)
	}
	dbf.file.ResetReadBuffer()

	// Check 1st sentinel
	var s [4]byte
	_, err = dbf.file.Read(s[:])
	if err != nil {
		return fmt.Errorf("failed to read the hash table (1st sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return fmt.Errorf("failed to read the hash table (1st sentinel %q does not match %q)", s, hashTableSentinel)
	}

	// Read the header
	header := hashTableHeader{}
	if err := header.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the hash table header. %w", err)
	}

	if dbf.header.FileEntriesCount != header.EntriesCount {
		return fmt.Errorf("the number of hash table entries %d does not match the number of file path entries %d in the database", header.EntriesCount, dbf.header.FileEntriesCount)
	}

	// Read the hash entries
	for i := range header.EntriesCount {
		entry := hashEntry{
			Hash: header.Algo.Buffer(),
		}
		if err := entry.read(dbf.file); err != nil {
			return fmt.Errorf("failed to read the hash table entry at index %d. %w", i, err)
		}

		idx, err := safe.Uint32ToInt(entry.Index)
		if err != nil {
			return fmt.Errorf("failed to read the hash table entry at index %d (path entry index %d will cause integer overflow). %w", i, entry.Index, err)
		}

		if err := fn(idx, entry.Hash); err != nil {
			if err == SkipAll {
				return nil
			}
			return err
		}
	}

	// Check 2nd sentinel
	_, err = dbf.file.Read(s[:])
	if err != nil {
		return fmt.Errorf("failed to read the hash table (2nd sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return fmt.Errorf("failed to read the hash table (2nd sentinel %q does not match %q)", s, hashTableSentinel)
	}

	return nil
}

//-----------------------------------------------------------------------------
// Header

type hashTableHeader struct {
	Algo         ajhash.Algo
	EntriesCount uint32 // This must match the db Header's EntriesCount
}

func (s *hashTableHeader) read(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *hashTableHeader) write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, s)
}

//-----------------------------------------------------------------------------
// Hash entry

type hashEntry struct {
	Index uint32 // Index of the matching file path entry
	Hash  []byte // File signature hash
}

func (s *hashEntry) read(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &s.Index); err != nil {
		return err
	}

	_, err := r.Read(s.Hash)
	return err
}

func (s *hashEntry) write(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, s.Index); err != nil {
		return err
	}

	_, err := w.Write(s.Hash)
	return err
}

//-----------------------------------------------------------------------------
// Constants and Misc

var (
	hashTableSentinel = [4]byte{0x41, 0x4A, 0x48, 0x58} // AJHX
)
