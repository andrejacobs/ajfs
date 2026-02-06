// Copyright (c) 2026 Andre Jacobs
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
	"github.com/andrejacobs/ajfs/internal/app/fix"
	"github.com/spf13/cobra"
)

// ajfs fix.
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Attempts to repair a damaged database.",
	Long:  `TODO.`,
	Example: `  # using the default ./db.ajfs database
  ajfs fix

  # using a specific database
  ajfs fix /path/to/database.ajfs`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fix.Config{
			CommonConfig: commonConfig,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := fix.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	fixCmd.Flags().BoolVar(&scanDryRun, "dry-run", false, "Only display the repairs that will need to be performed.")

	//AJ### Add a flag to make a backup copy first (does not apply when dry-run is true)
}
