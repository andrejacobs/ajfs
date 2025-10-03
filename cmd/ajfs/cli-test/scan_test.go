package clitest

import (
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScan(t *testing.T) {
	root := filepath.Join(testDataPath, "scan")
	cmd := exec.Command(execPath, "scan", dbPath, root)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err)
	assert.Empty(t, out)
}
