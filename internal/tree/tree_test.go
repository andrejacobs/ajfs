package tree_test

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/ajfs/internal/tree"
	"github.com/stretchr/testify/assert"
)

func TestTree(t *testing.T) {

	tr := tree.New("/test")
	assert.Equal(t, "/test", tr.RootPath())

	n1 := makePaths(tr, "a/b/c")
	assert.NotNil(t, n1)

	n2 := makePaths(tr, "a/d")
	assert.NotNil(t, n2)

	r := makePaths(tr, ".")
	assert.NotNil(t, r)

	n3 := makePaths(tr, "etc/httpd/config.json")
	assert.NotNil(t, n3)

	expected := []string{
		".",
		"a",
		"a/b",
		"a/b/c",
		"a/d",
		"etc",
		"etc/httpd",
		"etc/httpd/config.json",
	}
	result := list(tr)

	slices.Sort(expected)
	slices.Sort(result)
	assert.Equal(t, expected, result)

	assert.Equal(t, r, tr.Find("."))
	assert.Equal(t, tr.Root(), r)
	assert.Equal(t, n1, tr.Find("a/b/c"))
	assert.Equal(t, n2, tr.Find("a/d"))
	assert.Equal(t, n3, tr.Find("etc/httpd/config.json"))
	assert.Nil(t, tr.Find("a/b/zoo"))
}

func TestPrint(t *testing.T) {
	tr := tree.New("/test")
	assert.Equal(t, "/test", tr.RootPath())

	n1 := makePaths(tr, "a/b/c")
	assert.NotNil(t, n1)

	n2 := makePaths(tr, "a/d")
	assert.NotNil(t, n2)

	r := makePaths(tr, ".")
	assert.NotNil(t, r)

	n3 := makePaths(tr, "etc/httpd/config.json")
	assert.NotNil(t, n3)

	var buffer bytes.Buffer
	tr.Print(&buffer)

	expected := `/test
├── a
│   ├── b
│   │   └── c
│   └── d
└── etc
    └── httpd
        └── config.json

5 directories, 3 files
`
	assert.Equal(t, expected, buffer.String())

	buffer.Reset()
	n4 := tr.Find("a/b")
	assert.NotNil(t, n4)

	n4.Print(&buffer)

	expected = `b
└── c

1 directory, 1 file
`
	assert.Equal(t, expected, buffer.String())

	buffer.Reset()
	n5 := tr.Insert(path.Info{Path: "z/emptyDir", Mode: fs.ModeDir})
	assert.NotNil(t, n5)

	fn := tr.Find("z")
	assert.NotNil(t, fn)

	fn.Print(&buffer)

	expected = `z
└── emptyDir

1 directory, 0 files
`
	assert.Equal(t, expected, buffer.String())

}

func TestTreePanics(t *testing.T) {
	// Lol just imagine a big tree shaking with fear

	tr := tree.Tree{}
	assert.Panics(t, func() {
		tr.Insert(path.Info{Path: "this-should-panic"})
	})
}

//-----------------------------------------------------------------------------

func list(tr tree.Tree) []string {
	result := make([]string, 0, 100)
	result = listRecursive(tr.Root(), result)
	return result
}

func listRecursive(n *tree.Node, result []string) []string {
	if n.Info.Path != "" {
		result = append(result, n.Info.Path)
	}
	current := n.FirstChild

	for {
		if current == nil {
			break
		}
		result = listRecursive(current, result)
		current = current.NextSibling
	}

	return result
}

func makePaths(tr tree.Tree, p string) *tree.Node {
	tr.Insert(path.Info{Path: ".", Mode: fs.ModeDir})

	parts := strings.Split(p, string(filepath.Separator))
	count := len(parts)
	build := ""
	var result *tree.Node
	for i, part := range parts {
		build = filepath.Join(build, part)
		mode := fs.ModeDir
		if i == count-1 {
			mode = 0
		}
		result = tr.Insert(path.Info{Path: build, Mode: mode})
	}
	return result
}
