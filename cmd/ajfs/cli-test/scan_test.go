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

	cmd = exec.Command(execPath, "list", "--minimal", dbPath)
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

	cmd = exec.Command(execPath, "list", "--minimal", dbPath)
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

	cmd = exec.Command(execPath, "list", "--minimal", dbPath)
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
