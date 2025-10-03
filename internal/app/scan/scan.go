// Package scan provides the functionality for ajfs scan command.
package scan

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
)

// Config for the ajfs scan command.
type Config struct {
	config.CommonConfig
	config.FilterConfig

	Root string // The path to be scanned.

	ForceOverride bool // Override any existing database file.

	CalculateHashes bool        // Calculate file signature hashes.
	Algo            ajhash.Algo // Algorithm to use for calculating the hashes.

	BuildTree bool // Build and cache the tree.
	DryRun    bool // Only display files and directories that would have been stored in the database.
}

// Process the ajfs scan command.
func Run(cfg Config) error {
	if cfg.DryRun {
		return dryRun(cfg)
	}

	exists, err := file.FileExists(cfg.DbPath)
	if err != nil {
		return fmt.Errorf("failed to create the ajfs database. %w", err)
	}

	if exists {
		if cfg.ForceOverride {
			if err = os.Remove(cfg.DbPath); err != nil {
				return fmt.Errorf("failed to remove existing file %q with --force. %w", cfg.DbPath, err)
			}
		} else {
			return fmt.Errorf("failed to create the ajfs database because a file already exists at %q", cfg.DbPath)
		}
	}

	features := db.FeatureJustEntries
	if cfg.CalculateHashes {
		features |= db.FeatureHashTable
	}
	if cfg.BuildTree {
		features |= db.FeatureTree
	}

	dbf, err := db.CreateDatabase(cfg.DbPath, cfg.Root, db.FeatureFlags(features))
	if err != nil {
		return err
	}

	ctx := context.Background() // TODO: Hookup to a safe shutdown one

	// Perform the scan
	s := scanner.NewScanner()
	s.FileIncluder = cfg.FileIncluder
	s.DirIncluder = cfg.DirIncluder
	s.FileExcluder = cfg.FileExcluder
	s.DirExcluder = cfg.DirExcluder

	if err = s.Scan(ctx, dbf); err != nil {
		return err
	}

	if cfg.CalculateHashes {
		if err = calculateHashes(ctx, cfg, dbf); err != nil {
			return err
		}
	}

	//TODO: If tree, to it here

	if err = dbf.Close(); err != nil {
		return err
	}

	// TODO: Safe shutdown, cancel contex etc.
	return nil
}

func calculateHashes(ctx context.Context, cfg Config, dbf *db.DatabaseFile) error {
	// Write the initial hash table
	if err := dbf.StartHashTable(cfg.Algo); err != nil {
		return err
	}

	if err := dbf.FinishHashTable(); err != nil {
		return err
	}

	err := dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		path := filepath.Join(dbf.RootPath(), pi.Path)
		hash, _, err := file.Hash(ctx, path, cfg.Algo.Hasher(), nil)
		if err != nil {
			return fmt.Errorf("failed to calculate the hash for %q. %w", path, err)
		}
		if err = dbf.WriteHashEntry(idx, hash); err != nil {
			return fmt.Errorf("failed to write the hash for %q. %w", path, err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func dryRun(cfg Config) error {
	w := file.NewWalker()
	w.DirIncluder = cfg.DirIncluder
	w.FileIncluder = cfg.FileIncluder
	w.FileExcluder = cfg.FileExcluder
	w.DirExcluder = cfg.DirExcluder

	fn := func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		relPath, err := filepath.Rel(cfg.Root, rcvPath)
		if err != nil {
			return err
		}

		cfg.Println(relPath)

		return nil
	}

	if err := w.Walk(cfg.Root, fn); err != nil {
		return fmt.Errorf("failed to scan %q. %w", cfg.Root, err)
	}

	return nil
}
