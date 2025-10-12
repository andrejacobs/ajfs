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

// Package tree provides the functionality for ajfs tree command.
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

	OnlyDirs bool
	Limit    int
}

// Process the ajfs info command.
func Run(cfg Config) error {

	tr, err := FromDatabase(cfg.DbPath, cfg.OnlyDirs)
	if err != nil {
		return err
	}

	if cfg.Subpath != "" {
		node := tr.Find(cfg.Subpath)
		if node == nil {
			return fmt.Errorf("failed to find the path %q in the database %q", cfg.Subpath, cfg.DbPath)
		}
		node.PrintWithLimit(cfg.Stdout, cfg.Limit)
	} else {
		tr.PrintWithLimit(cfg.Stdout, cfg.Limit)
	}

	return nil
}

// Create a tree from the path entries in an ajfs database.
func FromDatabase(dbPath string, onlyDirs bool) (itree.Tree, error) {
	dbf, err := db.OpenDatabase(dbPath)
	if err != nil {
		return itree.Tree{}, err
	}
	defer dbf.Close()

	tr := itree.New(dbf.RootPath())

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if onlyDirs && !pi.IsDir() {
			return nil
		}

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
	tr, err := FromDatabase(dbPath, false)
	if err != nil {
		return itree.SignaturedTree{}, err
	}

	stree := itree.NewSignaturedTree(tr)
	return stree, nil
}
