package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ajfs info
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about a database",
	Long:  `Display information about a database`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello world!")
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
