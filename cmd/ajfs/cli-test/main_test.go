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

// Package clitest is used to compile the ajfs executable and perform integration/system testing.
package clitest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	execPath = filepath.Join(binDir, execName)
	if runtime.GOOS == "windows" {
		execPath += ".exe"
	}
	var err error
	execPath, err = filepath.Abs(execPath)
	exitOnError(err, 1)

	fmt.Printf("Building %q ...\n", execPath)

	build := exec.Command("go", "build", "-o", execPath, "../main.go")
	if err := build.Run(); err != nil {
		exitOnError(fmt.Errorf("failed to build %q. %w", execPath, err), 1)
	}

	dbPath = filepath.Join(os.TempDir(), "ajfs-test.ajfs")
	_ = os.Remove(dbPath)

	testDataPath, err = filepath.Abs(testDataDir)
	exitOnError(err, 1)

	fmt.Println("Running tests....")
	result := m.Run()

	fmt.Println("Cleaning up...")
	_ = os.Remove(dbPath)
	_ = os.Remove(execPath)
	os.Exit(result)
}

const (
	execName    = "ajfs-test"
	binDir      = "../../../build/bin"
	testDataDir = "../../../internal/testdata"
)

var (
	execPath     string // Path to the test executable to be run in tests
	dbPath       string // Path to the ajfs database to be used
	testDataPath string // Path to test data set
)

func exitOnError(err error, code int) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(code)
	}
}
