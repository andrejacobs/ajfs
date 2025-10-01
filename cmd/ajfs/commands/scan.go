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
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := scan.Config{
			CommonConfig: commonConfig,
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

		if err := scan.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)
}
