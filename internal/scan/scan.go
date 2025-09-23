// Package scan is used to walk a file hierarchy and produce the relevant path info entities.
package scan

import (
	"fmt"
	"io/fs"
	"time"

	"github.com/andrejacobs/go-aj/file"
)

// Unique identifier of a path.
// It is simply the SHA1 hash of the path.
type PathId file.PathHash

// Describe a found path while scanning.
type PathInfo struct {
	Id   PathId // The unique identifier
	Path string // The file system path

	Size    uint64      // Size in bytes, if it is a file
	Mode    fs.FileMode // Type and permission bits
	ModTime time.Time   // Last modification time
}

// Stringer implementation.
func (p PathInfo) String() string {
	return fmt.Sprintf("{%x}, %v, %q, %v, %v", p.Id, p.Size, p.Path, p.Mode, p.ModTime)
}

// Return true if the path is a directory.
func (p *PathInfo) IsDir() bool {
	return p.Mode.IsDir()
}

// Return true if the path is a regular file
func (p *PathInfo) IsFile() bool {
	return p.Mode.IsRegular()
}

// Return true if this path info is equal to another.
func (p *PathInfo) Equals(o *PathInfo) bool {
	return (p.Id == o.Id) &&
		(p.Path == o.Path) &&
		(p.Size == o.Size) &&
		(p.Mode == o.Mode) &&
		(p.ModTime.Equal(o.ModTime))
}

// Create a path identifier.
func IdFromPath(path string) PathId {
	return PathId(file.CalculatePathHash(path))
}
