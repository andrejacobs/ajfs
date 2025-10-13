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

// Package commands provide the subcommands for ajfs CLI.
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/go-aj/buildinfo"
	"github.com/andrejacobs/go-aj/stats"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:     "ajfs",
	Version: buildinfo.VersionString(),
	Short:   "Andre Jacobs' file hierarchy snapshot tool.",
	Long: `Andre Jacobs' file hierarchy snapshot tool is used to save a file system
hierarchy to a single flat file database.

Which can then be used in an offline and independant way to do the following:
* Find duplicate files or entire duplicate subtrees.
* Compare differences between databases (snapshots) and or file systems.
* Find out which files would still need to be synced to another system.
* Search for entries that match certain criteria.
* List or export the entries to CSV, JSON or Hashdeep.
* Display the entries as a tree.
`,
}

// Main entry point for ajfs CLI.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		os.Exit(1)
	}
}

// Init cobra.
func init() {
	cobra.OnInitialize(initApp)
	cobra.OnFinalize(cleanupApplication)

	versionTemplate := `{{printf "%s: %s\n" .Name .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)

	// Persistent flags that are available to every subcommand
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Display verbose information.")

	customHelp()
}

// Run before any commands are run.
func initApp() {
	commonConfig.Init()
	commonConfig.Verbose = verbose

	if commonConfig.Verbose {
		startTime = time.Now()
	}
}

// Run after a command is finished.
func cleanupApplication() {
	if commonConfig.Verbose {
		commonConfig.VerbosePrintln("")
		stats.PrintTimeTaken(commonConfig.Stdout, "ajfs", startTime, time.Now())
	}
}

// Log error message to STDERR and exit the program with the specified exit code.
func exitOnError(err error, code int) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(code)
}

// Database path from the args.
func dbPathFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return defaultDBPath
}

// Root cobra command.
func RootCmd() *cobra.Command {
	return rootCmd
}

func customHelp() {
	groups := []struct {
		Title    string
		Commands []string
	}{
		{
			Title:    "Creation commands",
			Commands: []string{"scan", "resume", "update"},
		},
		{
			Title:    "Information commands",
			Commands: []string{"info", "list", "export", "tree", "search"},
		},
		{
			Title:    "Comparison commands",
			Commands: []string{"diff", "tosync", "dupes"},
		},
	}

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Long)
		fmt.Printf("Usage:\n  %s [command]\n\n", cmd.UseLine())

		fmt.Println("Available commands:")
		cmds := cmd.Commands()
		cmdMap := make(map[string]*cobra.Command)
		for _, c := range cmds {
			cmdMap[c.Name()] = c
		}

		for _, group := range groups {
			fmt.Printf("  %s:\n", group.Title)
			for _, name := range group.Commands {
				if c, ok := cmdMap[name]; ok {
					fmt.Printf("    %-12s %s\n", c.Name(), c.Short)
				}
			}
			fmt.Println()
		}

		fmt.Println("Use \"ajfs [command] --help\" for more information about a command.")
	})
}

const (
	defaultDBPath = "./db.ajfs"
)

var (
	verbose      bool
	showProgress bool

	commonConfig config.CommonConfig

	startTime time.Time
)
