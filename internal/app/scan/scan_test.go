package scan_test

import (
	"os"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
)

func Test(t *testing.T) {
	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Verbose: true,
		},
	}
	scan.Run(cfg)
}
