package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/tosync"
	"github.com/spf13/cobra"
)

// ajfs tosync
var tosyncCmd = &cobra.Command{
	Use:   "tosync",
	Short: "Show which files need to be synced from the LHS to the RHS",
	Long:  `Show which files need to be synced from the LHS to the RHS`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := tosync.Config{
			CommonConfig: commonConfig,
			OnlyHashes:   tosyncHashesOnly,
		}

		switch len(args) {
		case 1:
			cfg.LhsPath = defaultDBPath
			cfg.RhsPath = args[0]
		case 2:
			cfg.LhsPath = args[0]
			cfg.RhsPath = args[1]
		}

		cfg.Fn = printDiff

		if err := tosync.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(tosyncCmd)

	tosyncCmd.Flags().BoolVarP(&tosyncHashesOnly, "hash", "s", false, "Compare only the file signature hashes")
}

var (
	tosyncHashesOnly bool
)
