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
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
)

// Config for the ajfs export command.
type Config struct {
	config.CommonConfig

	ExportPath string
	Format     int
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

	// We will be using a bit of manual writing and json encoding
	f := bufio.NewWriter(outFile)

	// Write the header
	fmt.Fprintf(f, "{\n\t\"database\": ")

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

	fmt.Fprintf(f, ",\n\t\"entries\": [\n\t\t")

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
				fmt.Fprintf(f, ",\n\t\t")
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
				fmt.Fprintf(f, ",\n\t\t")
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
	fmt.Fprintf(f, "\n\t]\n}\n")
	if err := f.Flush(); err != nil {
		return fmt.Errorf("failed to create the export file %q. %w", cfg.ExportPath, err)
	}

	return nil
}

//-----------------------------------------------------------------------------
// Hashdeep

func exportHashdeep(cfg Config) error {
	return fmt.Errorf("TODO")
}

//-----------------------------------------------------------------------------
// Constants

const (
	FormatCSV int = iota
	FormatJSON
	FormatHashdeep
)
