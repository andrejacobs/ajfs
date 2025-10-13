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
	"github.com/andrejacobs/ajfs/internal/app/dupes"
	"github.com/spf13/cobra"
)

// ajfs dupes.
var dupesCmd = &cobra.Command{
	Use:   "dupes",
	Short: "Display all duplicate files or directory trees.",
	Long: `Display all duplicate files or directory subtrees that are the same.

The database must contain the calculated file signature hashes if you are using
this command to find duplicate files. The default mode.

Duplicate files will be displayed in the following example format:

` + "```\n>>>\n" +
		`Hash: c88e6e3b20f8478468288d2bef9cf624f5707ebcdad6113d4a545469333271a1
Size: 11167407 [11 MB]

[0]: Bought/Books/Some-awesome-book.pdf
[1]: Another/dir/somewhere/with/backups/Same-book.pdf

Count: 2
Total Size: 22334814 [22 MB]
<<<
` + "```\n" +
		`To find all duplicate subtrees use the "-d, --dirs" option.
Each parent of a subtree in the hierarchy is given a unique signature that is 
calculated based on each of its children's signatures. Thus it can be used
to find subtrees in the hierarchy that share the same children regardless
of where in the hierarchy they are.

For example: We have 2 copies of the Day1 directory.

` + "```\n" +
		`root
├── Photos
│   └── Holiday2025
│       └── Day1
│           ├── Photo1.jpg
│           └── Photo2.jpg		
└── Backup
    └── MyPhotos
        └── 2025
            └── Day1
                ├── Photo1.jpg
                └── Photo2.jpg

` + "```\n" + `Using "--dirs" would produce the example format:

` + "```\n" + `Signature: eb14cf7cad5f771d25bc5d2fa8ac012473a58044
  Photos/Holiday2025/Day1
  Backup/MyPhotos/2025/Day1

` + "```\n" + `Using "--dirs --tree" would produce the example format:

` + "```\n" + `Signature: eb14cf7cad5f771d25bc5d2fa8ac012473a58044
  Photos/Holiday2025/Day1
  Backup/MyPhotos/2025/Day1
  ├── Photo1.jpg     [15730819566f2bc79c3c6f151c5572b58b14a1c6]
  └── Photo2.jpg     [9aff76baba26e2e51f7e94b16efbf0505ddb71a9]
` + "```\n",
	Example: `  # display duplicate files from the default ./db.ajfs database
  ajfs dupes

  # display duplicate files from the specified database
  ajfs dupes /path/to/database.ajfs

  # display duplicate subtrees in the tree format
  ajfs dupes --dirs --tree /path/to/database.ajfs`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := dupes.Config{
			CommonConfig: commonConfig,
			Subtrees:     dupesDirs,
			PrintTree:    dupesDirsPrintTree,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := dupes.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(dupesCmd)

	dupesCmd.Flags().BoolVarP(&dupesDirs, "dirs", "d", false, "Display duplicate subtree directories.")
	dupesCmd.Flags().BoolVarP(&dupesDirsPrintTree, "tree", "t", false, "Display the tree hierarchy of duplicate subtrees.")
}

var (
	dupesDirs          = false
	dupesDirsPrintTree = false
)
