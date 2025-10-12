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
