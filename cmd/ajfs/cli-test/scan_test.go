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

package clitest

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanAndList(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--force", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)

	cmd = exec.Command(execPath, "list", dbPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)

	expected, err := expectedScanListing()
	require.NoError(t, err)

	result, err := splitInput(out)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func TestScanIncludeFileFiltering(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--force", "-i", "f:blank\\.txt$", "-i", "f:3\\.txt$", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)

	cmd = exec.Command(execPath, "list", dbPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)

	temp, err := expectedScanListing()
	require.NoError(t, err)

	expected := make([]string, 0, len(temp))
	for _, s := range temp {
		if strings.HasSuffix(s, ".txt") {
			if strings.HasSuffix(s, "blank.txt") || strings.HasSuffix(s, "3.txt") {
				expected = append(expected, s)
			}
		} else {
			expected = append(expected, s)
		}
	}

	result, err := splitInput(out)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func TestScanIncludeDirFiltering(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--force", "-i", "d:b", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)

	cmd = exec.Command(execPath, "list", dbPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)

	temp, err := expectedScanListing()
	require.NoError(t, err)

	expected := make([]string, 0, len(temp))
	for _, s := range temp {
		if strings.HasPrefix(s, "b") || s == "." || s == "1.txt" {
			expected = append(expected, s)
		}
	}

	result, err := splitInput(out)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func TestScanExcludeFileFiltering(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--force", "-e", "f:blank\\.txt$", "-e", "f:same-as-", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)

	cmd = exec.Command(execPath, "list", dbPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)

	temp, err := expectedScanListing()
	require.NoError(t, err)

	expected := make([]string, 0, len(temp))
	for _, s := range temp {
		if !strings.Contains(s, "blank.txt") && !strings.Contains(s, "same-as-") {
			expected = append(expected, s)
		}
	}

	result, err := splitInput(out)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func TestScanExcludeDirFiltering(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--force", "-e", "d:a", "-e", "d:b", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)

	cmd = exec.Command(execPath, "list", dbPath)
	out, err = cmd.CombinedOutput()
	require.NoError(t, err)

	temp, err := expectedScanListing()
	require.NoError(t, err)

	expected := make([]string, 0, len(temp))
	for _, s := range temp {
		if strings.HasPrefix(s, "c") || s == "." || s == "1.txt" || s == "blank.txt" {
			expected = append(expected, s)
		}
	}

	result, err := splitInput(out)
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func TestScanDryRun(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", "--dry-run", root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)

	result, err := splitInput(out)
	require.NoError(t, err)

	expected, err := expectedScanListing()
	require.NoError(t, err)

	assert.ElementsMatch(t, expected, result)
}

func expectedScanListing() ([]string, error) {
	f, err := os.Open(filepath.Join(testDataPath, "expected/scan.txt"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return splitLines(f)
}

func splitLines(r io.Reader) ([]string, error) {
	result := make([]string, 0, 32)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func splitInput(input []byte) ([]string, error) {
	return splitLines(bytes.NewReader(input))
}
