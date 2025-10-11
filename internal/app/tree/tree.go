package tree

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	itree "github.com/andrejacobs/ajfs/internal/tree"
)

// Config for the ajfs tree command.
type Config struct {
	config.CommonConfig
	Subpath string
}

// Process the ajfs info command.
func Run(cfg Config) error {

	tr, err := FromDatabase(cfg.DbPath)
	if err != nil {
		return err
	}

	if cfg.Subpath != "" {
		node := tr.Find(cfg.Subpath)
		if node == nil {
			return fmt.Errorf("failed to find the path %q in the database %q", cfg.Subpath, cfg.DbPath)
		}
		node.Print(cfg.Stdout)
	} else {
		tr.Print(cfg.Stdout)
	}

	return nil
}

// Create a tree from the path entries in an ajfs database.
func FromDatabase(dbPath string) (itree.Tree, error) {
	dbf, err := db.OpenDatabase(dbPath)
	if err != nil {
		return itree.Tree{}, err
	}
	defer dbf.Close()

	tr := itree.New(dbf.RootPath())

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		node := tr.Insert(pi)
		if node == nil {
			return fmt.Errorf("failed to insert new node into the tree (index = %d, path = %q)", idx, pi.Path)
		}
		return nil
	})
	if err != nil {
		return itree.Tree{}, err
	}

	return tr, nil
}

// Create a signatured tree from the path entries in an ajfs database.
func SignaturedTreeFromDatabase(dbPath string) (itree.SignaturedTree, error) {
	tr, err := FromDatabase(dbPath)
	if err != nil {
		return itree.SignaturedTree{}, err
	}

	stree := itree.NewSignaturedTree(tr)
	return stree, nil
}
