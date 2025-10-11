package dupes

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/tree"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/human"
)

// Config for the ajfs info command.
type Config struct {
	config.CommonConfig

	Subtrees  bool
	PrintTree bool
}

// Process the ajfs info command.
func Run(cfg Config) error {
	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	if cfg.Subtrees {
		return duplicateSubtrees(cfg)
	}

	if !dbf.Features().HasHashTable() {
		return fmt.Errorf("require file signature hashes to be present in the database %q", cfg.DbPath)
	}

	grandTotalSize := uint64(0)

	totalSize := uint64(0)
	numberOfDupes := 0
	currentGroup := -1
	needFooter := false

	err = dbf.FindDuplicates(func(group, idx int, pi path.Info, hash string) error {
		if currentGroup != group {
			if pi.Size == 0 {
				needFooter = true
				return nil
			}

			if currentGroup != -1 {
				needFooter = false
				fmt.Fprintln(cfg.Stdout)
				fmt.Fprintf(cfg.Stdout, "Count: %d\n", numberOfDupes)
				fmt.Fprintf(cfg.Stdout, "Total Size: %d [%s]\n", totalSize, human.Bytes(uint64(totalSize)))
				fmt.Fprintln(cfg.Stdout, "<<<")
				fmt.Fprintln(cfg.Stdout)
			}

			fmt.Fprintln(cfg.Stdout, ">>>")
			fmt.Fprintf(cfg.Stdout, "Hash: %s\n", hash)
			fmt.Fprintf(cfg.Stdout, "Size: %d [%s]\n\n", pi.Size, human.Bytes(uint64(pi.Size)))

			currentGroup = group
			numberOfDupes = 0
			totalSize = uint64(0)
		}

		fmt.Fprintf(cfg.Stdout, "[%d]: %s\n", numberOfDupes, pi.Path)

		totalSize += pi.Size
		grandTotalSize += pi.Size
		numberOfDupes++
		needFooter = true
		return nil
	})
	if err != nil {
		return err
	}

	if needFooter {
		fmt.Fprintln(cfg.Stdout)
		fmt.Fprintf(cfg.Stdout, "Count: %d\n", numberOfDupes)
		fmt.Fprintf(cfg.Stdout, "Total Size: %d [%s]\n", totalSize, human.Bytes(uint64(totalSize)))
		fmt.Fprintln(cfg.Stdout, "<<<")
		fmt.Fprintln(cfg.Stdout)
	}

	fmt.Fprintf(cfg.Stdout, "Total size of all duplicates: %d [%s]\n", grandTotalSize, human.Bytes(grandTotalSize))
	return nil
}

func duplicateSubtrees(cfg Config) error {

	stree, err := tree.SignaturedTreeFromDatabase(cfg.DbPath)
	if err != nil {
		return err
	}

	stree.PrintDuplicateSubtrees(cfg.Stdout, cfg.PrintTree)

	return nil
}
