package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/search"
	"github.com/spf13/cobra"
)

// ajfs search
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search for matching path entries",
	Long:  `Search for matching path entries`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := search.Config{
			CommonConfig: commonConfig,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := search.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
