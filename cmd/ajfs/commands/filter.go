package commands

import (
	"github.com/andrejacobs/ajfs/internal/filter"
	"github.com/spf13/cobra"
)

var (
	includePathRegex []string // Regexes for path inclusion filtering
	excludePathRegex []string // Regexes for path exclusion filtering
)

// Add the path filtering flags to the cobra command.
func addPathFilteringFlags(c *cobra.Command) {
	c.Flags().StringArrayVarP(&includePathRegex, "include", "i", nil, "Include path regex filter")
	c.Flags().StringArrayVarP(&excludePathRegex, "exclude", "e", nil, "Exclude path regex filter")
}

// Parse the include path regex into file and dir expressions.
func parseIncludePathRegex() ([]string, []string) {
	return filter.ParsePathRegex(includePathRegex)
}

// Parse the exclude path regex into file and dir expressions.
func parseExcludePathRegex() ([]string, []string) {
	return filter.ParsePathRegex(excludePathRegex)
}
