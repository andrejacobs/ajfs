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

	"github.com/andrejacobs/ajfs/internal/app/diff"
	"github.com/spf13/cobra"
)

// ajfs diff.
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Display the differences between two databases and or file system hierarchies.",
	Long: `Display the differences between two databases and or file system hierarchies.

Compares the path entries from the left hand side (LHS) against those of the
right hand side (RHS) and displays what those differences are.

You can compare:
* A database against its root path to see what has changed since the database
  was created.
* A database against another database.
* A database against another file system hierarchy.
* One file system hierarchy against another one.

Differences are displayed in the following format:

* If the file or directory only exists in the left hand side (as in removed 
  from the LHS):

  ` + "`d---- Path/of/dir` or `f---- Path/of/file`." + `

* If the file or directory only exists in the right hand side (as in added 
  to the RHS):

  ` + "`d++++ Path/of/dir` or `f++++ Path/of/file`." + `

 * The item exists in both the LHS and RHS but has a change then the following
   format is used:
 
   fmslx Path/of/file

   * f or d: to denote a file or directory.
   * m: type and or permissions has changed.
   * s: size has changed.
   * l: last modification date has changed.
   * x: file signature hash has changed.
   * ~: this property has not changed.

   For example a file that has changed in size and its last modification date:

   f~sl~ Path/of/file

Differences are displayed in the following order:

* Items that only exist in the left hand side.
* Items that only exist in the right hand side.
* Items that exist on both sides and have changed.

You can also filter on items to be included or excluded from the diff output.
The filter uses the same f, d, m, s, l and x notation.
The filter can also include - for LHS, + for RHS or ~ for something has changed.
Include filters are checked first and at least one need to be matched for the item to appear in the output.
Exclude filters are checked after any include filters and an item need to not match any exclude filter to be kept
in the output.`,
	Example: `  # differences between the default ./db.ajfs database and the root path
  ajfs diff

  # differences between a specific database and its root path
  ajfs diff /path/to/database.ajfs

  # differences between two databases
  ajfs diff /path/to/lhs.ajfs /path/to/rhs.ajfs

  # differences between a database and the file system hierarchy
  ajfs diff /path/to/lhs.ajfs /path/to/rhs

  # differences between two file system hierarchies
  ajfs diff /path/to/lhs /path/to/rhs
 
  # only show differences where the size and hash has been changed
  ajfs diff --include=sx /path/to/lhs /path/to/rhs

  # only show differences where the last modification time has not been changed
  ajfs diff --exclude=l /path/to/lhs /path/to/rhs

  # ignore differences where a directory's size or a file's mode has changed (e.g. copying files from a Mac to a NAS)
  ajfs diff -e=ds -e=fm /path/to/lhs /path/to/rhs

  # only show differences for files on LHS or RHS and exclude if the size or last modification time has been changed
  ajfs diff -i=f- -i=f+ -e=s -e=l /path/to/lhs /path/to/rhs`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := diff.Config{
			CommonConfig: commonConfig,
		}

		switch len(args) {
		case 0:
			cfg.LhsPath = defaultDBPath
		case 1:
			cfg.LhsPath = args[0]
		case 2:
			cfg.LhsPath = args[0]
			cfg.RhsPath = args[1]
		}

		stats := diff.DiffStats{}
		if showStats {
			stats.Fn = printDiff
			cfg.Fn = stats.Compare
		} else if showOnlyStats {
			stats.Fn = func(d diff.Diff) error { return nil }
			cfg.Fn = stats.Compare
		} else {
			cfg.Fn = printDiff
		}

		var err error
		cfg.IncludeFilters, err = diff.ParseFilterFlagsArray(includeFilters)
		if err != nil {
			exitOnError(err, 1)
		}
		cfg.ExcludeFilters, err = diff.ParseFilterFlagsArray(excludeFilters)
		if err != nil {
			exitOnError(err, 1)
		}

		if err := diff.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

		if showStats || showOnlyStats {
			fmt.Println()
			fmt.Println("Statistics:")
			fmt.Println("-----------")
			fmt.Printf("Files:                          %d\n", stats.Files)
			fmt.Printf("Directories:                    %d\n", stats.Dirs)
			fmt.Printf("Left hand side only:            %d\n", stats.LeftOnly)
			fmt.Printf("Right hand side only:           %d\n", stats.RightOnly)
			fmt.Printf("Changed:                        %d\n", stats.Changed)
			fmt.Printf("Did not change:                 %d\n", stats.NotChanged)
			fmt.Printf("Mode changed:                   %d\n", stats.ModeChanged)
			fmt.Printf("Size changed:                   %d\n", stats.SizeChanged)
			fmt.Printf("Last modification time changed: %d\n", stats.ModTimeChanged)
			fmt.Printf("File signature hash changed:    %d\n", stats.HashChanged)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringArrayVarP(&includeFilters, "include", "i", nil, "Include filter")
	diffCmd.Flags().StringArrayVarP(&excludeFilters, "exclude", "e", nil, "Exclude filter")
	diffCmd.Flags().BoolVarP(&showStats, "stats", "s", false, "Display diffs and statistics")
	diffCmd.Flags().BoolVarP(&showOnlyStats, "only-stats", "o", false, "Display only statistics")
}

var (
	includeFilters []string
	excludeFilters []string
	showStats      bool
	showOnlyStats  bool
)

func printDiff(d diff.Diff) error {
	if d.Type == diff.TypeNothing {
		return nil
	}

	fmt.Println(d.String())
	return nil
}
