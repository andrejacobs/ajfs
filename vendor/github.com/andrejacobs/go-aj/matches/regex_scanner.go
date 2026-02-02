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
	"bufio"
	"fmt"
	"io"
	"regexp"
)

// Reference on the go regex support: https://github.com/google/re2/wiki/Syntax

// RegexScanner is used to read from an io.Reader line by line and then
// tries to match the line against a set of regular expressions.
type RegexScanner struct {
	entries []regexScannerEntry
	w       io.Writer
}

// Function that will be called when a regular expression found some matches.
type RegexScannerFoundMatches func(key string, line string, lineNumber int, matches []string) error

// Result from the Process function. A map of the key to matching substrings.
// NOTE: The result will always contain the last found match for a key (meaning the map is updated on each find).
type RegexScannerResult map[string][]string

// Register a regular expression that will try and find matches when the Process function is called
// NOTE: To match case-insensitive add the prefix (?i) to the regular expression.
func (r *RegexScanner) Add(key string, expression string, foundFn RegexScannerFoundMatches) error {
	regex, err := regexp.Compile(expression)
	if err != nil {
		return fmt.Errorf("failed to compile the regular expression for the key: %q expression: %q. %w", key, expression, err)
	}

	if r.entries == nil {
		r.entries = make([]regexScannerEntry, 0, 4)
	}

	r.entries = append(r.entries, regexScannerEntry{
		key:     key,
		regex:   regex,
		foundFn: foundFn,
	})

	return nil
}

// Set the io.Writer that will be used to write any line read from the io.Reader during the Process method.
// Useful for debugging.
func (r *RegexScanner) SetOut(w io.Writer) {
	r.w = w
}

// Read line by line from the io.Reader and try and find matching regular expressions.
// The read line will be written to any writter set by SetOut method.
func (r *RegexScanner) Process(rd io.Reader) (RegexScannerResult, error) {
	scanner := bufio.NewScanner(rd)
	result := make(RegexScannerResult)

	lineNumber := 0
	for scanner.Scan() {
		line := scanner.Text()

		if r.w != nil {
			if _, err := io.WriteString(r.w, line+"\n"); err != nil {
				return result, err
			}
		}

		for _, entry := range r.entries {
			found := entry.regex.FindStringSubmatch(line)
			if found != nil {
				result[entry.key] = found
				if entry.foundFn != nil {
					err := entry.foundFn(entry.key, line, lineNumber, found)
					if err != nil {
						return result, err
					}
				}
			}
		}
		lineNumber++
	}
	if err := scanner.Err(); err != nil {
		return result, err
	}

	return result, nil
}

//-----------------------------------------------------------------------------

type regexScannerEntry struct {
	key     string
	regex   *regexp.Regexp
	foundFn RegexScannerFoundMatches
}
