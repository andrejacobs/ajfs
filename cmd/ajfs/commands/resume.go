package commands

import (
	"github.com/andrejacobs/ajfs/internal/app/resume"
	"github.com/spf13/cobra"
)

// ajfs resume
var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Resume calculating file signature hashes",
	Long:  `Resume calculating file signature hashes`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		commonConfig.Progress = showProgress

		cfg := resume.Config{
			CommonConfig: commonConfig,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := resume.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(resumeCmd)

	resumeCmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Display progress information")
}
