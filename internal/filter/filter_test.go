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

	fileFn, dirFn := filter.ParsePathRegexToMatchPathFn(input)

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
