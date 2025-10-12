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

package filter_test

import (
	"io/fs"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePathRegex(t *testing.T) {
	input := []string{"f:file1", "f:file2", "d:dir1", "d:dir2", "both", ""}

	file, dir := filter.ParsePathRegex(input)
	assert.Equal(t, []string{"file1", "file2", "both"}, file)
	assert.Equal(t, []string{"dir1", "dir2", "both"}, dir)
}

func TestParsePathRegexToMatchPathFn(t *testing.T) {
	input := []string{"f:file1", "f:file2", "d:dir1", "d:dir2", "both", ""}

	fileFn, dirFn, err := filter.ParsePathRegexToMatchPathFn(input, false)
	require.NoError(t, err)

	r, err := fileFn("a/file1", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = fileFn("a/b/file2", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = fileFn("a/b/c/both/xyz", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = fileFn("a/no-match", fakeDirEntry{})
	require.NoError(t, err)
	assert.False(t, r)

	r, err = dirFn("a/dir1", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = dirFn("a/b/dir2", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = dirFn("a/b/c/both/xyz", fakeDirEntry{})
	require.NoError(t, err)
	assert.True(t, r)

	r, err = dirFn("a/no-match", fakeDirEntry{})
	require.NoError(t, err)
	assert.False(t, r)
}

//-----------------------------------------------------------------------------

// Pretend to be a fs.DirEntry
type fakeDirEntry struct {
}

func (f fakeDirEntry) Name() string {
	return ""
}

func (f fakeDirEntry) IsDir() bool {
	return false
}

func (f fakeDirEntry) Type() fs.FileMode {
	return 0
}

func (f fakeDirEntry) Info() (fs.FileInfo, error) {
	return fakeFileInfo{}, nil
}

// Pretend to be a fs.FileInfo
type fakeFileInfo struct {
}

func (f fakeFileInfo) Name() string {
	return ""
}

func (f fakeFileInfo) Size() int64 {
	return 0
}

func (f fakeFileInfo) Mode() fs.FileMode {
	return 0
}

func (f fakeFileInfo) ModTime() time.Time {
	return time.Now()
}
func (f fakeFileInfo) IsDir() bool {
	return false
}

func (f fakeFileInfo) Sys() any {
	return nil
}
