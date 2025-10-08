package commands

import (
	"fmt"

	"github.com/andrejacobs/ajfs/internal/app/diff"
	"github.com/spf13/cobra"
)

// ajfs diff
var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Display the differences between two databases and or file system",
	Long:  `Display the differences between two databases and or file system`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := diff.Config{
			CommonConfig: commonConfig,
		}

		switch len(args) {
		case 0:
			cfg.LhsPath = defaultDBPath
		case 1:
			cfg.LhsPath = args[0]
		case 2:
			cfg.LhsPath = args[0]
			cfg.RhsPath = args[1]
		}

		cfg.Fn = printDiff

		if err := diff.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)
}

func printDiff(d diff.Diff) error {
	if d.Type == diff.TypeNothing {
		return nil
	}

	fmt.Println(d.String())
	return nil
}
