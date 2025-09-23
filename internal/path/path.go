// Package path is used to represent path info objects.
package path

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/andrejacobs/go-aj/file"
)

// Unique identifier of a path.
// It is simply the SHA1 hash of the path.
type Id file.PathHash

// Describe a found path while scanning.
type Info struct {
	Id   Id     // The unique identifier
	Path string // The file system path

	Size    uint64      // Size in bytes, if it is a file
	Mode    fs.FileMode // Type and permission bits
	ModTime time.Time   // Last modification time
}

// Stringer implementation.
func (p Info) String() string {
	return fmt.Sprintf("{%x}, %v, %q, %v, %v", p.Id, p.Size, p.Path, p.Mode, p.ModTime)
}

// Return true if the path is a directory.
func (p *Info) IsDir() bool {
	return p.Mode.IsDir()
}

// Return true if the path is a regular file
func (p *Info) IsFile() bool {
	return p.Mode.IsRegular()
}

// Return true if this path info is equal to another.
func (p *Info) Equals(o *Info) bool {
	return (p.Id == o.Id) &&
		(p.Path == o.Path) &&
		(p.Size == o.Size) &&
		(p.Mode == o.Mode) &&
		(p.ModTime.Equal(o.ModTime))
}

// Create a path identifier.
func IdFromPath(path string) Id {
	return Id(file.CalculatePathHash(path))
}

// Create the path info from the results of a file system walk [filepath.WalkDir] or [file.Walker].
func InfoFromWalk(path string, entry fs.DirEntry) (Info, error) {
	fileInfo, err := entry.Info()
	if err != nil {
		return Info{}, fmt.Errorf("failed to create the path.Info object from path %q. %w", path, err)
	}

	return Info{
		Id:      IdFromPath(path),
		Path:    path,
		Size:    uint64(fileInfo.Size()),
		Mode:    fileInfo.Mode(),
		ModTime: fileInfo.ModTime(),
	}, nil
}
