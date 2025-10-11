package dupes_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/dupes"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoHashes(t *testing.T) {
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

	cfg := dupes.Config{
		CommonConfig: scanCfg.CommonConfig,
	}

	err = dupes.Run(cfg)
	require.ErrorContains(t, err, "require file signature hashes to be present in the database")
}

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
		Root:            "../../testdata/scan",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cfg := dupes.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
	}

	err = dupes.Run(cfg)
	require.NoError(t, err)

	expected := `>>>
Hash: e3d157020b35944b552ba9987eb668228c073d30
Size: 484 [484 B]

[0]: 1.txt
[1]: a/a1/a1a/a1a1/1.txt
[2]: a/a2/same-as-1.txt
[3]: b/b1/b1a/1.txt
[4]: b/b1/b1a/same-as-1.txt

Count: 5
Total Size: 2420 [2.4 kB]
<<<

Total size of all duplicates: 2420 [2.4 kB]
`
	assert.Equal(t, expected, outBuffer.String())
	assert.Equal(t, "", errBuffer.String())
}

func TestSubtrees(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/dupe-dirs",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	cfg := dupes.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
		Subtrees: true,
	}

	err = dupes.Run(cfg)
	require.NoError(t, err)

	absRoot, err := filepath.Abs(scanCfg.Root)
	require.NoError(t, err)

	expected := absRoot + `
Signature: 5c09ba250cd65d1d4e244c268346af99b77209ba
  a/a2
  dupes/c/a2

`
	assert.Equal(t, expected, outBuffer.String())
	assert.Equal(t, "", errBuffer.String())

	outBuffer.Reset()
	errBuffer.Reset()

	cfg.PrintTree = true
	err = dupes.Run(cfg)
	require.NoError(t, err)

	expected = absRoot + `
Signature: 5c09ba250cd65d1d4e244c268346af99b77209ba
  a/a2
  dupes/c/a2
  ├── 6.txt     [88a4f09d6fde8cfce369b00ca4b2193469f9d103]
  └── same-as-1.txt     [248d286d0d77ab8eb9349b456d2daf3e1066ea78]

`
	assert.Equal(t, expected, outBuffer.String())
	assert.Equal(t, "", errBuffer.String())
}
