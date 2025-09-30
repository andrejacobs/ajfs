package commands

import (
	"fmt"
	"os"

	"github.com/andrejacobs/go-aj/buildinfo"
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

func init() {
	versionTemplate := `{{printf "%s: %s\n" .Name .Version}}`
	rootCmd.SetVersionTemplate(versionTemplate)

	// Persistent flags that are available to every subcommand
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDBPath, "path to the database file.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "output verbose information.")
}

const (
	defaultDBPath = "./db.ajfs"
)

var (
	dbPath  string
	verbose bool
)
