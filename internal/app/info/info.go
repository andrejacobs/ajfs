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

// Package info provides the functionality for ajfs info command.
package info

import (
	"fmt"
	"os"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/go-aj/human"
)

// Config for the ajfs info command.
type Config struct {
	config.CommonConfig
}

// Process the ajfs info command.
func Run(cfg Config) error {

	fileInfo, err := os.Stat(cfg.DbPath)
	if err != nil {
		return fmt.Errorf("failed to get ajfs info for %q. %w", cfg.DbPath, err)
	}

	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	cfg.Println(fmt.Sprintf("Database path: %s", dbf.Path()))
	cfg.Println(fmt.Sprintf("Version:       %d", dbf.Version()))
	cfg.Println(fmt.Sprintf("Root path:     %s", dbf.RootPath()))
	cfg.Println(fmt.Sprintf("Tool:          %s", dbf.Meta().Tool))
	cfg.Println(fmt.Sprintf("OS:            %s", dbf.Meta().OS))
	cfg.Println(fmt.Sprintf("Architecture:  %s", dbf.Meta().Arch))
	cfg.Println(fmt.Sprintf("Created at:    %s", dbf.Meta().CreatedAt))
	cfg.Println(fmt.Sprintf("Entries:       %d", dbf.EntriesCount()))
	cfg.Println(fmt.Sprintf("File size:     %s", human.Bytes(uint64(fileInfo.Size())))) //nolint:gosec // disable G115
	cfg.Println(fmt.Sprintf("Features:      0x%x", dbf.Features()))

	if dbf.Features().HasHashTable() {
		cfg.Println("  Hash table:  yes")
		algo, err := dbf.HashTableAlgo()
		if err != nil {
			return err
		}
		cfg.Println("    Algo:      " + algo.String())
	} else {
		cfg.Println("  Hash table:  no")
	}

	cfg.Println("\nVerifying checksum...")
	if err = dbf.VerifyChecksums(); err != nil {
		cfg.Errorln("Invalid checksum!")
		return err
	} else {
		cfg.Println("  Valid checksum")
	}

	cfg.Println("\nCalculating statistics...")

	stats, err := dbf.CalculateStats()
	if err != nil {
		return fmt.Errorf("failed to calculate statistics. %w", err)
	}

	cfg.Println(fmt.Sprintf("File count:    %d", stats.FileCount))
	cfg.Println(fmt.Sprintf("Dir count:     %d", stats.DirCount))
	cfg.Println(fmt.Sprintf("Total size:    %s [all files together]", human.Bytes(stats.TotalFileSize)))
	cfg.Println(fmt.Sprintf("Max file size: %s [single biggest file]", human.Bytes(stats.MaxFileSize)))
	cfg.Println(fmt.Sprintf("Avg file size: %s", human.Bytes(stats.AvgFileSize)))

	// Hash table
	if dbf.Features().HasHashTable() {
		cfg.Println("\nCalculating Hash table statistics...")

		stats, err := dbf.CalculateHashTableStats()
		if err != nil {
			return fmt.Errorf("failed to calculate hash table statistics. %w", err)
		}

		cfg.Println(fmt.Sprintf("Hashed count:    %d", stats.HashedCount))
		cfg.Println(fmt.Sprintf("Pending count:   %d", stats.PendingCount))

		cfg.Println(fmt.Sprintf("Duplicate files: %d", stats.DupesCount))
		cfg.Println(fmt.Sprintf("  Total size:    %s [space taken up by all duplicates]", human.Bytes(stats.TotalDupeSize)))
		cfg.Println(fmt.Sprintf("  Save size:     %s [space that could be freed]", human.Bytes(stats.SaveDupeSize)))
	}

	return nil
}
