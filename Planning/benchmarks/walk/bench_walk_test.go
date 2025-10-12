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

package walk_test

// Run with: BENCH_DIR="./path/to/dataset" go test -run=^$ -bench=. -benchtime=3x -benchmem
// Optionally BENCH_OUT="~/temp/output"

// Reference:
// -benchtime=10x Limits iterations to 10 times
// -benchtime=10s Limits iteration time to 10 seconds
// -race for race detector

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/go-aj/file"
)

const (
	benchDirEnv = "BENCH_DIR"
	outDirEnv   = "BENCH_OUT"
)

var benchDir string
var outDir string

func setup() error {
	benchDir = os.Getenv(benchDirEnv)
	if benchDir == "" {
		return fmt.Errorf("environment variable %q needs to specify that directory to be used in the benchmark", benchDirEnv)
	}

	outDir = os.Getenv(outDirEnv)
	if outDir == "" {
		outDir = os.TempDir()
	}
	var err error
	outDir, err = file.ExpandPath(outDir)
	if err != nil {
		return err
	}

	return nil
}

func Benchmark(b *testing.B) {
	if err := setup(); err != nil {
		b.Fatal(err)
	}

	fmt.Printf("Using %q as the dataset\n", benchDir)
	fmt.Printf("Using %q as the output\n", outDir)

	b.Run("vanilla", func(b *testing.B) {
		for b.Loop() {
			if err := vanillaWalk(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("one-worker", func(b *testing.B) {
		for b.Loop() {
			if err := oneWorkerWalk(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func vanillaWalk() error {
	outPath := filepath.Join(outDir, "walk-vanilla.txt")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create the output file %q. %w", outPath, err)
	}
	defer f.Close()

	var fn fs.WalkDirFunc = func(path string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}
		fmt.Fprintln(f, path)
		return nil
	}

	w := file.NewWalker()
	err = w.Walk(benchDir, fn)
	if err != nil {
		return err
	}

	return nil
}

func oneWorkerWalk() error {
	outPath := filepath.Join(outDir, "walk-one-worker.txt")
	f, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("failed to create the output file %q. %w", outPath, err)
	}
	defer f.Close()

	recvCh := make(chan string)

	var hadErr error
	go func() {
		var fn fs.WalkDirFunc = func(path string, d fs.DirEntry, rcvErr error) error {
			if rcvErr != nil {
				return rcvErr
			}
			recvCh <- path
			return nil
		}

		w := file.NewWalker()
		hadErr = w.Walk(benchDir, fn)
		close(recvCh)
	}()

	for path := range recvCh {
		fmt.Fprintln(f, path)
	}

	return hadErr
}
