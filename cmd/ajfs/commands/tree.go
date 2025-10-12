package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/tree"
	"github.com/andrejacobs/go-aj/file"
	"github.com/spf13/cobra"
)

// ajfs tree
var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display the file hiearchy tree",
	Long:  `Display the file hiearchy tree`,
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := tree.Config{
			CommonConfig: commonConfig,
			OnlyDirs:     treeOnlyDirs,
			Limit:        treeLimit,
		}

		switch len(args) {
		case 0:
			cfg.DbPath = defaultDBPath
		case 1:
			exists, err := file.FileExists(args[0])
			if err != nil {
				exitOnError(err, 1)
			}

			if exists {
				cfg.DbPath = args[0]
			} else {
				cfg.DbPath = defaultDBPath
				cfg.Subpath = args[0]
			}
		case 2:
			cfg.DbPath = args[0]
			cfg.Subpath = args[1]
		default:
			panic("invalid args")
		}

		if err := tree.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)

	treeCmd.Flags().BoolVarP(&treeOnlyDirs, "dirs", "d", false, "Display only directories")
	treeCmd.Flags().IntVarP(&treeLimit, "limit", "l", 0, "Limit the tree depth")
}

var (
	treeOnlyDirs bool
	treeLimit    int
)
