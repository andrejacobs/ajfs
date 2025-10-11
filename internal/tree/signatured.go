package tree

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"sort"

	"github.com/andrejacobs/go-aj/file"
	"github.com/andrejacobs/go-collection/collection"
)

// SignaturedTree calculates a signature for each parent node.
// The signature is calculated based on the children node names.
// One use case is to use it to determine which sub-trees are the same (i.e. duplicates).
type SignaturedTree struct {
	rootPath string
	root     *SignaturedNode
}

// Create a new signatured tree from an existing file tree.
func NewSignaturedTree(t Tree) SignaturedTree {
	sroot := &SignaturedNode{
		Node: t.root,
	}
	stree := SignaturedTree{
		rootPath: t.rootPath,
		root:     sroot,
	}
	buildNodes(t.root, sroot, sha1.New())
	return stree
}

// Return the root node.
func (t *SignaturedTree) Root() *SignaturedNode {
	return t.root
}

// Display the entire tree.
func (t *SignaturedTree) Print(w io.Writer) {
	if t.root != nil {
		fmt.Fprintln(w, t.rootPath)
		st := stats{}
		t.root.printChildren(w, &st, "")
		fmt.Fprintln(w)
		fmt.Fprintln(w, st.String())
	}
}

// Map of signature to subtrees that have the same signature (i.e. duplicates).
type DuplicateMap map[file.PathHash][]*SignaturedNode

// Find all the subtrees that share the same signatures.
func (t *SignaturedTree) FindDuplicateSubtrees() DuplicateMap {
	dupes := make(DuplicateMap, 64)
	findDuplicateSubtrees(dupes, t.root)

	for k, v := range dupes {
		if len(v) < 2 {
			delete(dupes, k)
		}
	}

	return dupes
}

// Find all the duplicate subtrees and display them.
func (t *SignaturedTree) PrintDuplicateSubtrees(w io.Writer, printTree bool) {
	dupes := t.FindDuplicateSubtrees()
	fmt.Fprintln(w, t.rootPath)
	dupes.Print(w, printTree)
}

//-----------------------------------------------------------------------------

// Node in the tree with a signature.
type SignaturedNode struct {
	Node      *Node
	Signature file.PathHash

	FirstChild  *SignaturedNode
	NextSibling *SignaturedNode
}

// Insert the node as a child.
func (n *SignaturedNode) insertChild(c *SignaturedNode) {
	// Insert the child as the first child and thus we don't have to walk to the end of the list
	c.NextSibling = n.FirstChild
	n.FirstChild = c
}

// Recursively display the children nodes.
func (n *SignaturedNode) printChildren(w io.Writer, st *stats, prefix string) {
	// Based on kddnewton's implementation: https://github.com/kddnewton/tree/blob/main/tree.go
	children := n.sortedChildren()
	count := len(children)

	for i, child := range children {
		// Filter out hidden paths
		if len(child.Node.Name) > 0 && (child.Node.Name[0] == '.') {
			continue
		}

		if child.Node.Info.IsDir() {
			st.dirCount++
		} else {
			st.fileCount++
		}

		signature := fmt.Sprintf("%x", child.Signature)

		if i == count-1 {
			fmt.Fprintln(w, prefix+"└──", child.Node.Name, "    ["+signature+"]")
			child.printChildren(w, st, prefix+"    ")
		} else {
			fmt.Fprintln(w, prefix+"├──", child.Node.Name, "    ["+signature+"]")
			child.printChildren(w, st, prefix+"│   ")
		}

	}
}

// Return the children nodes.
func (n *SignaturedNode) children() []*SignaturedNode {
	result := make([]*SignaturedNode, 0, 8)

	current := n.FirstChild
	for {
		if current == nil {
			break
		}
		result = append(result, current)
		current = current.NextSibling
	}

	return result
}

// Return the children nodes sorted by name.
func (n *SignaturedNode) sortedChildren() []*SignaturedNode {
	result := n.children()
	sort.Slice(result, func(i, j int) bool {
		return result[i].Node.Name < result[j].Node.Name
	})
	return result
}

//-----------------------------------------------------------------------------

// Display the map of duplicates.
func (m DuplicateMap) Print(w io.Writer, printTree bool) {
	sorted := collection.MapSortedByKeysFunc(m, func(l, r file.PathHash) bool {
		lhex := hex.EncodeToString(l[:])
		rhex := hex.EncodeToString(r[:])
		return lhex < rhex
	})

	for _, kv := range sorted {
		fmt.Fprintln(w, "Signature:", fmt.Sprintf("%x", kv.Key))

		sort.Slice(kv.Value, func(i, j int) bool {
			return kv.Value[i].Node.Info.Path < kv.Value[j].Node.Info.Path
		})

		for _, node := range kv.Value {
			fmt.Fprintln(w, " ", node.Node.Info.Path)
		}

		if printTree && len(kv.Value) > 0 {
			kv.Value[0].printChildren(w, &stats{}, "  ")
		}

		fmt.Fprintln(w)
	}
}

//-----------------------------------------------------------------------------

// Build the signatured nodes from the normal tree nodes.
func buildNodes(parent *Node, signaturedParent *SignaturedNode, hasher hash.Hash) {
	if parent == nil {
		return
	}

	children := parent.sortedChildren()
	for _, child := range children {
		signaturedChild := &SignaturedNode{
			Node: child,
		}

		signaturedParent.insertChild(signaturedChild)
		buildNodes(child, signaturedChild, sha1.New())
		hasher.Write(signaturedChild.Signature[:])
	}

	io.WriteString(hasher, parent.Name)
	signaturedParent.Signature = file.PathHash(hasher.Sum(nil))
}

// Recursively build the map of duplicates.
func findDuplicateSubtrees(dupes DuplicateMap, parent *SignaturedNode) {
	if parent == nil || !parent.Node.Info.IsDir() {
		return
	}

	// Check if the signature has been seen before
	list, exists := dupes[parent.Signature]
	if !exists {
		list = make([]*SignaturedNode, 0, 4)
	}

	list = append(list, parent)
	dupes[parent.Signature] = list

	// Only traverse children for trees that have not been seen yet
	if !exists {
		child := parent.FirstChild
		for {
			if child == nil {
				break
			}
			findDuplicateSubtrees(dupes, child)
			child = child.NextSibling
		}
	}
}
