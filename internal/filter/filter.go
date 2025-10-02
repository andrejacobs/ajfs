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

// Parse the slice of strings (same as ParsePathRegex) and return the path matcher function
// to be used.
func ParsePathRegexToMatchPathFn(input []string) (file.MatchPathFn, file.MatchPathFn) {
	files, dirs := ParsePathRegex(input)
	return file.MatchRegex(files, file.MatchNever),
		file.MatchRegex(dirs, file.MatchNever)
}
