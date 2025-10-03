// Package config provides commonly shared configuration for the ajfs commands.
package config

import (
	"fmt"
	"io"
	"os"

	"github.com/andrejacobs/go-aj/file"
)

// Config used by most of the ajfs commands.
type CommonConfig struct {
	DbPath   string // Path to the database file.
	Verbose  bool   // Output verbose information to Stdout.
	Progress bool   // Output progression information to Stdout.

	Stdout io.Writer // Writer used for standard out
	Stderr io.Writer // Writer used for standard error
}

// Initialize with defaults.
func (c *CommonConfig) Init() {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
}

// Write output to Stdout.
func (c *CommonConfig) Println(a ...any) {
	fmt.Fprintln(c.Stdout, a...)
}

// Write output to Stdout only if verbose is enabled.
func (c *CommonConfig) VerbosePrintln(a ...any) {
	if c.Verbose {
		fmt.Fprintln(c.Stdout, a...)
	}
}

// Write output to Stderr.
func (c *CommonConfig) Errorln(a ...any) {
	fmt.Fprintln(c.Stderr, a...)
}

//-----------------------------------------------------------------------------

// Config used to filter paths.
type FilterConfig struct {
	DirIncluder  file.MatchPathFn // Determine which directories should be walked
	FileIncluder file.MatchPathFn // Determine which files should be walked

	DirExcluder  file.MatchPathFn // Determine which directories should not be walked
	FileExcluder file.MatchPathFn // Determine which files should not be walked
}
