package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/spf13/cobra"
)

// ajfs scan
var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "TODO",
	Long:  `TODO`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := scan.Config{
			CommonConfig: commonConfig,
			Root:         args[0],
		}

		if err := scan.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
