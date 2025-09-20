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
	"fmt"
	"io/fs"
	"path/filepath"
)

// Walker is used to walk a file hierarchy.
type Walker struct {
	DirExcluder  MatchPathFn // Determine which directories should not be walked
	FileExcluder MatchPathFn // Determine which files should not be walked
}

// Create a new Walker.
//
// By default all files and directories found will be walked and not be excluded.
func NewWalker() *Walker {
	return &Walker{
		DirExcluder:  NeverMatch,
		FileExcluder: NeverMatch,
	}
}

// Set the excluder used to determine which directories should not be walked.
func (w *Walker) SetDirExcluder(excluder MatchPathFn) *Walker {
	w.DirExcluder = excluder
	return w
}

// Set the excluder used to determine which files should not be walked.
func (w *Walker) SetFileExcluder(excluder MatchPathFn) *Walker {
	w.FileExcluder = excluder
	return w
}

// Walk walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
//
// All errors that arise visiting files and directories are filtered by fn:
// see the [fs.WalkDirFunc] documentation for details.
//
// The files are walked in lexical order, which makes the output deterministic
// but requires Walk to read an entire directory into memory before proceeding
// to walk that directory.
//
// Walk does not follow symbolic links.
//
// Walk calls fn with paths that use the separator character appropriate
// for the operating system.
//
// Walk uses [fs.WalkDir] for implementation.
//
// For each directory that is found, the DirExcluder will be called to determine
// if the path should not be walked.
//
// For each file that is found, the FileExcluder will be called to determine
// if the path should not be walked.
//
// The root path will be expanded using [file.ExpandPath] if needed.
func (w *Walker) Walk(root string, fn fs.WalkDirFunc) error {
	expandedRoot, err := ExpandPath(root)
	if err != nil {
		return fmt.Errorf("failed to expand the path %q. %w", root, err)
	}

	rErr := filepath.WalkDir(expandedRoot, func(path string, d fs.DirEntry, rcvErr error) error {
		// Did we receive an error?
		if rcvErr != nil {
			fnErr := fn(path, d, rcvErr)
			return fnErr
		}

		// Does the directory need to be excluded?
		exclude, err := w.DirExcluder(path, d)
		if err != nil {
			return err
		}
		if exclude {
			return fs.SkipDir
		}

		// Does the file need to be excluded?
		exclude, err = w.FileExcluder(path, d)
		if err != nil {
			return err
		}
		if exclude {
			return nil
		}

		// fmt.Printf("walker>>> %q\n", path)
		fnErr := fn(path, d, nil)
		return fnErr
	})

	return rErr
}

//-----------------------------------------------------------------------------
// Matchers

// MatchPathFn determines if the path matches a criteria and if so returns true.
type MatchPathFn func(path string, d fs.DirEntry) (bool, error)

// MatchPathMiddleware specifies the function signature for a wrapping MatchPathFn into a call chain.
// Similarly to how http middleware works in popular frameworks.
type MatchPathMiddleware func(next MatchPathFn) MatchPathFn

// NeverMatch is a MatchPathFn that will always return false.
func NeverMatch(path string, d fs.DirEntry) (bool, error) {
	return false, nil
}

//-----------------------------------------------------------------------------
// Matcher middleware

// MatchAppleDSStore middleware will match Apple .DS_Store files.
func MatchAppleDSStore(next MatchPathFn) MatchPathFn {
	return func(path string, d fs.DirEntry) (bool, error) {
		if !d.IsDir() && d.Name() == ".DS_Store" {
			return true, nil
		}
		return next(path, d)
	}
}
