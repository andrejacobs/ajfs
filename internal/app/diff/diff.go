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

// Package diff provides the functionality for ajfs diff command.
package diff

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/file"
	"github.com/andrejacobs/go-collection/collection"
)

// LHS = Left Hand Side (a)
// RHS = Right Hand Side (b)

// Config for the ajfs diff command.
type Config struct {
	config.CommonConfig

	LhsPath string
	RhsPath string

	IncludeFilters []FilterFlags
	ExcludeFilters []FilterFlags

	Fn CompareFn
}

// Process the ajfs diff command.
func Run(cfg Config) error {
	if cfg.Fn == nil {
		panic("expected a compare function")
	}

	lhsExists, err := file.FileExists(cfg.LhsPath)
	if err != nil {
		return err
	}
	if !lhsExists {
		cfg.VerbosePrintln(fmt.Sprintf("Creating temporary database for LHS: %q", cfg.LhsPath))
		dbPath, err := makeTempDatabase(cfg, cfg.LhsPath)
		if err != nil {
			return fmt.Errorf("failed to create temporary database for left hand side. %w", err)
		}
		cfg.LhsPath = dbPath
		defer os.Remove(dbPath)
	}

	if cfg.RhsPath == "" {
		lhs, err := db.OpenDatabase(cfg.LhsPath)
		if err != nil {
			return fmt.Errorf("failed to open the left hand side database %q. %w", cfg.LhsPath, err)
		}
		cfg.RhsPath = lhs.RootPath()
		lhs.Close()
	}

	rhsExists, err := file.FileExists(cfg.RhsPath)
	if err != nil {
		return err
	}
	if !rhsExists {
		cfg.VerbosePrintln(fmt.Sprintf("Creating temporary database for RHS: %q", cfg.RhsPath))
		dbPath, err := makeTempDatabase(cfg, cfg.RhsPath)
		if err != nil {
			return fmt.Errorf("failed to create temporary database for right hand side. %w", err)
		}
		cfg.RhsPath = dbPath
		defer os.Remove(dbPath)
	}

	if cfg.IncludeFilters == nil {
		cfg.IncludeFilters = []FilterFlags{}
	}
	if cfg.ExcludeFilters == nil {
		cfg.ExcludeFilters = []FilterFlags{}
	}

	cfg.VerbosePrintln("Checking differences ...")
	err = Compare(cfg.LhsPath, cfg.RhsPath, cfg.IncludeFilters, cfg.ExcludeFilters, cfg.Fn)
	if err != nil {
		return err
	}

	return nil
}

//-----------------------------------------------------------------------------

// Describe the type of difference.
type Type int

const (
	TypeNothing        = 0        // Nothing changed
	TypeLeftOnly  Type = 1 + iota // Only on the LHS (same as having been removed from the RHS)
	TypeRightOnly                 // Only on the RHS (same as having been added in the RHS)
	TypeChanged                   // Some of the file's meta data or the hash have been changed
)

// Describe what has changed for an item that exists on both sides.
type ChangedFlags uint8

const (
	ChangedNothing = 0         // Nothing changed
	ChangedMode    = 1 << iota // The path's type and or permissions has changed
	ChangedSize                // The size has changed
	ChangedModTime             // The last modification time has changed
	ChangedHash                // The hash is different
)

func (f ChangedFlags) ModeChanged() bool {
	return (f & ChangedMode) != 0
}

func (f ChangedFlags) SizeChanged() bool {
	return (f & ChangedSize) != 0
}

func (f ChangedFlags) ModTimeChanged() bool {
	return (f & ChangedModTime) != 0
}

func (f ChangedFlags) HashChanged() bool {
	return (f & ChangedHash) != 0
}

func (f ChangedFlags) FilterFlagsMask() FilterFlags {
	var result FilterFlags = FilterNoOp

	if f == ChangedNothing {
		return result
	}

	if f.ModeChanged() {
		result |= FilterChangedMode
	}

	if f.SizeChanged() {
		result |= FilterChangedSize
	}

	if f.ModTimeChanged() {
		result |= FilterChangedModTime
	}

	if f.HashChanged() {
		result |= FilterChangedHash
	}

	return result
}

// Flags used to decide which items to be included or excluded in a diff.
type FilterFlags uint16

const (
	FilterNoOp           = 0         // Don't apply a filter
	FilterDirs           = 1 << iota // Directories
	FilterFiles                      // Files
	FilterTypeLeft                   // LHS only (mutually exclusive with FilterOnlyRight)
	FilterTypeRight                  // RHS only
	FilterTypeChanged                // Both sides but has changes
	FilterChangedMode                // The path's type and or permissions has changed
	FilterChangedSize                // The size has changed
	FilterChangedModTime             // The last modification time has changed
	FilterChangedHash                // The hash is different

	FilterChangedMask = FilterChangedMode | FilterChangedSize | FilterChangedModTime | FilterChangedHash
)

func (f FilterFlags) Validate() error {
	if (f&FilterTypeLeft != 0) && (f&FilterTypeRight != 0) {
		return fmt.Errorf("filtering on left hand side only or right hand side only is mutually exclusive. Use FilterTypeChanged (~) instead")
	}

	if (f&FilterTypeLeft != 0) && (f&FilterChangedMask != 0) {
		return fmt.Errorf("can't filter on left hand side only and changes")
	}

	if (f&FilterTypeRight != 0) && (f&FilterChangedMask != 0) {
		return fmt.Errorf("can't filter on right hand side only and changes")
	}

	return nil
}

func (f FilterFlags) ChangedFlagsMask() ChangedFlags {
	var result ChangedFlags = ChangedNothing

	if f&FilterChangedMode != 0 {
		result |= ChangedMode
	}

	if f&FilterChangedSize != 0 {
		result |= ChangedSize
	}

	if f&FilterChangedModTime != 0 {
		result |= ChangedModTime
	}

	if f&FilterChangedHash != 0 {
		result |= ChangedHash
	}

	return result
}

func (f FilterFlags) String() string {
	sb := strings.Builder{}

	if f&FilterTypeLeft != 0 {
		sb.WriteRune('-')
	} else if f&FilterTypeRight != 0 {
		sb.WriteRune('+')
	} else if f&FilterTypeChanged != 0 {
		sb.WriteRune('~')
	}

	if f&FilterDirs != 0 {
		sb.WriteRune('d')
	}

	if f&FilterFiles != 0 {
		sb.WriteRune('f')
	}

	if f&FilterChangedMode != 0 {
		sb.WriteRune('m')
	}

	if f&FilterChangedSize != 0 {
		sb.WriteRune('s')
	}

	if f&FilterChangedModTime != 0 {
		sb.WriteRune('l')
	}

	if f&FilterChangedHash != 0 {
		sb.WriteRune('x')
	}

	return sb.String()
}

func ParseFilterFlags(input string) (FilterFlags, error) {
	var result FilterFlags = FilterNoOp
	for _, c := range strings.ToLower(input) {
		switch c {
		case '-':
			result |= FilterTypeLeft
		case '+':
			result |= FilterTypeRight
		case '~':
			result |= FilterTypeChanged
		case 'd':
			result |= FilterDirs
		case 'f':
			result |= FilterFiles
		case 'm':
			result |= FilterChangedMode
		case 's':
			result |= FilterChangedSize
		case 'l':
			result |= FilterChangedModTime
		case 'x':
			result |= FilterChangedHash
		default:
			return 0, fmt.Errorf("invalid filter: %s. unknown filter property: %c", input, c)
		}
	}

	return result, nil
}

func ParseFilterFlagsArray(input []string) ([]FilterFlags, error) {
	result := make([]FilterFlags, 0, len(input))

	for _, elem := range input {
		if len(elem) > 0 {
			f, err := ParseFilterFlags(elem)
			if err != nil {
				return nil, err
			}
			result = append(result, f)
		}
	}

	return result, nil
}

// Describe a difference between the LHS and RHS databases.
type Diff struct {
	Type    Type         // Type of difference
	Id      path.Id      // Identifier of the path info item
	Path    string       // Path of the item
	IsDir   bool         // Is this a directory
	Changed ChangedFlags // What was changed
	Size    uint64       // Size of the item. If the item exists on both sides, then this would be the size of the LHS item
}

// Stringer implementation.
func (d *Diff) String() string {
	var typeChar rune
	if d.IsDir {
		typeChar = 'd'
	} else {
		typeChar = 'f'
	}

	switch d.Type {
	case TypeLeftOnly:
		return fmt.Sprintf("%c---- %s", typeChar, d.Path)
	case TypeRightOnly:
		return fmt.Sprintf("%c++++ %s", typeChar, d.Path)
	case TypeChanged:
		// Mode, Size, ModTime
		sb := strings.Builder{}
		sb.WriteRune(typeChar)
		if d.Changed.ModeChanged() {
			sb.WriteString("m") // Mode changed (type and permissions)
		} else {
			sb.WriteString("~")
		}
		if d.Changed.SizeChanged() {
			sb.WriteString("s") // Size changed
		} else {
			sb.WriteString("~")
		}
		if d.Changed.ModTimeChanged() {
			sb.WriteString("l") // Last modification time changed
		} else {
			sb.WriteString("~") // Data unchanged
		}
		if d.Changed.HashChanged() {
			sb.WriteString("x") // Hash has changed
		} else {
			sb.WriteString("~") // Data unchanged
		}
		return fmt.Sprintf("%s %s", sb.String(), d.Path)
	default:
		return ""
	}
}

func (d Diff) FilterFlagsMask() FilterFlags {
	var result FilterFlags = FilterNoOp

	switch d.Type {
	case TypeLeftOnly:
		result |= FilterTypeLeft
	case TypeRightOnly:
		result |= FilterTypeRight
	case TypeChanged:
		result |= FilterTypeChanged
	}

	if d.IsDir {
		result |= FilterDirs
	} else {
		result |= FilterFiles
	}

	result |= d.Changed.FilterFlagsMask()

	return result
}

//-----------------------------------------------------------------------------

// Indicates to Compare to stop processing differences.
var SkipAll = errors.New("skip all") //nolint:staticcheck //ST1012: not an error and is more readable

// Called by Compare for each difference that was found.
// Return [SkipAll] to stop the process.
type CompareFn func(d Diff) error

// Compare the differences between two ajfs database files.
// fn Will be called for each difference that is found.
// If fn returns [SkipAll] then the process will be stopped and nil will be returned as the error.
func Compare(lhsPath string, rhsPath string,
	includeFilters []FilterFlags, excludeFilters []FilterFlags,
	fn CompareFn) error {

	for _, f := range includeFilters {
		if err := f.Validate(); err != nil {
			return fmt.Errorf("invalid include filter. %w", err)
		}
	}

	for _, f := range excludeFilters {
		if err := f.Validate(); err != nil {
			return fmt.Errorf("invalid exclude filter. %w", err)
		}
	}

	lhs, err := db.OpenDatabase(lhsPath)
	if err != nil {
		return fmt.Errorf("failed to open left hand side database. %w", err)
	}
	defer lhs.Close()

	rhs, err := db.OpenDatabase(rhsPath)
	if err != nil {
		return fmt.Errorf("failed to open right hand side database. %w", err)
	}
	defer rhs.Close()

	var compFn = fn

	hasIncludeFilters := len(includeFilters) > 0
	hasExcludeFilters := len(excludeFilters) > 0

	if hasIncludeFilters || hasExcludeFilters {
		compFn = func(d Diff) error {
			if d.Type == TypeNothing {
				return nil
			}

			keep := !hasIncludeFilters

			// Include filter
			for _, f := range includeFilters {
				if f != FilterNoOp && d.FilterFlagsMask()&f == f {
					keep = true
					break
				}
			}

			// Exclude filter
			if keep {
				for _, f := range excludeFilters {
					if f != FilterNoOp && (d.FilterFlagsMask()&f == f) {
						keep = false
						break
					}
				}
			}

			if keep {
				return fn(d)
			}
			return nil
		}
	}

	onlyLHS := false

	if lhs.Features().HasHashTable() && rhs.Features().HasHashTable() {
		err = compareWithHashes(lhs, rhs, onlyLHS, compFn)
		if err != nil {
			if err != SkipAll {
				return err
			}
			return nil
		}
	} else {
		err = CompareDatabases(lhs, rhs, onlyLHS, compFn)
		if err != nil {
			if err != SkipAll {
				return err
			}
			return nil
		}
	}

	return nil
}

func CompareDatabases(lhs *db.DatabaseFile, rhs *db.DatabaseFile, onlyLHS bool, fn CompareFn) error {
	lhsMap, err := lhs.BuildIdToInfoMap()
	if err != nil {
		return fmt.Errorf("left hand side error. %w", err)
	}

	rhsMap, err := rhs.BuildIdToInfoMap()
	if err != nil {
		return fmt.Errorf("right hand side error. %w", err)
	}

	lessFn := func(lhs path.Info, rhs path.Info) bool {
		return lhs.Path < rhs.Path
	}

	// What exists only on the LHS (removed from RHS)
	lhsOnly := collection.MapDifference(lhsMap, rhsMap)
	sortedLhsOnly := collection.MapSortedByValueFunc(lhsOnly, lessFn)

	for _, kv := range sortedLhsOnly {
		err = fn(Diff{
			Type:  TypeLeftOnly,
			Id:    kv.Value.Id,
			Path:  kv.Value.Path,
			IsDir: kv.Value.IsDir(),
			Size:  kv.Value.Size,
		})
		if err != nil {
			return err
		}
		// fmt.Printf("- %x : %s\n", kv.Key, kv.Value.Path)
	}

	if !onlyLHS {
		// What exists only on the RHS (added on the LHS)
		rhsOnly := collection.MapDifference(rhsMap, lhsMap)
		sortedRhsOnly := collection.MapSortedByValueFunc(rhsOnly, lessFn)

		for _, kv := range sortedRhsOnly {
			err = fn(Diff{
				Type:  TypeRightOnly,
				Id:    kv.Value.Id,
				Path:  kv.Value.Path,
				IsDir: kv.Value.IsDir(),
				Size:  kv.Value.Size,
			})
			if err != nil {
				return err
			}
			// fmt.Printf("+ %x : %s\n", kv.Key, kv.Value.Path)
		}
	}

	// What exists in both
	both := collection.MapIntersection(lhsMap, rhsMap)
	for k := range both {
		lv := lhsMap[k]
		rv := rhsMap[k]

		// Check what has changed
		var changed ChangedFlags
		if lv.Mode != rv.Mode {
			changed |= ChangedMode
		}
		if lv.Size != rv.Size {
			changed |= ChangedSize
		}
		if lv.ModTime != rv.ModTime {
			changed |= ChangedModTime
		}

		var diffType Type
		if changed != 0 {
			diffType = TypeChanged
		} else {
			diffType = TypeNothing
		}

		err = fn(Diff{
			Type:    diffType,
			Id:      lv.Id,
			Path:    lv.Path,
			Changed: changed,
			IsDir:   lv.IsDir(),
			Size:    lv.Size,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func compareWithHashes(lhs *db.DatabaseFile, rhs *db.DatabaseFile, onlyLHS bool, fn CompareFn) error {
	lhsAlgo, err := lhs.HashTableAlgo()
	if err != nil {
		return fmt.Errorf("failed to get the left hand side hashing algorithm. %w", err)
	}

	rhsAlgo, err := rhs.HashTableAlgo()
	if err != nil {
		return fmt.Errorf("failed to get the right hand side hashing algorithm. %w", err)
	}

	if lhsAlgo != rhsAlgo {
		// Can't compare hashes so just do normal compare
		return CompareDatabases(lhs, rhs, onlyLHS, fn)
	}

	lhsMap, err := lhs.BuildIdToHashMap()
	if err != nil {
		return fmt.Errorf("failed to build the left hand side hash map. %w", err)
	}

	rhsMap, err := rhs.BuildIdToHashMap()
	if err != nil {
		return fmt.Errorf("failed to build the right hand side hash map. %w", err)
	}

	err = CompareDatabases(lhs, rhs, onlyLHS, func(d Diff) error {
		// Check if the hashes are different if this diff is for a file (!dir)
		// and the diff thus far indicates nothing or meta has changed
		if !d.IsDir && ((d.Type == TypeNothing) || (d.Type == TypeChanged)) {
			lhsHash, lExists := lhsMap[d.Id]
			rhsHash, rExists := rhsMap[d.Id]

			if (lExists && rExists) && !slices.Equal(lhsHash, rhsHash) {
				d.Type = TypeChanged
				d.Changed |= ChangedHash
			}
		}
		return fn(d)
	})
	if err != nil {
		return err
	}

	return nil
}

// Create a temporary database by scanning the path.
// Returns the path of the temporary database.
func makeTempDatabase(cfg Config, path string) (string, error) {
	dbPath := filepath.Join(os.TempDir(), filepath.Base(path)+".ajfs")

	scanCfg := scan.Config{
		CommonConfig: cfg.CommonConfig,
		Root:         path,
	}
	scanCfg.DbPath = dbPath
	scanCfg.ForceOverride = true

	err := scan.Run(scanCfg)
	if err != nil {
		return "", err
	}

	return dbPath, nil
}

//-----------------------------------------------------------------------------

// DiffStats can be used to get some statistics out of a diff.
type DiffStats struct {
	LeftOnly   int // Count of left hand side only items
	RightOnly  int // Count of right hand side only items
	Changed    int // Count of changed items
	NotChanged int // Count of items that exist in both sides and that is unchanged

	Files int // Count of files
	Dirs  int // Count of directories

	ModeChanged    int // Count of items where the mode has changed
	SizeChanged    int // Count of items where the size has changed
	ModTimeChanged int // Count of items where the last modification time changed
	HashChanged    int // Count of items where the hash has changed

	Fn CompareFn // The compare function to be called
}

// Compare function that will update the stats.
func (ds *DiffStats) Compare(d Diff) error {
	if d.Type == TypeNothing {
		ds.NotChanged++
	} else {
		flags := d.FilterFlagsMask()

		if flags&FilterTypeLeft != 0 {
			ds.LeftOnly++
		} else if flags&FilterTypeRight != 0 {
			ds.RightOnly++
		} else {
			ds.Changed++
		}

		if flags&FilterFiles != 0 {
			ds.Files++
		} else if flags&FilterDirs != 0 {
			ds.Dirs++
		}

		if flags&FilterChangedMode != 0 {
			ds.ModeChanged++
		}

		if flags&FilterChangedSize != 0 {
			ds.SizeChanged++
		}

		if flags&FilterChangedModTime != 0 {
			ds.ModTimeChanged++
		}

		if flags&FilterChangedHash != 0 {
			ds.HashChanged++
		}
	}

	return ds.Fn(d)
}
