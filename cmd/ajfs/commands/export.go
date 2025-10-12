// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

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
			FullPaths:    exportFullPaths,
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
	exportCmd.Flags().BoolVarP(&exportFullPaths, "full", "f", false, "Export full paths for entries")
}

var (
	exportFormat    string
	exportFullPaths bool
)
