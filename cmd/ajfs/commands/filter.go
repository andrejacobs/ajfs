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

package commands

import (
	"fmt"
	"runtime"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/filter"
	"github.com/andrejacobs/go-aj/file"
	"github.com/spf13/cobra"
)

var (
	includePathRegex []string // Regexes for path inclusion filtering
	excludePathRegex []string // Regexes for path exclusion filtering
)

// Add the path filtering flags to the cobra command.
func addPathFilteringFlags(c *cobra.Command) {
	c.Flags().StringArrayVarP(&includePathRegex, "include", "i", nil, "Include path regex filter")
	c.Flags().StringArrayVarP(&excludePathRegex, "exclude", "e", nil, "Exclude path regex filter")
}

// Parse the include path regexes into file and dir path matchers.
func parseIncludePathRegex() (file.MatchPathFn, file.MatchPathFn, error) {
	return filter.ParsePathRegexToMatchPathFn(includePathRegex, true)
}

// Parse the exclude path regexes into file and dir path matchers.
func parseExcludePathRegex() (file.MatchPathFn, file.MatchPathFn, error) {
	return filter.ParsePathRegexToMatchPathFn(excludePathRegex, false)
}

// Parse the filtering config that can be used by commands.
func parseFilterConfig() (*config.FilterConfig, error) {
	result := &config.FilterConfig{}

	incF, incD, err := parseIncludePathRegex()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the include filtering flags. %w", err)
	}

	result.FileIncluder = incF
	result.DirIncluder = incD

	exclF, exclD, err := parseExcludePathRegex()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the exclude filtering flags. %w", err)
	}

	result.FileExcluder = file.MatchAppleDSStore(exclF)
	result.DirExcluder = exclD

	if runtime.GOOS == "darwin" {
		result.FileExcluder = file.MatchAppleProtected(result.FileExcluder)
		result.DirExcluder = file.MatchAppleProtected(result.DirExcluder)
	}

	return result, nil
}
