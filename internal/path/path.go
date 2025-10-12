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
	// See Header() to ensure these match if any changes are made
	return fmt.Sprintf("{%x}, %v, %q, %v, %v", p.Id, p.Size, p.Path, p.Mode, p.ModTime.Format(time.RFC3339Nano))
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

//-----------------------------------------------------------------------------

// Header returns a comma seperated list of the expected columns that will be outputted by Info.String()
func Header() string {
	// See Infor.String() to ensure they match if any changes are made
	return "Id, Size, Path, Mode, Modification time"
}

// Header returns a comma seperated list of the expected columns that will be outputted for paths with a file signature hash.
func HeaderWithHash() string {
	return "Id, Hash, Size, Path, Mode, Modification time"
}
