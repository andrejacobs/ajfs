// Package tosync provides the functionality for ajfs tosync command.
package tosync

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/diff"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/go-aj/human"
	"github.com/andrejacobs/go-collection/collection"
)

// Config for the ajfs diff command.
type Config struct {
	config.CommonConfig

	LhsPath string
	RhsPath string

	OnlyHashes bool

	Fn diff.CompareFn
}

// Process the ajfs diff command.
func Run(cfg Config) error {
	if cfg.Fn == nil {
		panic("expected a compare function")
	}

	return tosync(cfg)
}

func tosync(cfg Config) error {
	cfg.VerbosePrintln("Checking which files would need to be synced")
	cfg.VerbosePrintln(fmt.Sprintf("  from LHS: %q", cfg.LhsPath))
	cfg.VerbosePrintln(fmt.Sprintf("    to RHS: %q\n", cfg.RhsPath))

	lhs, err := db.OpenDatabase(cfg.LhsPath)
	if err != nil {
		return fmt.Errorf("failed to open left hand side database. %w", err)
	}
	defer lhs.Close()

	rhs, err := db.OpenDatabase(cfg.RhsPath)
	if err != nil {
		return fmt.Errorf("failed to open right hand side database. %w", err)
	}
	defer rhs.Close()

	if cfg.OnlyHashes {
		err = compareOnlyHashes(lhs, rhs, cfg.Fn)
		if err != nil {
			if err != diff.SkipAll {
				return err
			}
			return nil
		}
	} else {
		err = compare(cfg, lhs, rhs, cfg.Fn)
		if err != nil {
			if err != diff.SkipAll {
				return err
			}
			return nil
		}
	}

	return nil
}

func compare(cfg Config, lhs *db.DatabaseFile, rhs *db.DatabaseFile, fn diff.CompareFn) error {
	changedMask := ^diff.ChangedFlags(diff.ChangedModTime | diff.ChangedMode)

	count := 0
	totalSize := uint64(0)

	err := diff.CompareDatabases(lhs, rhs, true, func(d diff.Diff) error {
		// Ignore if the entry is a directory or if nothing has changed
		if d.IsDir || (d.Type == diff.TypeNothing) {
			return nil
		}

		// If only the modifaction time or mode (type and permissions) were changed then also ignore it
		// Since if you backup files to another system then the mod time and perms are bound to be different
		if (d.Type == diff.TypeChanged) && ((d.Changed & changedMask) == 0) {
			return nil
		}

		count++
		totalSize += d.Size

		return fn(d)
	})
	if err != nil {
		return err
	}

	cfg.VerbosePrintln(fmt.Sprintf("\nTotal of %d files with a size of %d bytes [%s] need to be synced", count, totalSize, human.Bytes(totalSize)))

	return nil
}

func compareOnlyHashes(lhs *db.DatabaseFile, rhs *db.DatabaseFile, fn diff.CompareFn) error {
	if !lhs.Features().HasHashTable() {
		return fmt.Errorf("left hand side database %q does not have a hash table", lhs.Path())
	}

	if !rhs.Features().HasHashTable() {
		return fmt.Errorf("right hand side database %q does not have a hash table", rhs.Path())
	}

	lhsAlgo, err := lhs.HashTableAlgo()
	if err != nil {
		return fmt.Errorf("failed to get left hand side's hashing algorith. %w", err)
	}

	rhsAlgo, err := rhs.HashTableAlgo()
	if err != nil {
		return fmt.Errorf("failed to get right hand side's hashing algorith. %w", err)
	}

	if lhsAlgo != rhsAlgo {
		return fmt.Errorf("can't compare the two databases because left uses %q and right uses %q", lhsAlgo, rhsAlgo)
	}

	lhsHashes, err := lhs.BuildHashStrToIndexMap()
	if err != nil {
		return fmt.Errorf("failed to get the left hand side's hash table. %w", err)
	}

	rhsHashes, err := rhs.BuildHashStrToIndexMap()
	if err != nil {
		return fmt.Errorf("failed to get the right hand side's hash table. %w", err)
	}

	// What exists only on the LHS (removed from RHS)
	lhsOnly := collection.MapDifference(lhsHashes, rhsHashes)

	for _, v := range lhsOnly {
		pi, err := lhs.ReadEntryAtIndex(v)
		if err != nil {
			return fmt.Errorf("failed to read left hand side entry with index %d. %w", v, err)
		}

		err = fn(diff.Diff{
			Type:  diff.TypeLeftOnly,
			Id:    pi.Id,
			Path:  pi.Path,
			IsDir: pi.IsDir(),
		})
		if err != nil {
			return err
		}
	}

	return nil
}
