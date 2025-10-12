// Package resume provides the functionality for ajfs resume command.
package resume

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/file"
	"github.com/andrejacobs/go-aj/human"
	"github.com/schollz/progressbar/v3"
)

// Config for the ajfs scan command.
type Config struct {
	config.CommonConfig
}

// Process the ajfs scan command.
func Run(cfg Config) error {
	cfg.ProgressPrintln(fmt.Sprintf("Resuming database file at %q", cfg.DbPath))
	dbf, err := db.ResumeDatabase(cfg.DbPath)
	if err != nil {
		return err
	}

	if !dbf.Features().HasHashTable() {
		cfg.VerbosePrintln("Nothing to resume")
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Hook into listening for the SIGINT (Ctrl+C) and SIGTERM signals
	signalCh := make(chan os.Signal, 1)
	interruptedCh := make(chan bool, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		rcv := <-signalCh
		cfg.VerbosePrintln(fmt.Sprintf("\nReceived signal: %s", rcv))

		cancel()

		interruptedCh <- true
	}()

	if err = resumeCalculatingHashes(ctx, cfg, dbf); err != nil {
		if !errors.Is(err, context.Canceled) {
			return err
		}
	}

	select {
	case <-interruptedCh:
		cfg.VerbosePrintln("App was interrupted.")
	default:
	}

	if err = dbf.Close(); err != nil {
		return err
	}

	cfg.VerbosePrintln("Done!")
	return nil
}

func resumeCalculatingHashes(ctx context.Context, cfg Config, dbf *db.DatabaseFile) error {
	algo, err := dbf.HashTableAlgo()
	if err != nil {
		return err
	}

	cfg.VerbosePrintln("Calculating file signature hashes ...")
	cfg.VerbosePrintln(fmt.Sprintf("  Algorithm: %s", algo))

	var progress *progressbar.ProgressBar
	count := 0
	totalCount := 0

	if cfg.Progress {
		cfg.ProgressPrintln("Calculating progress information ...")
		stats, err := dbf.CalculateStats()
		if err != nil {
			return err
		}

		totalCount = int(stats.FileCount)

		todoSize := uint64(0)
		todoCount := 0
		err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
			todoSize += pi.Size
			todoCount++
			return nil
		})
		if err != nil {
			return err
		}

		cfg.VerbosePrintln(fmt.Sprintf("Still need to process %d files [%s]", todoCount, human.Bytes(todoSize)))

		progress = progressbar.DefaultBytes(int64(stats.TotalFileSize))
		progress.Set64(int64(stats.TotalFileSize - todoSize))
		count = totalCount - todoCount
	}

	err = dbf.EntriesNeedHashing(func(idx int, pi path.Info) error {
		if progress != nil {
			progress.Describe(fmt.Sprintf("[%d/%d]", count+1, totalCount))
		} else {
			cfg.VerbosePrintln(fmt.Sprintf("Hashing %q", pi.Path))
		}

		path := filepath.Join(dbf.RootPath(), pi.Path)
		hash, _, err := file.Hash(ctx, path, algo.Hasher(), progress)
		if err != nil {
			return fmt.Errorf("failed to calculate the hash for %q. %w", path, err)
		}
		if err = dbf.WriteHashEntry(idx, hash); err != nil {
			return fmt.Errorf("failed to write the hash for %q. %w", path, err)
		}

		count++
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
