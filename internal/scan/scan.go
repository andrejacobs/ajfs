// Package scan is responsible for walking a file hierarchy and writing to an ajfs database.
package scan

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/file"
)

// Scanner is used to walk a file hierarchy, perform filtering and then to write to an ajfs database.
type Scanner struct {
	DirExcluder  file.MatchPathFn // Determine which directories should not be walked
	FileExcluder file.MatchPathFn // Determine which files should not be walked
}

func NewScanner() Scanner {
	fileExcluder := file.MatchAppleDSStore(file.NeverMatch)
	return Scanner{
		DirExcluder:  file.NeverMatch,
		FileExcluder: fileExcluder,
	}
}

// Scan starts the file hierarchy traversal and will write the found path info objects to the database.
// dbf should be a newly created database [db.CreateDatabase].
func (s Scanner) Scan(dbf *db.DatabaseFile) error {
	w := file.NewWalker()
	w.FileExcluder = s.FileExcluder
	w.DirExcluder = s.DirExcluder

	fn := func(rcvPath string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}

		relPath, err := filepath.Rel(dbf.RootPath(), rcvPath)
		if err != nil {
			return err
		}

		info, err := path.InfoFromWalk(relPath, d)
		if err != nil {
			return err
		}

		return dbf.WriteEntry(&info)
	}

	if err := w.Walk(dbf.RootPath(), fn); err != nil {
		return fmt.Errorf("failed to scan %q and create ajfs database %q. %w", dbf.RootPath(), dbf.Path(), err)
	}

	return dbf.FinishEntries()
}
