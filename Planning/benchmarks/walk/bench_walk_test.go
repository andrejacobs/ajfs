package walk_test

// Run with: BENCH_DIR="./path/to/dataset" go test -run=^$ -bench=. -benchtime=3x
// Optionally BENCH_OUT="~/temp/output"

// Reference:
// -benchtime=10x Limits iterations to 10 times
// -benchtime=10s Limits iteration time to 10 seconds

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
