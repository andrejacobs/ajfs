// Package scan provides the functionality for ajfs info command.
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
	cfg.Println(fmt.Sprintf("OS:            %s", dbf.Meta().OS))
	cfg.Println(fmt.Sprintf("Architecture:  %s", dbf.Meta().Arch))
	cfg.Println(fmt.Sprintf("Created at:    %s", dbf.Meta().CreatedAt))
	cfg.Println(fmt.Sprintf("Entries:       %d", dbf.EntriesCount()))
	cfg.Println(fmt.Sprintf("File size:     %s", human.Bytes(uint64(fileInfo.Size()))))

	cfg.Println("\nCalculating statistics...")

	stats, err := dbf.CalculateStats()
	if err != nil {
		return fmt.Errorf("failed to calculate statistics. %w", err)
	}

	cfg.Println(fmt.Sprintf("File count:    %d", stats.FileCount))
	cfg.Println(fmt.Sprintf("Dir count:     %d", stats.DirCount))
	cfg.Println(fmt.Sprintf("Total size:    %s [all files toghether]", human.Bytes(stats.TotalFileSize)))
	cfg.Println(fmt.Sprintf("Max file size: %s [single biggest file]", human.Bytes(stats.MaxFileSize)))
	cfg.Println(fmt.Sprintf("Avg file size: %s", human.Bytes(stats.AvgFileSize)))

	return nil
}
