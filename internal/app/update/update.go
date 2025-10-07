package update

import (
	"errors"
	"fmt"
	"os"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/resume"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
)

// Config for the ajfs update command.
type Config struct {
	config.CommonConfig
	config.FilterConfig
}

// Process the ajfs update command.
func Run(cfg Config) error {
	cfg.VerbosePrintln(fmt.Sprintf("Updating database file at %q", cfg.DbPath))

	// Rename existing file
	backupDbPath := cfg.DbPath + ".bak"
	cfg.VerbosePrintln(fmt.Sprintf("Backing up current database to: %q", backupDbPath))
	if err := os.Rename(cfg.DbPath, backupDbPath); err != nil {
		return fmt.Errorf("failed to backup database file %q to %q. %w", cfg.DbPath, backupDbPath, err)
	}

	var newDbf *db.DatabaseFile

	// Called when an error happened and we need to restore the backup
	errFn := func(rcvErr error) error {
		cfg.Errorln(rcvErr)

		if newDbf != nil {
			_ = newDbf.Interrupted()
		}

		if err := os.Rename(backupDbPath, cfg.DbPath); err != nil {
			return fmt.Errorf("failed to restore the backup file with error (%w). original error: %w", err, rcvErr)
		}
		return rcvErr
	}

	// Perform new scan
	oldDbf, err := db.OpenDatabase(backupDbPath)
	if err != nil {
		return errFn(err)
	}
	defer oldDbf.Close()

	scanCfg := scan.Config{
		CommonConfig: cfg.CommonConfig,
		FilterConfig: cfg.FilterConfig,
		Root:         oldDbf.RootPath(),
		InitOnly:     true,
	}

	if oldDbf.Features().HasHashTable() {
		scanCfg.CalculateHashes = true
		scanCfg.Algo, err = oldDbf.HashTableAlgo()
		if err != nil {
			return errFn(err)
		}
	}

	if err = scan.Run(scanCfg); err != nil {
		return errFn(err)
	}

	// Copy existing hashes over for matching entries
	if oldDbf.Features().HasHashTable() {
		newDbf, err = db.ResumeDatabase(cfg.DbPath)
		if err != nil {
			return errFn(err)
		}

		err = oldDbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
			v, err := newDbf.FindEntryIndexAndOffset(pi.Id)
			if err != nil {
				if !errors.Is(err, db.ErrNotFound) {
					return err
				}
				// Entry no longer exists in new database
				return nil
			}

			return newDbf.WriteHashEntry(int(v.Index), hash)
		})
		if err != nil {
			return errFn(err)
		}

		if err = newDbf.Close(); err != nil {
			return errFn(err)
		}

		// Start hashing new entries
		resumeCfg := resume.Config{
			CommonConfig: cfg.CommonConfig,
		}
		if err = resume.Run(resumeCfg); err != nil {
			// Only state in which we will keep the backup and new one
			return err
		}
	}

	// Delete the back up
	return os.Remove(backupDbPath)
}
