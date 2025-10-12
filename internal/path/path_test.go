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

package path_test

import (
	"crypto/sha1"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/stretchr/testify/assert"
)

func TestIdFromPath(t *testing.T) {
	id := path.IdFromPath("/usr/bin")
	assert.Equal(t, path.Id(sha1.Sum([]byte("/usr/bin"))), id)
}

func TestPathInfoEquals(t *testing.T) {
	p1 := path.Info{
		Id:      path.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: time.Now(),
	}

	p2 := path.Info{
		Id:      path.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime,
	}

	assert.True(t, p1.Equals(&p2))
	assert.True(t, p2.Equals(&p1))

	assert.False(t, p1.Equals(&path.Info{
		Id:      path.IdFromPath("a/b/d"),
		Path:    "a/b/d",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&path.Info{
		Id:      path.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    24,
		Mode:    0644,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&path.Info{
		Id:      path.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0744,
		ModTime: p1.ModTime,
	}))

	assert.False(t, p1.Equals(&path.Info{
		Id:      path.IdFromPath("a/b/c"),
		Path:    "a/b/c",
		Size:    42,
		Mode:    0644,
		ModTime: p1.ModTime.Add(10 * time.Second),
	}))

}
