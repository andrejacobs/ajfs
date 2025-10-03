package clitest

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanAndList(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", dbPath, root)
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
