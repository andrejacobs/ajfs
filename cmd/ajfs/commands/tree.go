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
	"github.com/andrejacobs/ajfs/internal/app/tree"
	"github.com/andrejacobs/go-aj/file"
	"github.com/spf13/cobra"
)

// ajfs tree.
var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display the file hiearchy tree.",
	Long:  `Display the file hiearchy as a tree in a similar way the popular tree command does.`,
	Example: `  # display the entire hierarchy from the default ./db.ajfs
  ajfs tree

  # display the entire hierarchy of the specified database
  ajfs tree /path/to/database.ajfs

  # display a subtree from the default ./db.ajfs
  ajfs tree /sub/tree/path/inside

  # display only directories
  ajfs tree --dirs /path/to/database.ajfs

  # display only directories and limit the depth to 3 layers starting at the subtree
  ajfs tree --dirs --limit 3 /path/to/database.ajfs /sub/tree/path/inside`,
	Args: cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := tree.Config{
			CommonConfig: commonConfig,
			OnlyDirs:     treeOnlyDirs,
			Limit:        treeLimit,
		}

		switch len(args) {
		case 0:
			cfg.DbPath = defaultDBPath
		case 1:
			exists, err := file.FileExists(args[0])
			if err != nil {
				exitOnError(err, 1)
			}

			if exists {
				cfg.DbPath = args[0]
			} else {
				cfg.DbPath = defaultDBPath
				cfg.Subpath = args[0]
			}
		case 2:
			cfg.DbPath = args[0]
			cfg.Subpath = args[1]
		default:
			panic("invalid args")
		}

		if err := tree.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)

	treeCmd.Flags().BoolVarP(&treeOnlyDirs, "dirs", "d", false, "Display only directories.")
	treeCmd.Flags().IntVarP(&treeLimit, "limit", "l", 0, "Limit the tree depth.")
}

var (
	treeOnlyDirs bool
	treeLimit    int
)
