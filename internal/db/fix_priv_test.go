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

package db

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/file"
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
	dbf, err := CreateDatabase(tempFile, "/test", FeatureJustEntries)
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

	err = FixDatabase(&out, tempFile, false, tempFile+".bak")
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries: 1
Files: 0
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Hash table: No")
	assert.Contains(t, outStr, "Nothing to be fixed")
	assert.Contains(t, outStr, "Entries offset:")
	assert.Contains(t, outStr, "Entries lookup table offset:")
}

func TestFixEmptyDatabaseWithHashes(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	// Create a valid empty database with hash table
	dbf, err := CreateDatabase(tempFile, "/test", FeatureHashTable)
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

	err = FixDatabase(&out, tempFile, false, tempFile+".bak")
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries: 1
Files: 0
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Entries offset:")
	assert.Contains(t, outStr, "Entries lookup table offset:")
	assert.Contains(t, outStr, "Hash table: Yes")
	assert.Contains(t, outStr, "Hash algorithm: SHA-1")
	assert.Contains(t, outStr, "Hash table offset:")
	assert.Contains(t, outStr, "Nothing to be fixed")
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

	err := FixDatabase(&out, tempFile, false, tempFile+".bak")
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries: 15
Files: 10
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Entries offset:")
	assert.Contains(t, outStr, "Entries lookup table offset:")
	assert.Contains(t, outStr, "Hash table: No")
	assert.Contains(t, outStr, "Nothing to be fixed")
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

	err := FixDatabase(&out, tempFile, false, tempFile+".bak")
	require.NoError(t, err)

	outStr := out.String()

	exp1 := `Signature: AJFS
Version: 1
Root: "/test"
`
	assert.Contains(t, outStr, exp1)
	assert.NotContains(t, outStr, ">>")

	exp2 := `Entries: 15
Files: 10
`
	assert.Contains(t, outStr, exp2)
	assert.Contains(t, outStr, "Entries offset:")
	assert.Contains(t, outStr, "Entries lookup table offset:")
	assert.Contains(t, outStr, "Hash table: Yes")
	assert.Contains(t, outStr, "Hash table offset:")
	assert.Contains(t, outStr, "Hash algorithm: SHA-1")
	assert.Contains(t, outStr, "Nothing to be fixed")
}

func TestFixNotADatabase(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	require.NoError(t, random.CreateFile(tempFile, 100))

	err := FixDatabase(io.Discard, tempFile, false, tempFile+".bak")
	require.ErrorContains(t, err, "not a valid ajfs file (invalid signature")
}

func TestFixZeroHeader(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	require.NoError(t, createTestDatabase(tempFile, false))

	expectedHeader, err := readHeader(tempFile)
	require.NoError(t, err)

	zeroHeader := header{}
	require.NoError(t, replaceHeader(zeroHeader, tempFile))

	var out bytes.Buffer

	bakPath := tempFile + ".bak"
	t.Cleanup(func() {
		os.Remove(bakPath)
	})

	// dry run
	err = FixDatabase(&out, tempFile, true, bakPath)
	require.ErrorContains(t, err, "database needs to be fixed")

	exists, err := file.FileExists(bakPath)
	require.NoError(t, err)
	assert.False(t, exists)

	outStr := out.String()

	assert.Contains(t, outStr, ">> Entries offset is expected to be")
	assert.Contains(t, outStr, ">> Entries count is expected to be 15, actual is 0")
	assert.Contains(t, outStr, ">> File entries count is expected to be 10, actual is 0")
	assert.Contains(t, outStr, ">> Entries lookup table offset is expected to be")
	assert.Contains(t, outStr, ">> Features offset is expected to be")
	assert.Contains(t, outStr, ">> Checksum is expected to be")
	assert.Contains(t, outStr, "Database needs to be fixed. Skipping because running in dry-run mode.")

	// fix
	out.Reset()
	require.NoError(t, FixDatabase(&out, tempFile, false, bakPath))
	outStr = out.String()
	assert.Contains(t, outStr, ">>")

	exists, err = file.FileExists(bakPath)
	require.NoError(t, err)
	assert.True(t, exists)

	bakSize, err := file.FileSize(bakPath)
	require.NoError(t, err)
	assert.Equal(t, headerOffset()+headerSize(), bakSize)

	bakHeader, err := readHeader(bakPath)
	require.NoError(t, err)
	assert.Equal(t, zeroHeader, bakHeader)

	// fix (with no fixes)
	out.Reset()
	require.NoError(t, FixDatabase(&out, tempFile, false, bakPath))
	outStr = out.String()
	assert.NotContains(t, outStr, ">>")

	resultHeader, err := readHeader(tempFile)
	require.NoError(t, err)
	assert.Equal(t, expectedHeader, resultHeader)
}

func TestFixZeroHeaderWithHashes(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})

	require.NoError(t, createTestDatabase(tempFile, true))

	expectedHeader, err := readHeader(tempFile)
	require.NoError(t, err)

	zeroHeader := header{}
	require.NoError(t, replaceHeader(zeroHeader, tempFile))

	var out bytes.Buffer

	bakPath := tempFile + ".bak"
	t.Cleanup(func() {
		os.Remove(bakPath)
	})

	// dry run
	err = FixDatabase(&out, tempFile, true, bakPath)
	require.ErrorContains(t, err, "database needs to be fixed")

	exists, err := file.FileExists(bakPath)
	require.NoError(t, err)
	assert.False(t, exists)

	outStr := out.String()

	assert.Contains(t, outStr, ">> Entries offset is expected to be")
	assert.Contains(t, outStr, ">> Entries count is expected to be 15, actual is 0")
	assert.Contains(t, outStr, ">> File entries count is expected to be 10, actual is 0")
	assert.Contains(t, outStr, ">> Entries lookup table offset is expected to be")
	assert.Contains(t, outStr, ">> Features offset is expected to be")
	assert.Contains(t, outStr, ">> Checksum is expected to be")
	assert.Contains(t, outStr, ">> Hash table offset is expected to be")
	assert.Contains(t, outStr, "Database needs to be fixed. Skipping because running in dry-run mode.")

	// fix
	out.Reset()
	require.NoError(t, FixDatabase(&out, tempFile, false, bakPath))
	outStr = out.String()
	assert.Contains(t, outStr, ">>")
	assert.Contains(t, outStr, "Backing up headers to:")

	exists, err = file.FileExists(bakPath)
	require.NoError(t, err)
	assert.True(t, exists)

	bakSize, err := file.FileSize(bakPath)
	require.NoError(t, err)
	assert.Equal(t, headerOffset()+headerSize(), bakSize)

	bakHeader, err := readHeader(bakPath)
	require.NoError(t, err)
	assert.Equal(t, zeroHeader, bakHeader)

	// fix (with no fixes)
	out.Reset()
	require.NoError(t, FixDatabase(&out, tempFile, false, bakPath))
	outStr = out.String()
	assert.NotContains(t, outStr, ">>")
	assert.NotContains(t, outStr, "Backing up headers to:")

	resultHeader, err := readHeader(tempFile)
	require.NoError(t, err)
	assert.Equal(t, expectedHeader, resultHeader)
}

func TestRestoreDatabaseHeaderInvalidFile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.not-ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})
	require.NoError(t, random.CreateFile(tempFile, 200))

	validFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(validFile)
	t.Cleanup(func() {
		os.Remove(validFile)
	})
	require.NoError(t, createTestDatabase(validFile, false))

	bakFile := filepath.Join(t.TempDir(), "unit-test.ajfs.bak")
	_ = os.Remove(bakFile)
	t.Cleanup(func() {
		os.Remove(bakFile)
	})
	require.NoError(t, saveDatabaseHeaders(validFile, bakFile))

	assert.ErrorContains(t, RestoreDatabaseHeader(tempFile, bakFile), "not a valid ajfs file")

	invalidBakFile := filepath.Join(t.TempDir(), "unit-test.not-ajfs-bak")
	_ = os.Remove(invalidBakFile)
	t.Cleanup(func() {
		os.Remove(invalidBakFile)
	})
	require.NoError(t, random.CreateFile(invalidBakFile, 200))

	assert.ErrorContains(t, RestoreDatabaseHeader(validFile, invalidBakFile), "not a valid backup file")
}

func TestRestoreDatabaseHeader(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	t.Cleanup(func() {
		os.Remove(tempFile)
	})
	require.NoError(t, createTestDatabase(tempFile, false))

	bakFile := filepath.Join(t.TempDir(), "unit-test.ajfs.bak")
	_ = os.Remove(bakFile)
	t.Cleanup(func() {
		os.Remove(bakFile)
	})
	require.NoError(t, saveDatabaseHeaders(tempFile, bakFile))

	// Damage database
	require.NoError(t, replaceHeader(header{}, tempFile))
	require.ErrorContains(t, FixDatabase(io.Discard, tempFile, true, ""), "database needs to be fixed")

	// Restore backup header
	assert.NoError(t, RestoreDatabaseHeader(tempFile, bakFile))
	require.NoError(t, FixDatabase(io.Discard, tempFile, true, bakFile))
}

//-----------------------------------------------------------------------------

func createTestDatabase(dbPath string, hashTable bool) error {
	// Create new database and write N path info objects
	var features FeatureFlags = FeatureJustEntries
	if hashTable {
		features = FeatureHashTable
	}

	dbf, err := CreateDatabase(dbPath, "/test", features)
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
