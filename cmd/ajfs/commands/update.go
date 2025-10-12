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
	"github.com/andrejacobs/ajfs/internal/app/update"
	"github.com/spf13/cobra"
)

// ajfs update.
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Perform a new scan and update an existing database",
	Long:  `Perform a new scan and update an existing database`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filterCfg, err := parseFilterConfig()
		if err != nil {
			exitOnError(err, 1)
		}

		commonConfig.Progress = showProgress

		cfg := update.Config{
			CommonConfig: commonConfig,
			FilterConfig: *filterCfg,
		}
		cfg.DbPath = dbPathFromArgs(args)

		if err := update.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	updateCmd.Flags().BoolVarP(&showProgress, "progress", "p", false, "Display progress information")

	addPathFilteringFlags(updateCmd)
}
