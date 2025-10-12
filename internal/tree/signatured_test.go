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

package tree_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/andrejacobs/ajfs/internal/tree"
	"github.com/stretchr/testify/assert"
)

func TestSignaturedTreeFindDuplicates(t *testing.T) {

	tr := tree.New("/test")

	makePaths(tr, "orig/a/b/c/1.txt")
	makePaths(tr, "orig/a/b/d/2.txt")

	makePaths(tr, "dupes/deeper/a/b/c/1.txt")
	makePaths(tr, "dupes/deeper/a/b/d/2.txt")

	makePaths(tr, "not-dupe/a/d/x")
	makePaths(tr, "not-dupe/a/d/x/more-files")

	stree := tree.NewSignaturedTree(tr)

	expected := []string{
		".",
		"dupes",
		"dupes/deeper",
		"dupes/deeper/a",
		"dupes/deeper/a/b",
		"dupes/deeper/a/b/c",
		"dupes/deeper/a/b/c/1.txt",
		"dupes/deeper/a/b/d",
		"dupes/deeper/a/b/d/2.txt",
		"not-dupe",
		"not-dupe/a",
		"not-dupe/a/d",
		"not-dupe/a/d/x",
		"not-dupe/a/d/x/more-files",
		"orig",
		"orig/a",
		"orig/a/b",
		"orig/a/b/c",
		"orig/a/b/c/1.txt",
		"orig/a/b/d",
		"orig/a/b/d/2.txt",
	}

	result := listSignatured(stree)

	slices.Sort(expected)
	slices.Sort(result)
	assert.Equal(t, expected, result)

	dupes := stree.FindDuplicateSubtrees()
	assert.Len(t, dupes, 1)
	for _, v := range dupes {
		assert.Len(t, v, 2)
		result := []string{v[0].Node.Info.Path, v[1].Node.Info.Path}
		expected := []string{"orig/a", "dupes/deeper/a"}
		slices.Sort(result)
		slices.Sort(expected)
		assert.Equal(t, expected, result)
	}

	// stree.Print(os.Stdout)
	// fmt.Printf("=======================================================\n\n")
	// stree.PrintDuplicateSubtrees(os.Stdout, false)
}

func TestSignaturedPrint(t *testing.T) {
	tr := tree.New("/test")
	makePaths(tr, "a/b/c.txt")
	makePaths(tr, "a/d/e.txt")

	stree := tree.NewSignaturedTree(tr)

	var buffer bytes.Buffer
	stree.Print(&buffer)

	expected := `/test
└── a     [738f85b5fb947ee653004952949083d597e1da45]
    ├── b     [f08d52d8c5774306baf4203c859efb1666e18df6]
    │   └── c.txt     [fe4c80bb098894b4d6ca36c16082d567bfd41b8b]
    └── d     [c9fc6719efc1551cdcae31a81c1d251019ce08fe]
        └── e.txt     [0acf62922eb3f6dc602c80fe8225cf7419f50adc]

3 directories, 2 files
`
	assert.Equal(t, expected, buffer.String())
}

func TestPrintDuplicateSubtrees(t *testing.T) {
	tr := tree.New("/test")
	makePaths(tr, "a/b/c.txt")
	makePaths(tr, "a/d/e.txt")
	makePaths(tr, "dupes/b/c.txt")
	makePaths(tr, "dupes/x/y/z/d/e.txt")

	stree := tree.NewSignaturedTree(tr)

	var buffer bytes.Buffer
	stree.PrintDuplicateSubtrees(&buffer, true)

	expected := `/test
Signature: c9fc6719efc1551cdcae31a81c1d251019ce08fe
  a/d
  dupes/x/y/z/d
  └── e.txt     [0acf62922eb3f6dc602c80fe8225cf7419f50adc]

Signature: f08d52d8c5774306baf4203c859efb1666e18df6
  a/b
  dupes/b
  └── c.txt     [fe4c80bb098894b4d6ca36c16082d567bfd41b8b]

`
	assert.Equal(t, expected, buffer.String())
}

//-----------------------------------------------------------------------------

func listSignatured(tr tree.SignaturedTree) []string {
	result := make([]string, 0, 100)
	result = listSignaturedRecursive(tr.Root(), result)
	return result
}

func listSignaturedRecursive(n *tree.SignaturedNode, result []string) []string {
	if n.Node.Info.Path != "" {
		result = append(result, n.Node.Info.Path)
	}
	current := n.FirstChild

	for {
		if current == nil {
			break
		}
		result = listSignaturedRecursive(current, result)
		current = current.NextSibling
	}

	return result
}
