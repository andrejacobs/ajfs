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

// file provide simple utilities for working with the file system
package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Check if the path exists.
// If the path exists then (true, nil) is returned.
// If the path does not exist then (false, nil) is returned.
// If an error occurred while trying to check if the path exists then (false, err) is returned.
func PathExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Check if the path exists and is a directory.
// If the path does not exists then (false, nil) will be returned.
// If the path exists but is not a directory then (false, nil) will be returned.
// An error is only returned if an error occurred while checking if the path exists.
func DirExists(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil {
		return info.IsDir(), nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Check if the path exists and is a regular file.
// If the path does not exists then (false, nil) will be returned.
// If the path exists but is not a regular file then (false, nil) will be returned.
// An error is only returned if an error occurred while checking if the path exists.
func FileExists(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil {
		return info.Mode().IsRegular(), nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}

// Recursively find all files in dir that matches the specified extension.
// NOTE: ext must include the dot (period) e.g.  .txt.
func GlobExt(dir string, ext string) ([]string, error) {
	files := []string{}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if filepath.Ext(path) == ext {
			files = append(files, path)
		}
		return nil
	})

	return files, err
}

// Convert the slice of paths to the absolute paths and optionally verify the paths exists.
func AbsPaths(paths []string, checkExists bool) ([]string, error) {
	absPaths := []string{}
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return nil, fmt.Errorf("failed to find the absolute path for %q. error: %w", p, err)
		}
		absPaths = append(absPaths, absPath)
	}

	// Ensure paths exist
	if checkExists {
		for _, p := range absPaths {
			exists, err := PathExists(p)
			if err != nil {
				return nil, fmt.Errorf("invalid path %q. error: %w", p, err)
			}
			if !exists {
				return nil, fmt.Errorf("the path %q does not exist", p)
			}
		}
	}

	return absPaths, nil
}

// Replace the path's file extension with a new one.
func ReplaceExt(path string, newExt string) string {
	ext := filepath.Ext(path)
	if len(ext) < 1 {
		return path + newExt
	}

	withoutExt := path[:len(path)-len(ext)]
	return withoutExt + newExt
}

// Delete the path if it exists and only return an error if something went wrong other
// than the fact that the path didn't exist.
func RemoveIfExists(path string) error {
	if err := os.Remove(path); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}
	return nil
}

// Expand the path to include the user's home directory if the path starts with the suffix ~
func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}
	return path, nil
}
