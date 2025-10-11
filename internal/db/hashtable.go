package db

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"slices"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// file format
// ... <entries and entries offset table>
// sentinel
// header
// n * hashEntry, where n == number of file path entries
// sentinel

// HashTable maps from path info index to the calculated file signature hash.
type HashTable map[int][]byte

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

	if len(hash) != dbf.createHashTable.header.Algo.Size() {
		panic(fmt.Sprintf("invalid hash size %d, expected size %d", len(hash), dbf.createHashTable.header.Algo.Size()))
	}

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

// Called by EntriesNeedHashing.
// idx Is the index of the path info entry that need it's file signature hash to be calculated.
// pi The path info entry in the database.
// Call WriteHashEntry with the calculated hash.
// Return [SkipAll] to stop processing.
type NeedHashingFn func(idx int, pi path.Info) error

// Look at the hash table and call the passed function for each entry that need the file signature has to be still calculated.
func (dbf *DatabaseFile) EntriesNeedHashing(fn NeedHashingFn) error {
	indices := make([]int, 0, 512)

	err := dbf.ReadHashTableEntries(func(idx int, hash []byte) error {
		if ajhash.AllZeroBytes(hash) {
			indices = append(indices, idx)
		}
		return nil
	})

	if err != nil {
		return err
	}

	for _, idx := range indices {
		pi, err := dbf.ReadEntryAtIndex(idx)
		if err != nil {
			return err
		}

		if err = fn(idx, pi); err != nil {
			if err == SkipAll {
				return nil
			}
			return err
		}
	}

	return nil
}

// Finish writing the hash table.
func (dbf *DatabaseFile) FinishHashTable() error {
	dbf.panicIfNotWriting()

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
	header, err := dbf.readHashTableHeader()
	if err != nil {
		return err
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
	var s [4]byte
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		return fmt.Errorf("failed to read the hash table (2nd sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return fmt.Errorf("failed to read the hash table (2nd sentinel %q does not match %q)", s, hashTableSentinel)
	}

	return nil
}

// Read the hash table.
// Will only contain the entries for which a file signature hash was calculated.
func (dbf *DatabaseFile) ReadHashTable() (HashTable, error) {
	if !dbf.Features().HasHashTable() {
		panic("database does not contain the hash table")
	}

	result := make(HashTable, 64)

	err := dbf.ReadHashTableEntries(func(idx int, hash []byte) error {
		if !ajhash.AllZeroBytes(hash) {
			result[idx] = hash
		}
		return nil
	})

	return result, err
}

// Duplicate hashes is a map from the hash (as hex encoded string) to all the indices of path info entries
// that share the same file signature hash.
type DuplicateHashes map[string][]uint32

// Find all the hashes that are duplicates with the indices to those path info entries.
func (dbf *DatabaseFile) FindDuplicateHashes() (DuplicateHashes, error) {
	if !dbf.Features().HasHashTable() {
		panic("database does not contain the hash table")
	}

	ht, err := dbf.ReadHashTable()
	if err != nil {
		return nil, err
	}

	result := make(DuplicateHashes, 64)

	keys := slices.Sorted(maps.Keys(ht))

	for _, idx := range keys {
		hash := ht[idx]
		hashStr := hex.EncodeToString(hash)

		var dupes []uint32
		var exists bool
		dupes, exists = result[hashStr]
		if !exists {
			dupes = make([]uint32, 0, 4)
		}
		dupes = append(dupes, uint32(idx))
		result[hashStr] = dupes
	}

	// Delete all entries that have only one entry (i.e. non dupe)
	for k, v := range result {
		if len(v) < 2 {
			// Deemed to be safe https://go.dev/doc/effective_go#for to delete from a map while spinning through it.
			delete(result, k)
		}
	}

	return result, nil
}

// FindDuplicatesFn will be called by FindDuplicates for each duplicate file that was found.
// group Each of the same duplicates will belong to the same group.
// idx Is the index of the entry.
// pi Is the path info object.
// hash Is the file signature hash (as a hex encoded string).
// Return [SkipAll] to stop reading all the entries.
type FindDuplicatesFn func(group int, idx int, pi path.Info, hash string) error

// Find duplicate file entries that share the same file signature hash.
func (dbf *DatabaseFile) FindDuplicates(fn FindDuplicatesFn) error {
	if !dbf.Features().HasHashTable() {
		panic("database does not contain the hash table")
	}

	dupes, err := dbf.FindDuplicateHashes()
	if err != nil {
		return err
	}

	keys := slices.Sorted(maps.Keys(dupes))

	group := 0
	for _, hashStr := range keys {
		indices := dupes[hashStr]
		for _, idx := range indices {
			pi, err := dbf.ReadEntryAtIndex(int(idx))
			if err != nil {
				return err
			}

			if err = fn(group, int(idx), pi, hashStr); err != nil {
				if err == SkipAll {
					return nil
				}
				return err
			}
		}
		group++
	}

	return nil
}

// ReadAllEntriesWithHashesFn will be called by ReadAllEntriesWithHashes for each entry that was read from the database.
// idx Is the index of the entry.
// pi Is the path info object.
// hash Is the file signature hash.
// Return [SkipAll] to stop reading all the entries.
type ReadAllEntriesWithHashesFn func(idx int, pi path.Info, hash []byte) error

// Read all the path info objects along with their file signature hash from the database and call the callback function.
// If the callback function returns [SkipAll] then the reading process will be stopped and nil will be returned as the error.
func (dbf *DatabaseFile) ReadAllEntriesWithHashes(fn ReadAllEntriesWithHashesFn) error {
	if !dbf.Features().HasHashTable() {
		panic("database does not contain the hash table")
	}

	hashTable, err := dbf.ReadHashTable()
	if err != nil {
		return err
	}

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		hash, ok := hashTable[idx]
		if !ok {
			return nil
		}
		return fn(idx, pi, hash)
	})
	return err
}

// Read the hash table header and return the hashing algorithm used.
func (dbf *DatabaseFile) HashTableAlgo() (ajhash.Algo, error) {
	header, err := dbf.readHashTableHeader()
	if err != nil {
		return ajhash.AlgoSHA1, err
	}
	return header.Algo, nil
}

// Read the hash table header and do basic validation
func (dbf *DatabaseFile) readHashTableHeader() (hashTableHeader, error) {
	if !dbf.header.Features.HasHashTable() || (dbf.header.HashTableOffset == 0) {
		panic("database contains no hash table")
	}

	_, err := dbf.file.Seek(int64(dbf.header.HashTableOffset), io.SeekStart)
	if err != nil {
		return hashTableHeader{}, fmt.Errorf("failed to read hash table entries. %w", err)
	}
	dbf.file.ResetReadBuffer()

	// Check 1st sentinel
	var s [4]byte
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		return hashTableHeader{}, fmt.Errorf("failed to read the hash table (1st sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return hashTableHeader{}, fmt.Errorf("failed to read the hash table (1st sentinel %q does not match %q)", s, hashTableSentinel)
	}

	// Read the header
	header := hashTableHeader{}
	if err := header.read(dbf.file); err != nil {
		return header, fmt.Errorf("failed to read the hash table header. %w", err)
	}

	if dbf.header.FileEntriesCount != header.EntriesCount {
		return header, fmt.Errorf("the number of hash table entries %d does not match the number of file path entries %d in the database", header.EntriesCount, dbf.header.FileEntriesCount)
	}

	return header, nil
}

// Get the database ready to resume calculating the file signature hashes
func (dbf *DatabaseFile) resumeHashTable() error {

	header, err := dbf.readHashTableHeader()
	if err != nil {
		return err
	}

	dbf.createHashTable = createHashTable{
		header:  header,
		offsets: make(map[uint32]uint32, dbf.header.FileEntriesCount),
	}

	buffer := header.Algo.Buffer()

	// Read the hash entries and construct the offset map
	for i := range header.EntriesCount {
		offset, err := safe.Uint64ToUint32(dbf.file.Offset())
		if err != nil {
			return fmt.Errorf("failed to read the hash table entry at index %d. %w", i, err)
		}

		entry := hashEntry{
			Hash: buffer,
		}
		if err := entry.read(dbf.file); err != nil {
			return fmt.Errorf("failed to read the hash table entry at index %d. %w", i, err)
		}

		dbf.createHashTable.offsets[entry.Index] = offset
	}

	// Check 2nd sentinel
	var s [4]byte
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		return fmt.Errorf("failed to read the hash table (2nd sentinel). %w", err)
	}
	if s != hashTableSentinel {
		return fmt.Errorf("failed to read the hash table (2nd sentinel %q does not match %q)", s, hashTableSentinel)
	}

	return nil
}

//-----------------------------------------------------------------------------
// Helpers

// Map from a path's identifier to the file signature hash.
type IdToHashMap map[path.Id][]byte

// Build a map from a path's identifier to the file signature hash.
func (dbf *DatabaseFile) BuildIdToHashMap() (IdToHashMap, error) {
	result := make(IdToHashMap, dbf.EntriesCount())

	err := dbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
		result[pi.Id] = hash
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// Map from a hash encoded string to the path entry index.
type HashStrToIndexMap map[string]int

// Build a map from a hash encoded string to the path entry index.
func (dbf *DatabaseFile) BuildHashStrToIndexMap() (HashStrToIndexMap, error) {
	result := make(HashStrToIndexMap, dbf.EntriesCount())

	ht, err := dbf.ReadHashTable()
	if err != nil {
		return nil, err
	}

	for k, v := range ht {
		hashStr := hex.EncodeToString(v)
		result[hashStr] = k
	}

	return result, nil
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

	_, err := io.ReadFull(r, s.Hash)
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
