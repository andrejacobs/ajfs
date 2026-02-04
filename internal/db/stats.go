// Copyright (c) 2025 Andre Jacobs
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
	"fmt"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// Stats is used to calculate statistics on the database.
type Stats struct {
	DirCount  uint64 // total number of directories
	FileCount uint64 // total number of files

	TotalFileSize uint64 // total size of files all summed together
	AvgFileSize   uint64 // totalFileSize / fileCount

	MaxFileSize uint64 // the biggest single file size
}

// Calculate statistics on the database.
func (dbf *DatabaseFile) CalculateStats() (Stats, error) {
	result := Stats{}

	err := dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if pi.IsDir() {
			result.DirCount++
		} else if pi.IsFile() {
			result.FileCount++
			result.TotalFileSize += pi.Size
			result.MaxFileSize = max(result.MaxFileSize, pi.Size)
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("failed to calculate statistics for %q. %w", dbf.Path(), err)
	}

	if result.FileCount > 0 {
		result.AvgFileSize = result.TotalFileSize / result.FileCount
	}

	return result, nil
}

// Stats used to calculate statistics on the hash table.
type HashTableStats struct {
	HashedCount  uint64 // number of entries that have a calculated hash
	PendingCount uint64 // number of entries that still need to be calculated

	DupesCount    uint64 // number of duplicate files found
	TotalDupeSize uint64 // total bytes of space used by found duplicates
	SaveDupeSize  uint64 // total bytes that could be saved after removing duplicates
}

// Calculate statistics for the hash table.
func (dbf *DatabaseFile) CalculateHashTableStats() (HashTableStats, error) {
	if !dbf.Features().HasHashTable() {
		panic("database does not contain the hash table")
	}

	stats := HashTableStats{}

	err := dbf.ReadHashTableEntries(func(idx int, hash []byte) error {
		if ajhash.AllZeroBytes(hash) {
			stats.PendingCount++
		} else {
			stats.HashedCount++
		}
		return nil
	})

	if err != nil {
		return HashTableStats{}, err
	}

	singleSizes := make(map[int]uint64, 64)

	err = dbf.FindDuplicates(func(group, idx int, pi path.Info, hash string) error {
		stats.DupesCount++

		var err error
		stats.TotalDupeSize, err = safe.Add64(stats.TotalDupeSize, pi.Size)
		if err != nil {
			return err
		}

		singleSizes[group] = pi.Size

		return nil
	})

	if err != nil {
		return HashTableStats{}, err
	}

	stats.SaveDupeSize = stats.TotalDupeSize
	for _, v := range singleSizes {
		stats.SaveDupeSize, err = safe.Sub64(stats.SaveDupeSize, v)
		if err != nil {
			return HashTableStats{}, err
		}
	}

	return stats, nil
}
