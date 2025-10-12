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
	Short: "Show which files need to be synced from the LHS to the RHS",
	Long:  `Show which files need to be synced from the LHS to the RHS`,
	Args:  cobra.RangeArgs(1, 2),
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

	tosyncCmd.Flags().BoolVarP(&tosyncHashesOnly, "hash", "s", false, "Compare only the file signature hashes")
	tosyncCmd.Flags().BoolVarP(&tosyncFullPaths, "full", "f", false, "Display full paths for entries")
}

var (
	tosyncHashesOnly bool
	tosyncFullPaths  bool
)

func printToSync(d diff.Diff) error {
	fmt.Println(d.Path)
	return nil
}
