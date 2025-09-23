package scan_test

import (
	"crypto/sha1"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/scan"
	"github.com/stretchr/testify/assert"
)

func TestIdFromPath(t *testing.T) {
	id := scan.IdFromPath("/usr/bin")
	assert.Equal(t, scan.PathId(sha1.Sum([]byte("/usr/bin"))), id)
}

func TestPathInfoEquals(t *testing.T) {
	p1 := scan.PathInfo{
		Id:      scan.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: time.Now(),
	}

	p2 := scan.PathInfo{
		Id:      scan.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime,
	}

	assert.True(t, p1.Equals(&p2))
	assert.True(t, p2.Equals(&p1))

	assert.False(t, p1.Equals(&scan.PathInfo{
		Id:      scan.IdFromPath("a/b/d"),
		Path:    "a/b/d",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&scan.PathInfo{
		Id:      scan.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    24,
		Mode:    0644,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&scan.PathInfo{
		Id:      scan.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0744,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&scan.PathInfo{
		Id:      scan.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime.Add(10 * time.Second),
	}))

}
