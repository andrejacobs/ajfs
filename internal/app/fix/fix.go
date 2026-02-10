// Copyright (c) 2026 Andre Jacobs
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

// Package fix provides the functionality for ajfs fix command.
package fix

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
)

// Config for the ajfs fix command.
type Config struct {
	config.CommonConfig

	Stdin       io.Reader
	DryRun      bool   // Only display what needs to be fixed.
	RestorePath string // Path to a backup header to be restored.
}

// Process the ajfs fix command.
func Run(cfg Config) error {

	// Confirm with user
	if !cfg.DryRun {
		r := bufio.NewReader(cfg.Stdin)
		fmt.Fprintf(cfg.Stdout, "WARNING: Changes might be made to the database: %q\n", cfg.DbPath)
		fmt.Fprintf(cfg.Stdout, "Type 'yes' to confirm you want to continue: ")
		input, _ := r.ReadString('\n')

		if input != "yes\n" {
			return fmt.Errorf("user cancelled")
		}
	}

	// Restore?
	if cfg.RestorePath != "" {
		fmt.Fprintf(cfg.Stdout, "Restoring backup headers from: %q to database file: %q\n", cfg.RestorePath, cfg.DbPath)
		if err := db.RestoreDatabaseHeader(cfg.DbPath, cfg.RestorePath); err != nil {
			return err
		}
		return nil
	}

	// Fix
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	bakPath := filepath.Join(cwd, filepath.Base(cfg.DbPath)+".bak")

	if err := db.FixDatabase(cfg.Stdout, cfg.DbPath, cfg.DryRun, bakPath); err != nil {
		fmt.Fprintf(cfg.Stderr, "!! ERROR: %v\n", err)
		return err
	}

	return nil
}
