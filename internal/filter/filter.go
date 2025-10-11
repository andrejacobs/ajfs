// Package filter provides helpers for creating the inclusion and exclusion filters used by ajfs.
package filter

import (
	"strings"

	"github.com/andrejacobs/go-aj/file"
)

// Given a slice of strings that contain regular expressions to be used
// in building the include or exclude filters this will split them
// depending on whether each regex is prefixed with "f:" for file paths
// or "d:" for directory paths. If neither prefix is found then the expression
// will be added to both file and directory.
func ParsePathRegex(input []string) ([]string, []string) {
	count := len(input)
	if count < 1 {
		return nil, nil
	}

	fileResult := make([]string, 0, count)
	dirResult := make([]string, 0, count)

	for _, s := range input {
		f, foundFile := strings.CutPrefix(s, "f:")
		if foundFile && len(f) > 0 {
			fileResult = append(fileResult, f)
			continue
		}

		d, foundDir := strings.CutPrefix(s, "d:")
		if foundDir && len(d) > 0 {
			dirResult = append(dirResult, d)
			continue
		}

		if len(s) > 0 {
			fileResult = append(fileResult, s)
			dirResult = append(dirResult, s)
		}
	}

	return fileResult, dirResult
}

// Parse the slice of strings (same as ParsePathRegex) and return the path matcher function to be used.
// include determines if the matchers will be used for includer or exclude filtering.
func ParsePathRegexToMatchPathFn(input []string, include bool) (file.MatchPathFn, file.MatchPathFn, error) {
	files, dirs := ParsePathRegex(input)

	var err error
	var fileFn file.MatchPathFn
	var dirFn file.MatchPathFn

	if len(files) > 0 {
		fileFn, err = file.MatchRegex(files, file.MatchNever)
		if err != nil {
			return nil, nil, err
		}
	} else {
		if include {
			fileFn = file.MatchAlways
		} else {
			fileFn = file.MatchNever
		}
	}

	if len(dirs) > 0 {
		dirFn, err = file.MatchRegex(dirs, file.MatchNever)
		if err != nil {
			return nil, nil, err
		}
	} else {
		if include {
			dirFn = file.MatchAlways
		} else {
			dirFn = file.MatchNever
		}
	}

	return fileFn, dirFn, nil
}
