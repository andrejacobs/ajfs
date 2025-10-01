package commands

import (
	"fmt"
	"strings"

	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/go-aj/ajhash"
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
			CommonConfig:  commonConfig,
			ForceOverride: scanForceOverride,
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

		if scanCalculateHashes {
			algo, err := algoFromFlag(scanHashAlgo)
			if err != nil {
				exitOnError(err, 1)
			}

			cfg.CalculateHashes = true
			cfg.Algo = algo
		}

		if err := scan.Run(cfg); err != nil {
			exitOnError(err, 1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().BoolVarP(&scanForceOverride, "force", "f", false, "Override any existing database")
	scanCmd.Flags().BoolVarP(&scanCalculateHashes, "hash", "s", false, "Calculate file signature hashes")
	scanCmd.Flags().StringVarP(&scanHashAlgo, "algo", "a", "sha256", "Hashing algorithm to use. Valid values are 'sha1', 'sha256' and 'sha512'.")
}

var (
	scanForceOverride   bool
	scanCalculateHashes bool
	scanHashAlgo        string
)

// Determine the hashing algorithm to use based on the flag that was passed
func algoFromFlag(flag string) (ajhash.Algo, error) {
	switch strings.ToLower(flag) {
	case "sha1":
		return ajhash.AlgoSHA1, nil
	case "sha256":
		return ajhash.AlgoSHA256, nil
	case "sha512":
		return ajhash.AlgoSHA512, nil
	}

	return ajhash.DefaultAlgo, fmt.Errorf("invalid hashing algorithm '%s'", flag)
}
