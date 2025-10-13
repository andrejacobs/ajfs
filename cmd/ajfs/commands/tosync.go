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
	"github.com/andrejacobs/ajfs/internal/app/tosync"
	"github.com/spf13/cobra"
)

// ajfs tosync.
var tosyncCmd = &cobra.Command{
	Use:   "tosync",
	Short: "Show which files need to be synced from the LHS to the RHS.",
	Long: `Show which files need to be synced from the left hand side (LHS) to the
right hand side (RHS).

Think of this as a quick way to see which files on the LHS has been changed or
added and have not yet been copied onto the RHS.
 
NOTE: Does not do any syncing. This is the job for the excellent rsync.

 Criteria are:
* Only files that appear on the LHS will be shown.
* Files that have changed will be shown and thus indicate that the ones on
  the RHS need to be overwritten.
* Permissions and last modification times are ignored since these are bound
  to be different between two systems.
* If both databases have compatible file signature hashes, then items with
  a different hash will also be shown.

One of the biggest use cases for this command is to be able to see which files
on one system (e.g. laptop) has not yet been backed up somewhere on another
system (e.g. Linux server). In which case the file locations are different
between the systems. In order to do this you need to perform a scan with
file signature hash calculations on both systems and the use:
  ajfs tosync lhs.ajfs rhs.ajfs
`,
	Example: `  # compares the default database ./db.ajfs as the LHS against the RHS database
  ajfs tosync /path/to/rhs.ajf

  # compares the LHS database against the RHS database
  ajfs tosync /path/to/lhs.ajfs /path/to/rhs.ajfs

  # only compare the file signature hashes. Useful when the files are in different locations
  ajfs tosync --hash lhs.ajfs rhs.ajfs
`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := tosync.Config{
			CommonConfig: commonConfig,
			OnlyHashes:   tosyncHashesOnly,
			FullPaths:    tosyncFullPaths,
		}

		switch len(args) {
		case 1:
			cfg.LhsPath = defaultDBPath
			cfg.RhsPath = args[0]
		case 2:
			cfg.LhsPath = args[0]
			cfg.RhsPath = args[1]
		}

		cfg.Fn = printToSync

		if err := tosync.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tosyncCmd)

	tosyncCmd.Flags().BoolVarP(&tosyncHashesOnly, "hash", "s", false, "Compare only the file signature hashes.")
	tosyncCmd.Flags().BoolVarP(&tosyncFullPaths, "full", "f", false, "Display full paths for entries.")
}

var (
	tosyncHashesOnly bool
	tosyncFullPaths  bool
)

func printToSync(d diff.Diff) error {
	fmt.Println(d.Path)
	return nil
}
