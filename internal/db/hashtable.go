package db

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/ajmath"
)

// file format
// ... <entries and entries offset table>
// sentinel
// header
// n * hash
// sentinel

//TODO: need a util dbf.HashTableAlgo etc.

//-----------------------------------------------------------------------------
// DatabaseFile

// Start writing the initial hash table.
func (dbf *DatabaseFile) StartHashTable(algo ajhash.Algo) error {
	dbf.panicIfNotWriting()

	// Determine the offset
	var err error
	dbf.header.HashTableOffset, err = ajmath.Uint64ToUint32(dbf.currentWriteOffset())
	if err != nil {
		return fmt.Errorf("failed to set the ajfs hash table offset. %w", err)
	}

	// Enable feature
	dbf.header.Features |= featureHashTable

	// 1st sentinel
	_, err = dbf.writer.Write(hashTableSentinel[:])
	if err != nil {
		return fmt.Errorf("failed to write the hash table (1st sentinel). %w", err)
	}

	// Write header
	header := hashTableHeader{
		Algo:         algo,
		EntriesCount: dbf.header.EntriesCount,
	}

	if err := header.write(dbf.writer); err != nil {
		return fmt.Errorf("failed to write the hash table header. %w", err)
	}

	// Write inital empty entries
	for i := range header.EntriesCount {
		if _, err := dbf.writer.Write(algo.ZeroValue()); err != nil {
			return fmt.Errorf("failed to write the initial hash table entries (index %d). %w", i, err)
		}
	}

	// 2nd sentinel
	_, err = dbf.writer.Write(hashTableSentinel[:])
	if err != nil {
		return fmt.Errorf("failed to write the hash table (1st sentinel). %w", err)
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

	_, err := dbf.reader.Seek(int64(dbf.header.HashTableOffset), io.SeekStart)
	if err != nil {
		return fmt.Errorf("failed to read hash table entries. %w", err)
	}

	// Check 1st sentinel
	var s [4]byte
	_, err = dbf.reader.Read(s[:])
	if err != nil {
		return fmt.Errorf("failed to read the hash table (1st sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return fmt.Errorf("failed to read the hash table (1st sentinel %q does not match %q)", s, hashTableSentinel)
	}

	// Read the header
	header := hashTableHeader{}
	if err := header.read(dbf.reader); err != nil {
		return fmt.Errorf("failed to read the hash table header. %w", err)
	}

	if dbf.header.EntriesCount != header.EntriesCount {
		return fmt.Errorf("the number of hash table entries %d does not match the number of path entries %d in the database", header.EntriesCount, dbf.header.EntriesCount)
	}

	// Read the hash entries
	hash := make([]byte, header.Algo.Size())

	for idx := range header.EntriesCount {
		_, err = dbf.reader.Read(hash)
		if err != nil {
			return fmt.Errorf("failed to read the hash table entry at index %d. %w", idx, err)
		}

		if err := fn(int(idx), hash); err != nil {
			if err == SkipAll {
				return nil
			}
			return err
		}
	}

	// Check 2nd sentinel
	_, err = dbf.reader.Read(s[:])
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

func (s *hashTableHeader) read(r io.ReadSeeker) error {
	return binary.Read(r, binary.LittleEndian, s)
}

func (s *hashTableHeader) write(w io.Writer) error {
	return binary.Write(w, binary.LittleEndian, s)
}

//-----------------------------------------------------------------------------
// Constants and Misc

var (
	hashTableSentinel = [4]byte{0x41, 0x4A, 0x48, 0x58} // AJHX
)
