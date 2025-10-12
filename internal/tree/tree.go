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

// Package tree is used to represent a file hierarchy tree.
package tree

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/andrejacobs/ajfs/internal/path"
)

// Tree represents a file hierarchy.
type Tree struct {
	rootPath string
	root     *Node
}

// Create a new Tree representing the file hierarchy at the specified root path.
// All paths are relative to this root path.
func New(rootPath string) Tree {
	return Tree{
		rootPath: rootPath,
		root:     newNode("."),
	}
}

// Return the path that represents the root of the file hierarchy.
func (t *Tree) RootPath() string {
	return t.rootPath
}

// Return the root node.
func (t *Tree) Root() *Node {
	return t.root
}

// Insert a new path info object into the tree and return the new node.
func (t *Tree) Insert(pi path.Info) *Node {
	if t.root == nil {
		panic("the tree has not been initialized correctly")
	}

	if pi.Path == "." {
		t.root.Info = pi
		return t.root
	}

	// Ensure all parent nodes exist
	name := filepath.Base(pi.Path)
	parent := t.makeParents(pi.Path)

	// Check if we need to create the new node or update an existing one
	var n *Node
	existing := parent.findChild(name)
	if existing == nil {
		n = newNode(name)
		parent.insertChild(n)
	} else {
		n = existing
	}
	n.Info = pi
	return n
}

// Find the node for the specified path.
func (t *Tree) Find(path string) *Node {
	if path == "." {
		return t.root
	}

	parts := strings.Split(path, string(filepath.Separator))
	head, tail := split(parts)

	return t.findRecursive(t.root, head, tail)
}

// Display the entire tree.
func (t *Tree) Print(w io.Writer) {
	t.PrintWithLimit(w, 0)
}

// Display the tree with a maximum specified depth.
func (t *Tree) PrintWithLimit(w io.Writer, limit int) {
	if t.root != nil {
		fmt.Fprintln(w, t.rootPath)
		st := stats{
			dirCount: 1,
		}
		t.root.printChildren(w, &st, "", 1, limit)
		fmt.Fprintln(w)
		fmt.Fprintln(w, st.String())
	}
}

// Ensure all parent nodes for the path exists in the tree.
func (t *Tree) makeParents(path string) *Node {
	parts := strings.Split(path, string(filepath.Separator))
	head, tail := split(parts)

	return t.makeParentsRecursive(t.root, head, tail)
}

// Recursively create the required parent nodes.
func (t *Tree) makeParentsRecursive(current *Node, head string, tail []string) *Node {
	if len(tail) < 1 {
		return current
	}

	child := current.findChild(head)
	if child == nil {
		child = newNode(head)
		current.insertChild(child)
	}

	nextHead, nextTail := split(tail)
	return t.makeParentsRecursive(child, nextHead, nextTail)
}

// Recursively find the node matching the path.
func (t *Tree) findRecursive(parent *Node, head string, tail []string) *Node {
	child := parent.findChild(head)
	if child == nil {
		// This sub tree doesn't contain the node
		return nil
	} else {
		if len(tail) == 0 {
			if child.Name != head {
				panic("failed to find the node. algorithm is broken!")
			}
			return child
		}
	}

	nextHead, nextTail := split(tail)
	return t.findRecursive(child, nextHead, nextTail)
}

//-----------------------------------------------------------------------------

// Node in the tree describing a path entry.
type Node struct {
	Name        string
	Info        path.Info
	FirstChild  *Node
	NextSibling *Node
}

func newNode(name string) *Node {
	return &Node{
		Name: name,
	}
}

// Insert the node as a child.
func (n *Node) insertChild(c *Node) {
	// Insert the child as the first child and thus we don't have to walk to the end of the list
	c.NextSibling = n.FirstChild
	n.FirstChild = c
}

// Find the first child with the specified name.
func (n *Node) findChild(named string) *Node {
	current := n.FirstChild
	for {
		if current == nil { //nolint:staticcheck // QF1006
			break
		}
		if current.Name == named {
			return current
		}
		current = current.NextSibling
	}

	return nil
}

// Recursively display this node and children.
func (n *Node) Print(w io.Writer) {
	n.PrintWithLimit(w, 0)
}

func (n *Node) PrintWithLimit(w io.Writer, limit int) {
	st := stats{}
	if n.Info.IsDir() {
		st.dirCount = 1
	}
	fmt.Fprintln(w, n.Name)
	n.printChildren(w, &st, "", 1, limit)
	fmt.Fprintln(w)
	fmt.Fprintln(w, st.String())
}

func (n *Node) printChildren(w io.Writer, st *stats, prefix string, currentDepth int, maxDepth int) {
	if (maxDepth > 0) && (currentDepth > maxDepth) {
		return
	}

	// Based on kddnewton's implementation: https://github.com/kddnewton/tree/blob/main/tree.go
	children := n.sortedChildren()
	count := len(children)

	for i, child := range children {
		// Filter out hidden paths
		if len(child.Name) > 0 && (child.Name[0] == '.') {
			continue
		}

		if child.Info.IsDir() {
			st.dirCount++
		} else {
			st.fileCount++
		}

		if i == count-1 {
			fmt.Fprintln(w, prefix+"└──", child.Name)
			child.printChildren(w, st, prefix+"    ", currentDepth+1, maxDepth)
		} else {
			fmt.Fprintln(w, prefix+"├──", child.Name)
			child.printChildren(w, st, prefix+"│   ", currentDepth+1, maxDepth)
		}
	}
}

// Return the children nodes.
func (n *Node) children() []*Node {
	result := make([]*Node, 0, 8)

	current := n.FirstChild
	for {
		if current == nil { //nolint:staticcheck // QF1006
			break
		}
		result = append(result, current)
		current = current.NextSibling
	}

	return result
}

// Return the children nodes sorted by name.
func (n *Node) sortedChildren() []*Node {
	result := n.children()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

//-----------------------------------------------------------------------------

type stats struct {
	dirCount  int64
	fileCount int64
}

func (s *stats) String() string {
	var dirStr string
	if s.dirCount == 1 {
		dirStr = "directory"
	} else {
		dirStr = "directories"
	}

	var fileStr string
	if s.fileCount == 1 {
		fileStr = "file"
	} else {
		fileStr = "files"
	}

	return fmt.Sprintf("%d %s, %d %s", s.dirCount, dirStr, s.fileCount, fileStr)
}

//-----------------------------------------------------------------------------

// Split the path parts into a head and the tail.
func split(parts []string) (head string, tail []string) {
	var h string
	var t []string
	pos := 0
	for {
		h, t = parts[pos], parts[pos+1:]
		if h != "" {
			break
		}
		pos += 1
	}

	return h, t
}
