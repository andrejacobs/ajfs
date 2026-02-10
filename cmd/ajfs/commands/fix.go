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
	"os"

	"github.com/andrejacobs/ajfs/internal/app/fix"
	"github.com/spf13/cobra"
)

// ajfs fix.
var fixCmd = &cobra.Command{
	Use:   "fix",
	Short: "Attempts to repair a damaged database.",
	Long: `Attempts to repair a damaged database.

Use '--dry-run' to check the integrity of a database without making changes to
the database. This is equavalent to running 'ajfs check'.

A backup of the database header will be made before applying any changes.
The backup will be created in the current working directory using the same
filename as the database with the extension '.bak' added.

Use '--restore /path/to/___.bak' to restore a backup header to a database. 

>> Is used to display database errors that were found and that can be corrected.
!! Is used when an error happened during the process.

`,
	Example: `  # using the default ./db.ajfs database
  ajfs fix

  # using a specific database
  ajfs fix /path/to/database.ajfs

  # restore a backup header file
  ajfs fix --restore /path/to/header.ajfs.bak /path/to/database.ajfs`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := fix.Config{
			CommonConfig: commonConfig,
			Stdin:        os.Stdin,
			DryRun:       fixDryRun,
			RestorePath:  fixRestorePath,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := fix.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fixCmd)

	fixCmd.Flags().BoolVar(&fixDryRun, "dry-run", false, "Only display the repairs that will need to be performed.")
	fixCmd.Flags().StringVar(&fixRestorePath, "restore", "", "Path to a backup header to be restored.")

}

var (
	fixDryRun      bool
	fixRestorePath string
)
