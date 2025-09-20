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
	"os"
	"sort"
)

// Unsorted version of os.ReadDir for small optimisation. It requires less allocs if you are not concerned about sorted order.
func ReadDirUnsorted(name string) ([]os.DirEntry, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dirs, err := f.ReadDir(-1)
	//sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
	return dirs, err
}

// Sort a slice of os.DirEntry.
// This performs the same sort found in os.ReadDir.
func SortDirEntries(dirs []os.DirEntry) {
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].Name() < dirs[j].Name() })
}

// Return true if two os.DirEntry are the same without comparing .Info() (which involves fetching more data).
func IsDirEntryEqual(a os.DirEntry, b os.DirEntry) bool {
	return (a.IsDir() == b.IsDir()) &&
		(a.Type() == b.Type()) &&
		(a.Name() == b.Name())
}

// Return true if two os.DirEntry are the same. This will also make a call to .Info() which involves fetching more data and potentially result in an error.
func IsDirEntryWithInfoEqual(a os.DirEntry, b os.DirEntry) (bool, error) {
	if !IsDirEntryEqual(a, b) {
		return false, nil
	}

	ai, err := a.Info()
	if err != nil {
		return false, err
	}

	bi, err := b.Info()
	if err != nil {
		return false, err
	}

	return (ai.Size() == bi.Size()) &&
		(ai.Mode() == bi.Mode()) &&
		(ai.ModTime() == bi.ModTime()), nil
}
