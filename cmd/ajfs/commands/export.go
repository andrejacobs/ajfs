package commands

import (
	"fmt"
	"strings"

	"github.com/andrejacobs/ajfs/internal/app/export"
	"github.com/spf13/cobra"
)

// ajfs export
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export the database",
	Long:  `Export the database`,
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := export.Config{
			CommonConfig: commonConfig,
		}

		switch len(args) {
		case 1:
			cfg.DbPath = defaultDBPath
			cfg.ExportPath = args[0]
		case 2:
			cfg.DbPath = args[0]
			cfg.ExportPath = args[1]
		default:
			panic("invalid args")
		}

		switch strings.ToLower(exportFormat) {
		case "csv":
			cfg.Format = export.FormatCSV
		case "json":
			cfg.Format = export.FormatJSON
		case "hashdeep":
			cfg.Format = export.FormatHashdeep
		default:
			exitOnError(fmt.Errorf("invalid export format %q", exportFormat), 1)
		}

		if err := export.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringVar(&exportFormat, "format", "csv", "Export format: csv, json or hashdeep.")
}

var (
	exportFormat string
)
