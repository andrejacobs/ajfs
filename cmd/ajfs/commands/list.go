package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/list"
	"github.com/spf13/cobra"
)

// ajfs list
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Display the database path entries",
	Long:  `Display the database path entries`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := list.Config{
			CommonConfig:     commonConfig,
			DisplayFullPaths: listDisplayFullPaths,
			DisplayHashes:    listDisplayHashes,
			DisplayMinimal:   !listDisplayMore,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := list.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().BoolVarP(&listDisplayFullPaths, "full", "f", false, "Display full paths for entries")
	listCmd.Flags().BoolVarP(&listDisplayHashes, "hash", "s", false, "Display file signature hashes if available")
	listCmd.Flags().BoolVarP(&listDisplayMore, "more", "m", false, "Display more information about the paths")
}

var (
	listDisplayFullPaths bool
	listDisplayHashes    bool
	listDisplayMore      bool
)
