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
