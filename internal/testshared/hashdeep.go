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

package testshared

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// A single entry in the hashdeep file.
type HashDeepEntry struct {
	FileSize int
	Hash     string
	Path     string
}

// Parse a hashdeep file.
func ReadHashDeepFile(path string) ([]HashDeepEntry, error) {
	result := make([]HashDeepEntry, 0, 32)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "%%") {
			continue
		}
		if strings.HasPrefix(text, "##") {
			continue
		}

		entry := HashDeepEntry{}

		parts := strings.Split(text, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("failed to parse the line: %s", text)
		}
		entry.FileSize, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		entry.Hash = parts[1]
		entry.Path = strings.TrimPrefix(parts[2], "./")
		result = append(result, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
