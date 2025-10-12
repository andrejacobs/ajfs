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

// Package export provides the functionality for ajfs export command.
package export

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
)

// Config for the ajfs export command.
type Config struct {
	config.CommonConfig

	ExportPath string
	Format     int
	FullPaths  bool
}

// Process the ajfs export command.
func Run(cfg Config) error {
	switch cfg.Format {
	case FormatCSV:
		return exportCSV(cfg)
	case FormatJSON:
		return exportJSON(cfg)
	case FormatHashdeep:
		return exportHashdeep(cfg)
	}

	return fmt.Errorf("invalid export format %v", cfg.Format)
}

//-----------------------------------------------------------------------------
// CSV

func exportCSV(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	outFile, err := os.OpenFile(cfg.ExportPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}
	defer outFile.Close()

	cfg.VerbosePrintln(fmt.Sprintf("Exporting database %q to CSV file %q", cfg.DbPath, cfg.ExportPath))

	csvWriter := csv.NewWriter(outFile)

	// With a hash table
	if dbf.Features().HasHashTable() {
		algo, err := dbf.HashTableAlgo()
		if err != nil {
			return err
		}

		hashTable, err := dbf.ReadHashTable()
		if err != nil {
			return err
		}

		csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "IsDir", "Hash (" + algo.String() + ")", "Path"})

		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			var hashStr string
			if !pi.IsDir() {
				hash, ok := hashTable[idx]

				if ok {
					hashStr = hex.EncodeToString(hash)
				}
			}

			if cfg.FullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			err := csvWriter.Write([]string{
				fmt.Sprintf("%x", pi.Id),
				fmt.Sprintf("%d", pi.Size),
				pi.Mode.String(),
				pi.ModTime.Format(time.RFC3339Nano),
				fmt.Sprintf("%t", pi.IsDir()),
				hashStr,
				pi.Path,
			})
			if err != nil {
				return err
			}

			csvWriter.Flush()
			return csvWriter.Error()
		})
		if err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}
	} else {
		// Without a hash table
		csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "IsDir", "Path"})

		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			if cfg.FullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			err := csvWriter.Write([]string{
				fmt.Sprintf("%x", pi.Id),
				fmt.Sprintf("%d", pi.Size),
				pi.Mode.String(),
				pi.ModTime.Format(time.RFC3339Nano),
				fmt.Sprintf("%t", pi.IsDir()),
				pi.Path,
			})
			if err != nil {
				return err
			}

			csvWriter.Flush()
			return csvWriter.Error()
		})
		if err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}
	}

	csvWriter.Flush()
	if err = csvWriter.Error(); err != nil {
		return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
	}

	cfg.VerbosePrintln("Done!")
	return nil
}

//-----------------------------------------------------------------------------
// JSON

type jsonEntry struct {
	Id      string      `json:"id"`
	Path    string      `json:"path"`
	Size    uint64      `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	ModeStr string      `json:"modeStr"`
	ModTime time.Time   `json:"modTime"`

	Hash string `json:"hash,omitempty"`
}

func exportJSON(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	outFile, err := os.OpenFile(cfg.ExportPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}
	defer outFile.Close()

	cfg.VerbosePrintln(fmt.Sprintf("Exporting database %q to JSON file %q", cfg.DbPath, cfg.ExportPath))

	// We will be using a bit of manual writing and json encoding
	f := bufio.NewWriter(outFile)

	// Write the header
	_, err = fmt.Fprintf(f, "{\n\t\"database\": ")
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	var hashAlgo string
	if dbf.Features().HasHashTable() {
		algo, err := dbf.HashTableAlgo()
		if err != nil {
			return err
		}
		hashAlgo = algo.String()
	}

	data, err := json.MarshalIndent(struct {
		Version          int             `json:"version"`
		DbPath           string          `json:"dbPath"`
		Root             string          `json:"root"`
		Features         db.FeatureFlags `json:"features"`
		EntriesCount     int             `json:"entriesCount"`
		FileEntriesCount int             `json:"fileCount"`
		Meta             db.MetaEntry    `json:"meta"`
		HashTableAlgo    string          `json:"hashTableAlgo,omitempty"`
	}{
		Version:          dbf.Version(),
		DbPath:           dbf.Path(),
		Root:             dbf.RootPath(),
		Features:         dbf.Features(),
		EntriesCount:     dbf.EntriesCount(),
		FileEntriesCount: dbf.FileEntriesCount(),
		Meta:             dbf.Meta(),
		HashTableAlgo:    hashAlgo,
	}, "\t", "\t")
	if err != nil {
		return fmt.Errorf("failed to export json. encoding of header failed. %w", err)
	}
	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("failed to export json. writing of header failed. %w", err)
	}
	if err = f.Flush(); err != nil {
		return fmt.Errorf("failed to export json. writing of header failed. %w", err)
	}

	_, err = fmt.Fprintf(f, ",\n\t\"entries\": [\n\t\t")
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	// With a hash table
	if dbf.Features().HasHashTable() {
		hashTable, err := dbf.ReadHashTable()
		if err != nil {
			return err
		}

		count := 0
		expectedCount := dbf.EntriesCount()

		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			var hashStr string
			if !pi.IsDir() {
				hash, ok := hashTable[idx]

				if ok {
					hashStr = hex.EncodeToString(hash)
				}
			}

			if cfg.FullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			data, err := json.MarshalIndent(jsonEntry{
				Id:      hex.EncodeToString(pi.Id[:]),
				Path:    pi.Path,
				Size:    pi.Size,
				Mode:    pi.Mode,
				ModeStr: pi.Mode.String(),
				ModTime: pi.ModTime,
				Hash:    hashStr,
			}, "\t\t", "\t")

			if err != nil {
				return fmt.Errorf("failed to export json. encoding entry (index = %d) failed. %w", idx, err)
			}
			_, err = f.Write(data)
			if err != nil {
				return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
			}

			count++
			if count < expectedCount {
				_, err = fmt.Fprintf(f, ",\n\t\t")
				if err != nil {
					return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
				}
			}

			if err = f.Flush(); err != nil {
				return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
			}

			return err

		})
		if err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}

	} else {
		// Without a hash table
		count := 0
		expectedCount := dbf.EntriesCount()

		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			if cfg.FullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			data, err := json.MarshalIndent(jsonEntry{
				Id:      hex.EncodeToString(pi.Id[:]),
				Path:    pi.Path,
				Size:    pi.Size,
				Mode:    pi.Mode,
				ModeStr: pi.Mode.String(),
				ModTime: pi.ModTime,
			}, "\t\t", "\t")

			if err != nil {
				return fmt.Errorf("failed to export json. encoding entry (index = %d) failed. %w", idx, err)
			}
			_, err = f.Write(data)
			if err != nil {
				return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
			}

			count++
			if count < expectedCount {
				_, err = fmt.Fprintf(f, ",\n\t\t")
				if err != nil {
					return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
				}
			}

			if err = f.Flush(); err != nil {
				return fmt.Errorf("failed to export json. writing entry (index = %d) failed. %w", idx, err)
			}

			return err

		})
		if err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}
	}

	// Finish up
	_, err = fmt.Fprintf(f, "\n\t]\n}\n")
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	if err := f.Flush(); err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	cfg.VerbosePrintln("Done!")
	return nil
}

//-----------------------------------------------------------------------------
// Hashdeep

func exportHashdeep(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	if !dbf.Features().HasHashTable() {
		return fmt.Errorf("failed to create the export file %q because the ajfs database %q does not contain a hash table",
			cfg.ExportPath, cfg.DbPath)
	}

	algo, err := dbf.HashTableAlgo()
	if err != nil {
		return err
	}

	cfg.VerbosePrintln(fmt.Sprintf("Exporting database %q to hashdeep file %q", cfg.DbPath, cfg.ExportPath))

	outFile, err := os.OpenFile(cfg.ExportPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}
	defer outFile.Close()

	f := bufio.NewWriter(outFile)

	// Write header
	_, err = fmt.Fprintf(f, "%%%%%%%% HASHDEEP-1.0\n")
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	var hashStr string
	switch algo {
	case ajhash.AlgoSHA1:
		hashStr = "sha1"
	case ajhash.AlgoSHA256:
		hashStr = "sha256"
	default:
		return fmt.Errorf("failed to create the export file %q. hashdeep does not support %q", cfg.ExportPath, algo.String())
	}

	_, err = fmt.Fprintf(f, "%%%%%%%% size,%s,filename\n", hashStr)
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	_, err = fmt.Fprintf(f, "## Generated by: ajfs export --format=hashdeep %q\n", cfg.DbPath)
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	_, err = fmt.Fprintf(f, "## Invoked from: %s\n##\n", dbf.RootPath())
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	err = dbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
		hashStr := hex.EncodeToString(hash)

		var err error
		if cfg.FullPaths {
			pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			_, err = fmt.Fprintf(f, "%d,%s,%s\n", pi.Size, hashStr, pi.Path)
		} else {
			_, err = fmt.Fprintf(f, "%d,%s,./%s\n", pi.Size, hashStr, pi.Path)
		}

		return err
	})
	if err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	if err := f.Flush(); err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	cfg.VerbosePrintln("Done!")
	return nil
}

//-----------------------------------------------------------------------------
// Constants

const (
	FormatCSV int = iota
	FormatJSON
	FormatHashdeep
)
