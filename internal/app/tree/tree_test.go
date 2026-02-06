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
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
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
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
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
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
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

func TestOnlyDirs(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
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
		OnlyDirs: true,
	}

	err = tree.Run(config)
	require.NoError(t, err)

	absRoot, err := filepath.Abs(scanCfg.Root)
	require.NoError(t, err)

	expected := absRoot + `
├── a
│   ├── a1
│   │   ├── a1a
│   │   │   └── a1a1
│   │   └── a1b
│   └── a2
├── b
│   └── b1
│       └── b1a
└── c

11 directories, 0 files
`

	result := outBuffer.String()
	assert.Equal(t, expected, result)
	assert.Equal(t, "", errBuffer.String())
}

func TestLimit(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-testing")
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
		OnlyDirs: true,
		Limit:    2,
	}

	err = tree.Run(config)
	require.NoError(t, err)

	absRoot, err := filepath.Abs(scanCfg.Root)
	require.NoError(t, err)

	expected := absRoot + `
├── a
│   ├── a1
│   └── a2
├── b
│   └── b1
└── c

7 directories, 0 files
`

	result := outBuffer.String()
	assert.Equal(t, expected, result)
	assert.Equal(t, "", errBuffer.String())

	outBuffer.Reset()

	config.Subpath = "a"
	err = tree.Run(config)
	require.NoError(t, err)

	expected = `a
├── a1
│   ├── a1a
│   └── a1b
└── a2

5 directories, 0 files
`

	result = outBuffer.String()
	assert.Equal(t, expected, result)
	assert.Equal(t, "", errBuffer.String())

}
