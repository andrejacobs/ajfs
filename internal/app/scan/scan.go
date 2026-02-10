// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package scan provides the functionality for ajfs scan command.
package scan

import (
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
	"github.com/schollz/progressbar/v3"
)

// Config for the ajfs scan command.
type Config struct {
	config.CommonConfig
	config.FilterConfig

	Root string // The path to be scanned.

	ForceOverride bool // Override any existing database file.

	CalculateHashes bool        // Calculate file signature hashes.
	Algo            ajhash.Algo // Algorithm to use for calculating the hashes.
	hashFn          hashFn      // Hashing function

	DryRun   bool // Only display files and directories that would have been stored in the database.
	InitOnly bool // The initial database will be created without long running processes (hashing).

	simulateHashingError bool // Cause an error while calculating file signature hashes.
}

// The hashing function to be used for calculating file signature hashes.
type hashFn func(ctx context.Context, path string, hasher hash.Hash, w io.Writer) ([]byte, uint64, error)

// Process the ajfs scan command.
func Run(cfg Config) error {
	if cfg.hashFn == nil {
		cfg.hashFn = file.Hash
	}

	if cfg.DryRun {
		return dryRun(cfg)
	}

	cfg.VerbosePrintln(fmt.Sprintf("Scanning root path %q", cfg.Root))

	exists, err := file.FileExists(cfg.DbPath)
	if err != nil {
		return fmt.Errorf("failed to create the ajfs database. %w", err)
	}

	if exists {
		if cfg.ForceOverride {
			cfg.VerbosePrintln(fmt.Sprintf("Removing database file %q because --force is specified", cfg.DbPath))
			if err = os.Remove(cfg.DbPath); err != nil {
				return fmt.Errorf("failed to remove existing file %q with --force. %w", cfg.DbPath, err)
			}
		} else {
			return fmt.Errorf("failed to create the ajfs database because a file already exists at %q", cfg.DbPath)
		}
	}

	features := db.FeatureFlags(db.FeatureJustEntries)
	if cfg.CalculateHashes {
		features |= db.FeatureHashTable
		cfg.VerbosePrintln("Will be creating a hash table")
	}

	cfg.VerbosePrintln(fmt.Sprintf("Creating database file at %q", cfg.DbPath))
	dbf, err := db.CreateDatabase(cfg.DbPath, cfg.Root, db.FeatureFlags(features))
	if err != nil {
		return err
	}

	defer func() {
		if err := dbf.Close(); err != nil {
			fmt.Fprintln(cfg.Stderr, err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Hook into listening for the SIGINT (Ctrl+C) and SIGTERM signals
	signalCh := make(chan os.Signal, 1)
	interruptedCh := make(chan bool, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	safeToShutdown := false

	go func() {
		rcv := <-signalCh
		cfg.VerbosePrintln(fmt.Sprintf("\nReceived signal: %s", rcv))

		cancel()

		interruptedCh <- true
	}()

	// Perform the scan
	s := scanner.NewScanner()
	s.FileIncluder = cfg.FileIncluder
	s.DirIncluder = cfg.DirIncluder
	s.FileExcluder = cfg.FileExcluder
	s.DirExcluder = cfg.DirExcluder

	cfg.ProgressPrintln("Scanning ...")
	if err = s.Scan(ctx, dbf); err != nil {
		if !errors.Is(err, context.Canceled) {
			return err
		}
	} else {
		safeToShutdown = true
	}

	if cfg.CalculateHashes && (ctx.Err() == nil) {
		if err = calculateHashes(ctx, cfg, dbf); err != nil {
			if !errors.Is(err, context.Canceled) {
				return err
			}
		}
	}

	select {
	case <-interruptedCh:
		if !safeToShutdown {
			cfg.Errorln("\nApp was interrupted and the ajfs database file is incomplete. File will be deleted.")
			if err = dbf.Interrupted(); err != nil {
				return err
			}
			return nil
		}
		cfg.VerbosePrintln("App was interrupted, however the ajfs database file is still valid.")
	default:
	}

	cfg.VerbosePrintln("Done!")

	return nil
}

func calculateHashes(ctx context.Context, cfg Config, dbf *db.DatabaseFile) error {
	cfg.VerbosePrintln("Calculating file signature hashes ...")
	cfg.VerbosePrintln(fmt.Sprintf("  Algorithm: %s", cfg.Algo))

	// Write the initial hash table
	cfg.VerbosePrintln("Creating initial hash table ...")
	if err := dbf.StartHashTable(cfg.Algo); err != nil {
		return err
	}

	if err := dbf.FinishHashTable(); err != nil {
		return err
	}

	if cfg.InitOnly {
		cfg.VerbosePrintln("Skipping calculation because of InitOnly")
		return nil
	}

	var progress *progressbar.ProgressBar
	count := 0
	totalCount := uint64(0)

	if cfg.Progress {
		cfg.ProgressPrintln("Calculating progress information ...")
		stats, err := dbf.CalculateStats()
		if err != nil {
			return err
		}

		progress = progressbar.DefaultBytes(int64(stats.TotalFileSize)) //nolint:gosec // disable G115
		totalCount = stats.FileCount
	}

	if cfg.simulateHashingError {
		return fmt.Errorf("simulating an error while calculating file signature hashes")
	}

	err := dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {

		if progress != nil {
			progress.Describe(fmt.Sprintf("[%d/%d]", count+1, totalCount))
		} else {
			cfg.VerbosePrintln(fmt.Sprintf("Hashing %q", pi.Path))
		}

		path := filepath.Join(dbf.RootPath(), pi.Path)
		hash, _, err := cfg.hashFn(ctx, path, cfg.Algo.Hasher(), progress)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}

			// Continue hashing
			fmt.Fprintf(cfg.Stderr, "failed to calculate the hash for %q. %v\n", path, err)
		} else {
			if err = dbf.WriteHashEntry(idx, hash); err != nil {
				return fmt.Errorf("failed to write the hash for %q. %w", path, err)
			}
		}

		count++
		return nil
	})

	if err != nil {
		if progress != nil {
			_ = progress.Exit()
		}
		return err
	}

	return nil
}

func dryRun(cfg Config) error {
	cfg.VerbosePrintln(fmt.Sprintf("[DRY-RUN] Scan root path %q", cfg.Root))

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
