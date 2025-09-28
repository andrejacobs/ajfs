// Package scan provides the functionality for ajfs scan command.
package scan

import "github.com/andrejacobs/ajfs/internal/app/config"

// Config for the ajfs scan command.
type Config struct {
	config.CommonConfig
}

func Run(cfg Config) error {
	cfg.Println("Hello World!")
	cfg.VerbosePrintln("Some more verbose output")

	// TODO: Safe shutdown, cancel contex etc.
	return nil
}
