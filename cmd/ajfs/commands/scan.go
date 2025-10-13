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
	"strings"

	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/spf13/cobra"
)

// ajfs scan.
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Create a new database.",
	Long: `Create a new database by walking an existing file system hierarchy.

The directory to be walked is also known as the root path and this path is
stored in the database. Use "ajfs info" to see the root path.
The root path can be used to display the full path for database entries while
using the "-f, --full" flag on some of the other commands.

Additionally the file signature hashes of files can be calculated and stored in
the database. This can be very valuable for later finding duplicates or
differences. Calculating the file signature hashes can be a long running
process depending on the number of files and sizes.

The file signature hash calculation process can be safely interrupted using
Ctrl+C (SIGTERM) and be resumed at another time using "ajfs resume".
However it is not safe to interrupt the initial database creation, 
use "--verbose" or "--progress" to know when the calculation process has
started.

Path filtering:

Used to check whether a file or directory should be included or if it should
be excluded. An include filter will always be performed first and thus skip
any exclude filters.

You can include multiple filters on the CLI (e.g. "-i someting -i another")

  "-i, --include {pattern}"
  "-e, --exclude {pattern}"

Pattern is a regular expression that can be optionally prefixed with "f:" for
file or "d:" for directory.
For example to include all files that match the extension .pdf and exclude
any directories that end with temp, you could use this on the CLI
-i "f:\.pdf$" -e "d:temp$".
If the prefix (f: or d:) is not specified then the regular expression will be
applied to both files and directories.

See https://pkg.go.dev/regexp/syntax for the syntax.`,
	Example: `  # create the default ./db.ajfs database from the specified path
  ajfs scan /path/to/be/scanned

  # create a new database from the specified path
  ajfs scan /path/to/database.ajfs /path/to/be/scanned

  # see which paths will be included without creating the database
  ajfs scan --dry-run -i "f:\.pdf$" /path/to/be/scanned

  # override the existing database if it exists
  ajfs scan --force /path/to/database.ajfs /path/to/be/scanned

  # create a new database and calculate the file signature hashes using SHA-256
  ajfs scan --hash /path/to/database.ajfs /path/to/be/scanned

  # create a new database and calculate the file signature hashes using SHA-1 while showing a progress bar
  ajfs scan --hash --algo=sha1 --progress /path/to/database.ajfs /path/to/be/scanned

  # create a new database and only include PDF and EPUB files
  ajfs scan -i "f:\.pdf$" -i "f:\.epub$" /path/to/be/scanned

  # create a new database and exclude all directories that contain the word "temp"
  ajfs scan -e "d:temp" /path/to/be/scanned`,
	Args: cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		filterCfg, err := parseFilterConfig()
		if err != nil {
			exitOnError(err, 1)
		}

		commonConfig.Progress = showProgress

		cfg := scan.Config{
			CommonConfig:  commonConfig,
			FilterConfig:  *filterCfg,
			ForceOverride: scanForceOverride,
			DryRun:        scanDryRun,
		}

		switch len(args) {
		case 1:
			cfg.DbPath = defaultDBPath
			cfg.Root = args[0]
		case 2:
			cfg.DbPath = args[0]
			cfg.Root = args[1]
		default:
			panic("invalid args")
		}

		if scanCalculateHashes {
			algo, err := algoFromFlag(scanHashAlgo)
			if err != nil {
				exitOnError(err, 1)
			}

			cfg.CalculateHashes = true
			cfg.Algo = algo
		}

		if err := scan.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVar(&scanForceOverride, "force", false, "Override any existing database.")
	scanCmd.Flags().BoolVarP(&scanCalculateHashes, "hash", "s", false, "Calculate file signature hashes.")
	scanCmd.Flags().BoolVar(&scanDryRun, "dry-run", false, "Only display files and directories that would be stored in the database.")
	scanCmd.Flags().StringVarP(&scanHashAlgo, "algo", "a", "sha256", "Hashing algorithm to use. Valid values are 'sha1', 'sha256' and 'sha512'.")
	scanCmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Display progress information.")

	addPathFilteringFlags(scanCmd)
}

var (
	scanForceOverride   bool
	scanCalculateHashes bool
	scanHashAlgo        string
	scanDryRun          bool
)

// Determine the hashing algorithm to use based on the flag that was passed.
func algoFromFlag(flag string) (ajhash.Algo, error) {
	switch strings.ToLower(flag) {
	case "sha1":
		return ajhash.AlgoSHA1, nil
	case "sha256":
		return ajhash.AlgoSHA256, nil
	case "sha512":
		return ajhash.AlgoSHA512, nil
	}

	return ajhash.DefaultAlgo, fmt.Errorf("invalid hashing algorithm '%s'", flag)
}
