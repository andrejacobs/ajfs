// Package list provides the functionality for ajfs list command.
package list

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
)

// Config for the ajfs list command.
type Config struct {
	config.CommonConfig

	DisplayFullPaths bool // If true then each path entry will be prefixed with the root path of the database.
	DisplayHashes    bool // Display file signature hashes if available.
	DisplayMinimal   bool // Display only the paths.
}

// Process the ajfs list command.
func Run(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	if cfg.DisplayMinimal {
		if err = displayOnlyMinimal(cfg, dbf); err != nil {
			return err
		}
		return nil
	}

	if cfg.CommonConfig.Verbose {
		if dbf.Features().HasHashTable() {
			cfg.Println(path.HeaderWithHash())
		} else {
			cfg.Println(path.Header())
		}
	}

	if cfg.DisplayHashes && dbf.Features().HasHashTable() {
		err = dbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
			if cfg.DisplayFullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			hashStr := hex.EncodeToString(hash)
			cfg.Println(fmt.Sprintf("{%x}, %q, %v, %q, %v, %v", pi.Id, hashStr, pi.Size, pi.Path, pi.Mode, pi.ModTime.Format(time.RFC3339Nano)))
			return nil
		})
		return err
	} else {
		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			if cfg.DisplayFullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			cfg.Println(pi)
			return nil
		})
		return err
	}
}

func displayOnlyMinimal(cfg Config, dbf *db.DatabaseFile) error {
	err := dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if cfg.DisplayFullPaths {
			pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
		}

		cfg.Println(pi.Path)
		return nil
	})

	return err
}
