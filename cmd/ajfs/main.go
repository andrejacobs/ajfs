// Package main is the main entry point for the ajfs CLI
package main

import (
	"github.com/andrejacobs/ajfs/cmd/ajfs/commands"
)

// Main entry point for ajfs CLI
func main() {
	commands.Execute()
}
