// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package file

import (
	"io/fs"
	"os"
	"path/filepath"
)

// Get the size of the file in bytes.
func FileSize(path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// List a directory and calculate the size of all files in the directory.
// NOTE: This does not calculate recursively into subdirectories.
func CalculateDirSizeShallow(path string) (int64, []os.FileInfo, error) {
	entries, err := ReadDirUnsorted(path)
	if err != nil {
		return 0, nil, err
	}

	total := int64(0)
	infos := make([]os.FileInfo, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return 0, nil, err
		}
		total += info.Size()
		infos = append(infos, info)
	}

	return total, infos, nil
}

// Results from CalculateSize.
type CalculateSizeResult struct {
	Dirs      int    // The number of directories
	Files     int    // The number of regular files
	TotalSize uint64 // The total size in bytes of all the regular files
}

// Walk the path recusively and count the number of directories, files and the total file size.
func CalculateSize(path string) (CalculateSizeResult, error) {
	result := CalculateSizeResult{}

	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			result.Dirs++
		} else if info.Mode().IsRegular() {
			result.Files++
			result.TotalSize += uint64(info.Size())
		}
		return nil
	})

	return result, err
}
