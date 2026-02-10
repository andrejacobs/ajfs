// Copyright (c) 2026 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package db

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"slices"

	"github.com/andrejacobs/go-aj/ajio/trackedoffset"
	"github.com/andrejacobs/go-aj/ajmath/safe"
	"github.com/andrejacobs/go-aj/file"
)

// Attempts to repair a damaged database.
// out is used to display information to the user (normally routed to STDOUT). Things to be fixed will be prefixed with >>.
// path is the file path to an existing database file.
// dryRun when set to true will only output issues to the output writer and not make any changes.
// bakPath path to where the backup file will be created. NOTE: only the headers are saved.
func FixDatabase(out io.Writer, dbPath string, dryRun bool, bakPath string) error {
	// > OpenDatabase -----------------------------------------------

	dbf := &DatabaseFile{
		path: dbPath,
	}

	var err error
	dbf.file, err = trackedoffset.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open the ajfs database file. path: %q. %w", dbPath, err)
	}

	// > readHeadersAndVerify ---------------------------------------

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

	fmt.Fprintf(out, "Signature: %s\n", string(dbf.prefixHeader.Signature[:]))
	fmt.Fprintf(out, "Version: %d\n", dbf.prefixHeader.Version)

	// Read the header
	if err := dbf.header.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs header. path: %q. %w", dbf.path, err)
	}

	fixHeader := dbf.header

	checksumHasher := crc32.NewIEEE()

	// Read the root info
	if err := dbf.root.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs root entry. path: %q. %w", dbf.path, err)
	}
	_ = dbf.root.write(checksumHasher)

	fmt.Fprintf(out, "Root: %q\n", dbf.root.path)

	// Read the meta info
	if err := dbf.meta.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs meta entry. path: %q. %w", dbf.path, err)
	}
	_ = dbf.meta.write(checksumHasher)

	fmt.Fprintf(out, "Meta | OS: %q\n", dbf.meta.OS)
	fmt.Fprintf(out, "Meta | Arch: %q\n", dbf.meta.Arch)
	fmt.Fprintf(out, "Meta | Created at: %q\n", dbf.Meta().CreatedAt)

	//AJ### TODO: Meta | Tool: " but I need to merge first

	// Read entries -------------------------------------------------
	entriesOffset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}

	if dbf.header.EntriesOffset != entriesOffset {
		fixHeader.EntriesOffset = entriesOffset
		fmt.Fprintf(out, ">> Entries offset is expected to be 0x%x, actual is 0x%x\n", entriesOffset, dbf.header.EntriesOffset)
	}

	fmt.Fprintf(out, "Entries offset: 0x%x\n", entriesOffset)

	keepGoing := true
	entriesCount := uint32(0)
	fileEntriesCount := uint32(0)
	expectedEntryLookups := make([]entryLookup, 0, 64)
	fileIndices := make([]uint32, 0, 64)
	var s [4]byte

	for keepGoing {
		offset, err := safe.Uint64ToUint32(dbf.file.Offset())
		if err != nil {
			return err
		}

		entry := pathEntry{}
		if err := entry.read(dbf.file); err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("database is corrupted. reached EOF while reading the entries")
			}

			return fmt.Errorf("failed to read entry at index %d (offset %d). %w", entriesCount, offset, err)
		}
		entriesCount++
		_ = entry.write(checksumHasher)

		expectedEntryLookups = append(expectedEntryLookups, entryLookup{
			Id:     entry.header.Id,
			Offset: offset,
		})

		if entry.header.Mode.IsRegular() {
			fileEntriesCount++
			fileIndices = append(fileIndices, entriesCount-1)
		}

		// Check for entries lookup table sentinel
		buf, err := dbf.file.Peek(4)
		if err != nil {
			return fmt.Errorf("failed to check for the entry lookup table (1st sentinel). %w", err)
		}

		if bytes.Equal(buf, sentinel[:]) {
			keepGoing = false
			_, _ = checksumHasher.Write(sentinel[:])
			_, err = dbf.file.Discard(4)
			if err != nil {
				return fmt.Errorf("failed to discard 4 bytes while looking for the entries offset table. %w", err)
			}
		}
	}

	if dbf.header.EntriesCount != entriesCount {
		fixHeader.EntriesCount = entriesCount
		fmt.Fprintf(out, ">> Entries count is expected to be %d, actual is %d\n", entriesCount, dbf.header.EntriesCount)
	}

	if dbf.header.FileEntriesCount != fileEntriesCount {
		fixHeader.FileEntriesCount = fileEntriesCount
		fmt.Fprintf(out, ">> File entries count is expected to be %d, actual is %d\n", fileEntriesCount, dbf.header.FileEntriesCount)
	}

	fmt.Fprintf(out, "Entries: %d\nFiles: %d\n", entriesCount, fileEntriesCount)

	// Read entries lookup table ------------------------------------
	entriesLookupTableOffset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}
	entriesLookupTableOffset -= 4

	if dbf.header.EntriesLookupTableOffset != entriesLookupTableOffset {
		fixHeader.EntriesLookupTableOffset = entriesLookupTableOffset
		fmt.Fprintf(out, ">> Entries lookup table offset is expected to be 0x%x, actual is 0x%x\n", entriesLookupTableOffset, dbf.header.EntriesLookupTableOffset)
	}

	fmt.Fprintf(out, "Entries lookup table offset: 0x%x\n", entriesLookupTableOffset)

	entryLookups := make([]entryLookup, entriesCount)

	for i := range entriesCount {
		entry := &entryLookups[i]

		err := entry.read(dbf.file)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("database is corrupted. reached EOF while reading the entries lookup table")
			}
			return fmt.Errorf("failed to read the entry lookup table (near index %d). %w", i, err)
		}
		_ = entry.write(checksumHasher)
	}

	// Check 2nd sentinel
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		return fmt.Errorf("failed to read the entry lookup table (2nd sentinel). %w", err)
	}
	if s != sentinel {
		return fmt.Errorf("failed to read the entry lookup table (2nd sentinel %q does not match %q)", s, sentinel)
	}
	_, _ = checksumHasher.Write(sentinel[:])

	if len(expectedEntryLookups) != len(entryLookups) {
		return fmt.Errorf("database is corrupted. expected %d entries in the entries lookup table, actual is %d", len(expectedEntryLookups), len(entryLookups))
	}

	featuresOffset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}

	if dbf.header.FeaturesOffset != featuresOffset {
		fixHeader.FeaturesOffset = featuresOffset
		fmt.Fprintf(out, ">> Features offset is expected to be 0x%x, actual is 0x%x\n", featuresOffset, dbf.header.FeaturesOffset)
	}

	for i := range expectedEntryLookups {
		lhs := expectedEntryLookups[i]
		rhs := entryLookups[i]

		if lhs.Id != rhs.Id {
			return fmt.Errorf("database is corrupted. expected entry lookup at index %d to have path Id 0x%x, actual is 0x%x", i, lhs.Id, rhs.Id)
		}

		if lhs.Offset != rhs.Offset {
			return fmt.Errorf("database is corrupted. expected entry lookup at index %d to have offset 0x%x, actual is 0x%x", i, lhs.Offset, rhs.Offset)
		}
	}

	// Check checksum -----------------------------------------------
	expectedChecksum := checksumHasher.Sum32()
	if expectedChecksum != dbf.header.Checksum {
		fixHeader.Checksum = expectedChecksum
		fmt.Fprintf(out, ">> Checksum is expected to be 0x%x, actual is 0x%x\n", expectedChecksum, dbf.header.Checksum)
	}

	fmt.Fprintf(out, "Checksum: 0x%x\n", expectedChecksum)

	// Check the hash table if present ------------------------------
	hashTableOffset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}

	eof := false

	// 1st sentinel
	_, err = io.ReadFull(dbf.file, s[:])
	if err != nil {
		if errors.Is(err, io.EOF) {
			eof = true

			if dbf.Features().HasHashTable() {
				return fmt.Errorf("database is corrupted. expected a hash table to be present")
			}
			// this is fine, EOF and not expecting a hash table, continue
		} else {
			return fmt.Errorf("failed to read the hash table (1st sentinel). %w", err)
		}
	}

	if !eof {
		fmt.Fprintln(out, "Hash table: Yes")

		// Hash table checks
		if s != hashTableSentinel {
			return fmt.Errorf("database is corrupted. expected hash table sentinel 0x%x, actual 0x%x)", hashTableSentinel, s)
		}

		fixHeader.Features |= FeatureHashTable

		if hashTableOffset != dbf.header.HashTableOffset {
			fixHeader.HashTableOffset = hashTableOffset
			fmt.Fprintf(out, ">> Hash table offset is expected to be 0x%x, actual is 0x%x\n", hashTableOffset, dbf.header.HashTableOffset)
		}

		fmt.Fprintf(out, "Hash table offset: 0x%x\n", hashTableOffset)

		header := hashTableHeader{}
		if err := header.read(dbf.file); err != nil {
			return fmt.Errorf("failed to read the hash table header. %w", err)
		}

		fmt.Fprintf(out, "Hash algorithm: %s\n", header.Algo)

		if fileEntriesCount != header.EntriesCount {
			return fmt.Errorf("database is corrupted. the number of hash table entries %d does not match the number of file path entries %d in the database", header.EntriesCount, fileEntriesCount)
		}

		hashFileIndices := make([]uint32, 0, 64)

		for i := range header.EntriesCount {
			entry := hashEntry{
				Hash: header.Algo.Buffer(),
			}
			if err := entry.read(dbf.file); err != nil {
				if errors.Is(err, io.EOF) {
					return fmt.Errorf("database is corrupted. reached EOF while reading the hash table entries")
				}
				return fmt.Errorf("failed to read the hash table entry at index %d. %w", i, err)
			}
			hashFileIndices = append(hashFileIndices, entry.Index)
		}

		// 2nd sentinel
		_, err = io.ReadFull(dbf.file, s[:])
		if err != nil {
			return fmt.Errorf("failed to read the hash table (2nd sentinel). %w", err)
		}
		if s != hashTableSentinel {
			return fmt.Errorf("failed to read the hash table (2nd sentinel %q does not match %q)", s, hashTableSentinel)
		}

		// Validate indices
		slices.Sort(fileIndices)
		slices.Sort(hashFileIndices)
		if !slices.Equal(fileIndices, hashFileIndices) {
			return fmt.Errorf("database is corrupted. file indices does not match hash table's file indices")
		}
	} else {
		fmt.Fprintln(out, "Hash table: No")
	}

	if err := dbf.file.Close(); err != nil {
		return err
	}

	needFixing := fixHeader != dbf.header

	// Dry-run / validate finished, next is actual file changes
	if dryRun {
		if needFixing {
			fmt.Fprintln(out, "Database needs to be fixed. Skipping because running in dry-run mode.")
			return fmt.Errorf("database needs to be fixed")
		} else {
			fmt.Fprintln(out, "Nothing to be fixed")
			return nil
		}
	}
	//=========================================================================

	if !needFixing {
		fmt.Fprintln(out, "Nothing to be fixed")
		return nil
	}

	// Make backup of the headers
	fmt.Fprintf(out, "Backing up headers to: %q\n", bakPath)

	if err = saveDatabaseHeaders(dbPath, bakPath); err != nil {
		return err
	}

	f, err := trackedoffset.OpenFile(dbPath, os.O_RDWR|os.O_EXCL, 0)
	if err != nil {
		return fmt.Errorf("failed to open the database for applying fixes. %w", err)
	}
	defer f.Close()

	_, err = f.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return err
	}

	if err = fixHeader.write(f); err != nil {
		return fmt.Errorf("failed to write the fixed header to the database. %w", err)
	}

	if err = f.Flush(); err != nil {
		return err
	}

	return nil
}

// Restore the headers from a backup file.
func RestoreDatabaseHeader(dbPath string, bakPath string) error {

	bakHeader, err := readHeader(bakPath)
	if err != nil {
		return fmt.Errorf("not a valid backup file. %w", err)
	}

	_, err = readHeader(dbPath)
	if err != nil {
		return err
	}

	return replaceHeader(bakHeader, dbPath)
}

//-----------------------------------------------------------------------------

func saveDatabaseHeaders(dbPath string, bakPath string) error {
	bakSize := headerOffset() + headerSize()
	_, err := file.CopyFileN(context.Background(), dbPath, bakPath, bakSize)
	if err != nil {
		return fmt.Errorf("failed to make a backup of the headers. %w", err)
	}
	return nil
}

func readHeader(dbPath string) (header, error) {
	f, err := os.Open(dbPath)
	if err != nil {
		return header{}, err
	}
	defer f.Close()

	// readHeadersAndVerify

	// Check the signature and version
	var ph prefixHeader
	if err := ph.read(f); err != nil {
		return header{}, fmt.Errorf("error reading the ajfs prefix header. path: %q. %w", dbPath, err)
	}
	if ph.Signature != signature {
		return header{}, fmt.Errorf("not a valid ajfs file (invalid signature %q, expected %q). path: %q", ph.Signature, signature, dbPath)
	}
	if ph.Version > currentVersion {
		return header{}, fmt.Errorf("not a supported ajfs file (invalid version %d, expected <= %d). path: %q", ph.Version, currentVersion, dbPath)
	}

	// Read the header
	var result header
	if err := result.read(f); err != nil {
		return header{}, fmt.Errorf("failed to read the ajfs header. path: %q. %w", dbPath, err)

	}
	return result, err
}

func replaceHeader(newHeader header, dbPath string) error {
	f, err := os.OpenFile(dbPath, os.O_RDWR|os.O_EXCL, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Seek(headerOffset(), io.SeekStart)
	if err != nil {
		return err
	}

	return newHeader.write(f)
}
