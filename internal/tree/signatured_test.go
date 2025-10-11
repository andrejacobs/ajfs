package tree_test

import (
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
