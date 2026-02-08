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
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/random"
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

	// Create a valid empty database
	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureJustEntries)
	require.NoError(t, err)

	p := path.Info{
		Id:      path.IdFromPath("."),
		Path:    ".",
		Size:    0,
		Mode:    fs.ModeDir | 0744,
		ModTime: time.Now(),
	}
	require.NoError(t, dbf.WriteEntry(&p))
	require.NoError(t, dbf.FinishEntries())
	require.NoError(t, dbf.Close())

	// Fix
	var out bytes.Buffer

	err = db.FixDatabase(&out, tempFile, false)
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries offset: 0x67
Entries: 1
Files: 0
Entries lookup table offset: 0x9d
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Hash table: No")
}

func TestFixEmptyDatabaseWithHashes(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	// Create a valid empty database with hash table
	dbf, err := db.CreateDatabase(tempFile, "/test", db.FeatureHashTable)
	require.NoError(t, err)

	p := path.Info{
		Id:      path.IdFromPath("."),
		Path:    ".",
		Size:    0,
		Mode:    fs.ModeDir,
		ModTime: time.Now(),
	}
	require.NoError(t, dbf.WriteEntry(&p))
	require.NoError(t, dbf.FinishEntries())
	require.NoError(t, dbf.StartHashTable(ajhash.AlgoSHA1))
	require.NoError(t, dbf.FinishHashTable())
	require.NoError(t, dbf.Close())

	// Fix
	var out bytes.Buffer

	err = db.FixDatabase(&out, tempFile, false)
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries offset: 0x67
Entries: 1
Files: 0
Entries lookup table offset: 0x9d
`
	assert.Contains(t, outStr, exp2)

	exp3 := `
Hash table: Yes
Hash table offset: 0xbd
Hash algorithm: SHA-1
`
	assert.Contains(t, outStr, exp3)
}

func TestFixValidDatabase(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	require.NoError(t, createTestDatabase(tempFile, false))

	// Fix
	var out bytes.Buffer

	err := db.FixDatabase(&out, tempFile, false)
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries offset: 0x67
Entries: 15
Files: 10
Entries lookup table offset: 0x454
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Hash table: No")
}

func TestFixValidDatabaseWithHashes(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	require.NoError(t, createTestDatabase(tempFile, true))

	// Fix
	var out bytes.Buffer

	err := db.FixDatabase(&out, tempFile, false)
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries offset: 0x67
Entries: 15
Files: 10
Entries lookup table offset: 0x454
`
	assert.Contains(t, outStr, exp2)

	exp3 := `
Hash table: Yes
Hash table offset: 0x5c4
Hash algorithm: SHA-1
`
	assert.Contains(t, outStr, exp3)
}

//-----------------------------------------------------------------------------

func createTestDatabase(dbPath string, hashTable bool) error {
	// Create new database and write N path info objects
	var features db.FeatureFlags = db.FeatureJustEntries
	if hashTable {
		features = db.FeatureHashTable
	}

	dbf, err := db.CreateDatabase(dbPath, "/test", features)
	if err != nil {
		return err
	}

	// Dirs
	for i := range 5 {
		dirPath := fmt.Sprintf("some/dir-%d", i)
		p := path.Info{
			Id:      path.IdFromPath(dirPath),
			Path:    dirPath,
			Size:    uint64(142),
			Mode:    0644 | fs.ModeDir,
			ModTime: time.Now().Add(-20 * time.Minute),
		}
		if err := dbf.WriteEntry(&p); err != nil {
			return err
		}
	}

	// Files
	expCount := 10
	expTime := time.Now().Add(-10 * time.Minute)

	for i := range expCount {
		filePath := fmt.Sprintf("/some/path/%d.txt", i)
		p := path.Info{
			Id:      path.IdFromPath(filePath),
			Path:    filePath,
			Size:    uint64(i),
			Mode:    0740,
			ModTime: expTime,
		}
		if err := dbf.WriteEntry(&p); err != nil {
			return err
		}
	}

	if err := dbf.FinishEntries(); err != nil {
		return err
	}

	if hashTable {
		algo := ajhash.AlgoSHA1
		if err := dbf.StartHashTable(algo); err != nil {
			return nil
		}

		for i := range expCount {
			h := make([]byte, algo.Size())
			random.SecureBytes(h)
			dbf.WriteHashEntry(i, h)
		}

		if err := dbf.FinishHashTable(); err != nil {
			return err
		}
	}

	if err := dbf.Close(); err != nil {
		return err
	}

	return nil
}
