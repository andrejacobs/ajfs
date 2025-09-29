// Package list provides the functionality for ajfs list command.
package list

import (
	"path/filepath"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
)

// Config for the ajfs list command.
type Config struct {
	config.CommonConfig

	DisplayFullPaths bool // If true then each path entry will be prefixed with the root path of the database.
}

// Process the ajfs list command.
func Run(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if cfg.DisplayFullPaths {
			pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
		}
		cfg.Println(pi)
		return nil
	})

	return err
}
