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
