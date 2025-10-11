package tree_test

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/app/tree"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/scan",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	config := tree.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
	}

	err = tree.Run(config)
	require.NoError(t, err)

	absRoot, err := filepath.Abs(scanCfg.Root)
	require.NoError(t, err)

	expected := absRoot + `
├── 1.txt
├── a
│   ├── 2.txt
│   ├── 3.txt
│   ├── a1
│   │   ├── a1a
│   │   │   └── a1a1
│   │   │       ├── 1.txt
│   │   │       ├── 4.txt
│   │   │       └── blank.txt
│   │   └── a1b
│   │       └── 5.txt
│   └── a2
│       ├── 6.txt
│       └── same-as-1.txt
├── b
│   └── b1
│       └── b1a
│           ├── 1.txt
│           ├── 7.txt
│           ├── blank.txt
│           └── same-as-1.txt
├── blank.txt
└── c
    └── c.txt

11 directories, 15 files
`

	result := outBuffer.String()
	assert.Equal(t, expected, result)
	assert.Equal(t, "", errBuffer.String())

	// Compare against tree CLI if it exists
	cmd := exec.Command("command", "-v", "tree")
	err = cmd.Run()
	if err == nil {
		t.Log("comparing against installed tree CLI")
		cmd = exec.Command("tree", absRoot)
		out, err := cmd.CombinedOutput()
		require.NoError(t, err)

		outStr := strings.ReplaceAll(string(out), "\u00a0", " ")
		assert.Equal(t, outStr, result)
	}
}

func TestSubpath(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/scan",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	config := tree.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
		Subpath: "a/a1",
	}

	err = tree.Run(config)
	require.NoError(t, err)

	expected := `a1
├── a1a
│   └── a1a1
│       ├── 1.txt
│       ├── 4.txt
│       └── blank.txt
└── a1b
    └── 5.txt

4 directories, 4 files
`

	result := outBuffer.String()
	assert.Equal(t, expected, result)
	assert.Equal(t, "", errBuffer.String())
}

func TestSubpathDoesNotExist(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/scan",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	config := tree.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
		Subpath: "the/quick/brown/fox",
	}

	err = tree.Run(config)
	assert.ErrorContains(t, err, "failed to find the path")
}
