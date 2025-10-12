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

package commands

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/search"
	"github.com/spf13/cobra"
)

// ajfs search
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for matching path entries",
	Long:  `Search for matching path entries`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := search.Config{
			CommonConfig:     commonConfig,
			DisplayFullPaths: searchDisplayFullPaths,
			DisplayMinimal:   !searchDisplayMore,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := buildSearchExpression(&cfg); err != nil {
			exitOnError(err, 1)
		}

		if err := search.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVarP(&searchDisplayFullPaths, "full", "f", false, "Display full paths for entries")
	searchCmd.Flags().BoolVarP(&searchDisplayMore, "more", "m", false, "Display more information about the matching paths")

	searchCmd.Flags().StringArrayVarP(&searchRegex, "exp", "e", nil, "Match path against the regular expression")
	searchCmd.Flags().StringArrayVarP(&searchRegexInsensitive, "iexp", "i", nil, "Case insensitive match path against the regular expression")

	searchCmd.Flags().StringArrayVarP(&searchName, "name", "n", nil, "Match base name against the shell pattern (e.g. * ?)")
	searchCmd.Flags().StringArrayVar(&searchNameInsensitive, "iname", nil, "Case insensitive match base name against the shell pattern (e.g. * ?)")

	searchCmd.Flags().StringArrayVarP(&searchPath, "path", "p", nil, "Match path against the shell pattern (e.g. * ?)")
	searchCmd.Flags().StringArrayVar(&searchPathInsensitive, "ipath", nil, "Case insensitive match path against the shell pattern (e.g. * ?)")

	searchCmd.Flags().StringVarP(&searchType, "type", "t", "", `Match if the type is one of the following:
  d  directory
  f  regular file
  l  symbolic link
  p  named pipe (FIFO)
  s  socket`)

	searchCmd.Flags().StringVarP(&searchHash, "hash", "s", "", "Match if the file signature hash starts with this prefix")
	searchCmd.Flags().StringVar(&searchId, "id", "", "Match if the entry's identifier starts with this prefix")

	searchCmd.Flags().StringArrayVar(&searchSize, "size", nil, `Match the file size according to:
  <n> with no suffix means exactly <n> bytes. e.g. --size 100

  With one of the following scaling suffixes:
  k/K   Kilobytes (1 KB = 1000 bytes). e.g. --size 1k
  m/M   Megabytes (1 MB = 1000 KB). e.g. --size 1m
  g/G   Gigabytes (1 GB = 1000 MB). e.g. --size 1g
  t/T   Terrabytes (1 TB = 1000 GB). e.g. --size 1t
  p/P   Petabytes (1 PB = 1000 TB). e.g. --size 1p

  With one of the following operation prefixes:
  +   Greater than. e.g. --size +1k
  -   Less than. e.g. --size -1k`)

	searchCmd.Flags().StringVarP(&searchModTimeBefore, "before", "b", "", `Match if the entry's last modification time is before this time.
  The following formats are allowed:
  YYYY-MM-DD
  YYYY-MM-DD HH:mm:ss   Also supports YYYY-MM-DDTHH:mm:ss
  <n>D  n Days before now
  <n>M  n Months before now
  <n>Y  n Years before now
`)

	searchCmd.Flags().StringVarP(&searchModTimeAfter, "after", "a", "", `Match if the entry's last modification time is after this time.
  The following formats are allowed:
  YYYY-MM-DD
  YYYY-MM-DD HH:mm:ss   Also supports YYYY-MM-DDTHH:mm:ss
`)
}

var (
	searchRegex            []string
	searchRegexInsensitive []string

	searchName            []string
	searchNameInsensitive []string

	searchPath            []string
	searchPathInsensitive []string

	searchSize             []string
	searchType             string
	searchHash             string
	searchModTimeBefore    string
	searchModTimeAfter     string
	searchId               string
	searchDisplayFullPaths bool
	searchDisplayMore      bool
)

func buildSearchExpression(cfg *search.Config) error {

	var prev search.Expression
	var and search.Expression

	// Regex
	prev = &search.Always{}
	for _, regexStr := range searchRegex {
		exp, err := search.NewRegex(regexStr)
		if err != nil {
			return fmt.Errorf("failed to parse regular expression %q. %v", regexStr, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Case insensitive regex
	for _, regexStr := range searchRegexInsensitive {
		exp, err := search.NewRegex("(?i)" + regexStr)
		if err != nil {
			return fmt.Errorf("failed to parse regular expression '(?i)%s'. %v", regexStr, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Name (base name only)
	for _, pattern := range searchName {
		exp, err := search.NewShellPattern(pattern, true, false)
		if err != nil {
			return fmt.Errorf("failed to parse shell pattern %q. %v", pattern, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Case insensitive name (base name only)
	for _, pattern := range searchNameInsensitive {
		exp, err := search.NewShellPattern(pattern, true, true)
		if err != nil {
			return fmt.Errorf("failed to parse shell pattern %q. %v", pattern, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Path
	for _, pattern := range searchPath {
		exp, err := search.NewShellPattern(pattern, false, false)
		if err != nil {
			return fmt.Errorf("failed to parse shell pattern %q. %v", pattern, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Case insensitive path
	for _, pattern := range searchPathInsensitive {
		exp, err := search.NewShellPattern(pattern, false, true)
		if err != nil {
			return fmt.Errorf("failed to parse shell pattern %q. %v", pattern, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Size
	for _, sizeStr := range searchSize {
		exp, err := search.NewSize(sizeStr)
		if err != nil {
			return fmt.Errorf("failed to parse size expression from %q'. %v", sizeStr, err)
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Type
	if searchType != "" {
		exp, err := search.NewType(searchType)
		if err != nil {
			return err
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Hash
	if searchHash != "" {
		exp := &search.Hash{Prefix: searchHash}
		and = search.NewAnd(prev, exp)
		prev = and

		cfg.AlsoHashes = true
	}

	// Id
	if searchId != "" {
		exp := &search.Id{Prefix: searchId}
		and = search.NewAnd(prev, exp)
		prev = and
	}

	// Before date/time
	if searchModTimeBefore != "" {
		exp, err := search.NewModTimeBefore(searchModTimeBefore)
		if err != nil {
			return err
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// After date/time
	if searchModTimeAfter != "" {
		exp, err := search.NewModTimeAfter(searchModTimeAfter)
		if err != nil {
			return err
		}

		and = search.NewAnd(prev, exp)
		prev = and
	}

	// If no flags then match nothing
	if and == nil {
		and = &search.Never{}
	}

	cfg.Expresion = and
	return nil
}
