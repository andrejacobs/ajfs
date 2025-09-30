package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/info"
	"github.com/spf13/cobra"
)

// ajfs info
var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about a database",
	Long:  `Display information about a database`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := info.Config{
			CommonConfig: commonConfig,
		}

		if err := info.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
