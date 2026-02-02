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

package random

// Provide utility functions for creating random file paths. Mainly used in unit-testing.

import (
	"fmt"
	"os"
	"path"
	"strings"
)

// Generate a path consisting of random depth (subdirectories) between min and max
// minDirs, maxDirs: random range between the minimum  and maximum amount of subdirectories to create
// minNameLen, maxNameLen: random range of length of characters used to generate each random subdirectory's name.
// The function will always return the base + range(min, max) paths.
func Path(base string, minDirs int, maxDirs int, minNameLen int, maxNameLen int) string {
	sb := strings.Builder{}
	count := Int(minDirs, maxDirs)
	minNameLen = max(1, minNameLen)
	for depth := 0; depth < count; depth++ {
		sb.WriteString(String(Int(minNameLen, maxNameLen)))
		if depth < (count - 1) {
			sb.WriteRune(os.PathSeparator)
		}
	}
	return path.Join(base, sb.String())
}

// Generate a slice of random paths
// count: is the number of random paths to create and return
func Paths(base string, count int, min int, max int, minNameLen int, maxNameLen int) []string {
	paths := make([]string, 0, count)
	for i := 0; i < count; i++ {
		paths = append(paths, Path(base, min, max, minNameLen, maxNameLen))
	}
	return paths
}

// Generate random files inside the specified directory
// Files will be created using data copied from the crypto random generator.
// dir: is the parent directory
// minFile: minimum number of files to create
// maxFile: maximum number of files to create
// minSize: the minimum size in bytes of a file.
// maxSize: the maximum size in bytes of a file.
// maxTotalSize: the maximum number of bytes to be used for all files being created.
// Return the total number of bytes written.
func CreateFiles(dir string,
	minFiles int, maxFiles int,
	minSize uint64, maxSize uint64,
	maxTotalSize uint64) (uint64, error) {

	currentTotalSize := uint64(0)

	for i := 0; i < Int(minFiles, maxFiles); i++ {
		path := path.Join(dir, fmt.Sprintf("%s-%d", String(Int(1, 16)), i))
		if currentTotalSize < maxTotalSize {
			amount := min(int64(Int(0, int(maxSize))), int64(maxTotalSize-currentTotalSize))
			err := CreateFile(path, amount)
			if err != nil {
				return currentTotalSize, err
			}
			currentTotalSize += uint64(amount)
			if currentTotalSize >= maxTotalSize {
				break
			}
		}
	}

	return currentTotalSize, nil
}
