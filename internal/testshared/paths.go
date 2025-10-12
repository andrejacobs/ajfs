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

package testshared

import (
	"io/fs"
	"path/filepath"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/scanner"
	"github.com/andrejacobs/go-aj/file"
)

// Walk a file hierarchy and produce the expected path info entries.
func ExpectedPaths(root string, filterCfg *config.FilterConfig) ([]path.Info, error) {
	w := file.NewWalker()

	if filterCfg != nil {
		w.DirIncluder = filterCfg.DirIncluder
		w.FileIncluder = filterCfg.FileIncluder
		w.DirExcluder = filterCfg.DirExcluder
		w.FileExcluder = filterCfg.FileExcluder
	} else {
		w.FileExcluder = scanner.DefaultFileExcluder()
	}

	result := make([]path.Info, 0, 32)

	err := w.Walk(root, func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		relPath, err := filepath.Rel(root, rcvPath)
		if err != nil {
			return err
		}

		expInfo, err := path.InfoFromWalk(relPath, d)
		if err != nil {
			return err
		}

		result = append(result, expInfo)

		return nil
	})

	return result, err
}

// Read all the stored path info entries from a database.
func DatabasePaths(dbPath string) ([]path.Info, error) {
	dbf, err := db.OpenDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	defer dbf.Close()

	result := make([]path.Info, 0, 32)

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		result = append(result, pi)
		return nil
	})

	return result, err
}
