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

package resume_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/export"
	"github.com/andrejacobs/ajfs/internal/app/resume"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/testshared"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResume(t *testing.T) {
	testCases := []struct {
		algo         ajhash.Algo
		hashDeepFile string
	}{
		{
			algo:         ajhash.AlgoSHA1,
			hashDeepFile: "../../testdata/expected/scan.sha1",
		},
		{
			algo:         ajhash.AlgoSHA256,
			hashDeepFile: "../../testdata/expected/scan.sha256",
		},
		// Can't test SHA-512 atm because hashdeep doesn't support it
	}
	for _, tC := range testCases {
		t.Run(tC.algo.String(), func(t *testing.T) {
			tempFile := filepath.Join(os.TempDir(), "unit-testing")
			_ = os.Remove(tempFile)
			defer os.Remove(tempFile)

			// Create initial database
			cfg := scan.Config{
				CommonConfig: config.CommonConfig{
					DbPath: tempFile,
					Stdout: io.Discard,
					Stderr: io.Discard,
				},
				Root:            "../../testdata/scan",
				CalculateHashes: true,
				Algo:            tC.algo,
				InitOnly:        true,
			}

			err := scan.Run(cfg)
			require.NoError(t, err)

			// Resume calculating hashes
			resumeCfg := resume.Config{
				CommonConfig: cfg.CommonConfig,
			}

			err = resume.Run(resumeCfg)
			require.NoError(t, err)

			// Export hashdeep
			tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.hashdeep")
			_ = os.Remove(tempExportFile)
			defer os.Remove(tempExportFile)

			exportCfg := export.Config{
				CommonConfig: cfg.CommonConfig,
				Format:       export.FormatHashdeep,
				ExportPath:   tempExportFile,
			}
			err = export.Run(exportCfg)
			require.NoError(t, err)

			// Validate
			expectedHashDeep, err := testshared.ReadHashDeepFile(tC.hashDeepFile)
			require.NoError(t, err)

			exportedHashDeep, err := testshared.ReadHashDeepFile(tempExportFile)
			require.NoError(t, err)

			assert.ElementsMatch(t, expectedHashDeep, exportedHashDeep)
		})
	}
}
