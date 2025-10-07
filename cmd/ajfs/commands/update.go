package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/update"
	"github.com/spf13/cobra"
)

// ajfs update
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Perform a new scan and update an existing database",
	Long:  `Perform a new scan and update an existing database`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterCfg, err := parseFilterConfig()
		if err != nil {
			exitOnError(err, 1)
		}

		commonConfig.Progress = showProgress

		cfg := update.Config{
			CommonConfig: commonConfig,
			FilterConfig: *filterCfg,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := update.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Display progress information")

	addPathFilteringFlags(updateCmd)
}
