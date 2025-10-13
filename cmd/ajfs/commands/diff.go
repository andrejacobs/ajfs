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
`,
	Example: `  # differences between the default ./db.ajfs database and the root path
  ajfs diff

  # differences between a specific database and its root path
  ajfs diff /path/to/database.ajfs

  # differences between two databases
  ajfs diff /path/to/lhs.ajfs /path/to/rhs.ajfs

  # differences between a database and the file system hierarchy
  ajfs diff /path/to/lhs.ajfs /path/to/rhs

  # differences between two file system hierarchies
  ajfs diff /path/to/lhs /path/to/rhs`,
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

		cfg.Fn = printDiff

		if err := diff.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func printDiff(d diff.Diff) error {
	if d.Type == diff.TypeNothing {
		return nil
	}

	fmt.Println(d.String())
	return nil
}
