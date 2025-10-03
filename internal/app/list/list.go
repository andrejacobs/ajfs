// Package list provides the functionality for ajfs list command.
package list

import (
	"encoding/hex"
	"fmt"
	"path/filepath"

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

	var hashTable db.HashTable

	if cfg.DisplayHashes && dbf.Features().HasHashTable() {
		hashTable, err = dbf.ReadHashTable()
		if err != nil {
			return err
		}
	}

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if cfg.DisplayFullPaths {
			pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
		}

		if hashTable != nil {
			hash, ok := hashTable[idx]
			var hashStr string
			if ok {
				hashStr = hex.EncodeToString(hash)
			}
			cfg.Println(fmt.Sprintf("{%x}, %v, %q, %v, %v, %s", pi.Id, pi.Size, pi.Path, pi.Mode, pi.ModTime, hashStr))
		} else {
			cfg.Println(pi)
		}
		return nil
	})

	return err
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
