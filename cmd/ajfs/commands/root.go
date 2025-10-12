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

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "ajfs",
	Version: buildinfo.VersionString(),
	Short:   "todo",
	Long:    `todo`,
}

// Main entry point for ajfs CLI
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stdout, "%v\n", err)
		os.Exit(1)
	}
}

// Init cobra
func init() {
	cobra.OnInitialize(initApp)
	cobra.OnFinalize(cleanupApplication)

	versionTemplate := `{{printf "%s: %s\n" .Name .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)

	// Persistent flags that are available to every subcommand
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Output verbose information.")
}

// Run before any commands are run
func initApp() {
	commonConfig.Init()
	commonConfig.Verbose = verbose

	if commonConfig.Verbose {
		startTime = time.Now()
	}
}

// Run after a command is finished
func cleanupApplication() {
	if commonConfig.Verbose {
		commonConfig.VerbosePrintln("")
		stats.PrintTimeTaken(commonConfig.Stdout, "ajfs", startTime, time.Now())
	}
}

// Log error message to STDERR and exit the program with the specified exit code
func exitOnError(err error, code int) {
	fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	os.Exit(code)
}

// Database path from the args
func dbPathFromArgs(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return defaultDBPath
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
