package commands

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/filter"
	"github.com/andrejacobs/go-aj/file"
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

// Parse the include path regexes into file and dir path matchers.
func parseIncludePathRegex() (file.MatchPathFn, file.MatchPathFn, error) {
	return filter.ParsePathRegexToMatchPathFn(includePathRegex, true)
}

// Parse the exclude path regexes into file and dir path matchers.
func parseExcludePathRegex() (file.MatchPathFn, file.MatchPathFn, error) {
	return filter.ParsePathRegexToMatchPathFn(excludePathRegex, false)
}

// Parse the filtering config that can be used by commands.
func parseFilterConfig() (*config.FilterConfig, error) {
	result := &config.FilterConfig{}

	incF, incD, err := parseIncludePathRegex()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the include filtering flags. %w", err)
	}

	result.FileIncluder = incF
	result.DirIncluder = incD

	exclF, exclD, err := parseExcludePathRegex()
	if err != nil {
		return nil, fmt.Errorf("failed to parse the exclude filtering flags. %w", err)
	}

	result.FileExcluder = file.MatchAppleDSStore(exclF)
	result.DirExcluder = exclD
	return result, nil
}
