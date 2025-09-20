package walk_test

// Run with: BENCH_DIR="./path/to/dataset" go test -run=^$ -bench=. -benchtime=3x
// -benchtime=10x Limits iterations to 10 times
// -benchtime=10s Limits iteration time to 10 seconds

import (
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/andrejacobs/go-aj/file"
)

const (
	benchDirEnv = "BENCH_DIR"
)

var benchDir string

func setup() error {
	benchDir = os.Getenv(benchDirEnv)
	if benchDir == "" {
		return fmt.Errorf("environment variable %q needs to specify that directory to be used in the benchmark", benchDirEnv)
	}

	return nil
}

func Benchmark(b *testing.B) {
	if err := setup(); err != nil {
		b.Fatal(err)
	}

	fmt.Printf("Using %q as the dataset\n", benchDir)

	b.Run("vanilla", func(b *testing.B) {
		for b.Loop() {
			if err := vanillaWalk(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func vanillaWalk() error {
	var fn fs.WalkDirFunc = func(path string, d fs.DirEntry, rcvErr error) error {
		if rcvErr != nil {
			return rcvErr
		}
		return nil
	}

	w := file.NewWalker()
	err := w.Walk(benchDir, fn)
	if err != nil {
		return err
	}

	return nil
}
