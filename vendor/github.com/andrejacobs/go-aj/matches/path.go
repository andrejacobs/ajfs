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

package matches

import (
	"fmt"
	"path/filepath"
)

// PathMatcher is used to determine if a file system path matches.
type PathMatcher interface {
	// Match checks if the path matches and returns true if it does
	Match(path string) (bool, error)
}

//-----------------------------------------------------------------------------
// RegexPathMatcher

// RegexPathMatcher will match a file system path against a set of regular expressions.
type RegexPathMatcher struct {
	regexList *RegexList
}

// Create a new RegexPathMatcher using the regular expression patterns.
func NewRegexPathMatcher(expressions []string) (*RegexPathMatcher, error) {
	regexList, err := NewRegexList(expressions)
	if err != nil {
		return nil, fmt.Errorf("failed to create the RegexPathMatcher. %w", err)
	}

	matcher := RegexPathMatcher{
		regexList: regexList,
	}

	return &matcher, nil
}

func (r *RegexPathMatcher) Match(path string) (bool, error) {
	matched := r.regexList.MatchesAny(path)
	return matched, nil
}

//-----------------------------------------------------------------------------
// ShellPatternPathMatcher

// ShellPatternPathMatcher will match a file system path against a set of shell patterns.
// See https://pkg.go.dev/path/filepath#Match for details.
type ShellPatternPathMatcher struct {
	patterns []string
}

// Create a new ShellPatternPathMatcher using the shell patterns.
func NewShellPatternPathMatcher(patterns []string) *ShellPatternPathMatcher {
	matcher := ShellPatternPathMatcher{
		patterns: patterns,
	}
	return &matcher
}

func (s *ShellPatternPathMatcher) Match(path string) (bool, error) {
	matched := false

	for _, pattern := range s.patterns {
		var err error
		matched, err = filepath.Match(pattern, path)
		if err != nil {
			return matched, err
		}
		if matched {
			break
		}
	}

	return matched, nil
}
