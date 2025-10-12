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

// If Progress is enabled then output to Stdout else output using VerbosePrintln.
func (c *CommonConfig) ProgressPrintln(a ...any) {
	if c.Progress {
		c.Println(a...)
	} else {
		c.VerbosePrintln(a...)
	}
}

//-----------------------------------------------------------------------------

// Config used to filter paths.
type FilterConfig struct {
	DirIncluder  file.MatchPathFn // Determine which directories should be walked
	FileIncluder file.MatchPathFn // Determine which files should be walked

	DirExcluder  file.MatchPathFn // Determine which directories should not be walked
	FileExcluder file.MatchPathFn // Determine which files should not be walked
}
