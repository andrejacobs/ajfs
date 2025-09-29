// Package scan provides the functionality for ajfs scan command.
package scan

import (
	"context"
	"fmt"
	"os"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
)

// Config for the ajfs scan command.
type Config struct {
	config.CommonConfig

	Root string // The path to be scanned.

	ForceOverride bool // Override any existing database file.

	CalculateHashes bool        // Calculate file signature hashes.
	Algo            ajhash.Algo // Algorithm to use for calculating the hashes.

	BuildTree bool // Build and cache the tree.
}

// Process the ajfs scan command.
func Run(cfg Config) error {
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
	if err = s.Scan(ctx, dbf); err != nil {
		return err
	}

	//TODO: If hash, to it here
	//TODO: If tree, to it here

	if err = dbf.Close(); err != nil {
		return err
	}

	// TODO: Safe shutdown, cancel contex etc.
	return nil
}
