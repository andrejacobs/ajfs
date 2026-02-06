// Copyright (c) 2026 Andre Jacobs
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

package db_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Empty database (nothing to fix)
// Empty header, but has entries

func TestFixEmptyDatabase(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	var out bytes.Buffer

	err = db.FixDatabase(&out, tempFile, false)
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)

}

//-----------------------------------------------------------------------------

func createTestDatabase(path string, hashTable bool) error {

	// // Create new database and write N path info objects
	// dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	// require.NoError(t, err)

	// expCount := 10
	// expTime := time.Now().Add(-10 * time.Minute)

	// for i := range expCount {
	// 	filePath := fmt.Sprintf("/some/path/%d.txt", i)
	// 	p := path.Info{
	// 		Id:      path.IdFromPath(filePath),
	// 		Path:    filePath,
	// 		Size:    uint64(i),
	// 		Mode:    0740,
	// 		ModTime: expTime,
	// 	}
	// 	require.NoError(t, dbf.WriteEntry(&p))
	// }

	// require.NoError(t, dbf.FinishEntries())
	// require.NoError(t, dbf.Close())

	return fmt.Errorf("Whoosh")
}
