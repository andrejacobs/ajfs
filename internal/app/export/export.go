// Package export provides the functionality for ajfs export command.
package export

import (
	"encoding/csv"
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

	if dbf.Features().HasHashTable() {
		//TODO: get info about the hashing algo
		csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "TODO", "Path"})
	} else {
		csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "Path"})

		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			csvWriter.Write([]string{
				fmt.Sprintf("%x", pi.Id),
				fmt.Sprintf("%d", pi.Size),
				pi.Mode.String(),
				pi.ModTime.Format(time.RFC3339Nano),
				pi.Path,
			})
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}

		csvWriter.Flush()
		if err = csvWriter.Error(); err != nil {
			return fmt.Errorf("failed to export to file %q. %w", cfg.ExportPath, err)
		}
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
