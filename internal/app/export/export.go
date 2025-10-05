// Package export provides the functionality for ajfs export command.
package export

import (
	"encoding/csv"
	"encoding/hex"
	"fmt"
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

func exportJSON(cfg Config) error {
	return fmt.Errorf("TODO")
}

func exportHashdeep(cfg Config) error {
	return fmt.Errorf("TODO")
}

const (
	FormatCSV int = iota
	FormatJSON
	FormatHashdeep
)
