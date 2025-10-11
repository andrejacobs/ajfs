package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/dupes"
	"github.com/spf13/cobra"
)

// ajfs dupes
var dupesCmd = &cobra.Command{
	Use:   "dupes",
	Short: "Display all duplicate files or directory trees",
	Long:  `Display all duplicate files or directory trees`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := dupes.Config{
			CommonConfig: commonConfig,
			Subtrees:     dupesDirs,
			PrintTree:    dupesDirsPrintTree,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := dupes.Run(cfg); err != nil {
			exitOnError(err, 1)
		}

	},
}

func init() {
	rootCmd.AddCommand(dupesCmd)

	dupesCmd.Flags().BoolVarP(&dupesDirs, "dirs", "d", false, "Display duplicate subtree directories")
	dupesCmd.Flags().BoolVarP(&dupesDirsPrintTree, "tree", "t", false, "Display the tree hierarchy of duplicate subtrees")
}

var (
	dupesDirs          = false
	dupesDirsPrintTree = false
)
