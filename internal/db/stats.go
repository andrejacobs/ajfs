package db

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/path"
)

// Stats is used to calculate statistics on the database.
type Stats struct {
	DirCount  uint64 // total number of directories
	FileCount uint64 // total number of files

	TotalFileSize uint64 // total size of files all summed toghether
	AvgFileSize   uint64 // totalFileSize / fileCount

	MaxFileSize uint64 // the biggest single file size
}

// Calculate statistics on the database.
func (dbf *DatabaseFile) CalculateStats() (Stats, error) {
	dbf.panicIfNotReading()

	result := Stats{}

	err := dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		if pi.IsDir() {
			result.DirCount++
		} else if pi.IsFile() {
			result.FileCount++
			result.TotalFileSize += pi.Size
			result.MaxFileSize = max(result.MaxFileSize, pi.Size)
		}
		return nil
	})
	if err != nil {
		return result, fmt.Errorf("failed to calculate statistics for %q. %w", dbf.Path(), err)
	}

	result.AvgFileSize = result.TotalFileSize / result.FileCount
	return result, nil
}
