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

	cfg.VerbosePrintln("Checking differences ...")
	err = Compare(cfg.LhsPath, cfg.RhsPath, false, cfg.Fn)
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
			sb.WriteString("m") // Mode changed (type  and permissions)
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

//-----------------------------------------------------------------------------

// Indicates to Compare to stop processing differences.
var SkipAll = errors.New("skip all") //lint:ignore ST1012 not an error and is more readable

// Called by Compare for each difference that was found.
// Return [SkipAll] to stop the process.
type CompareFn func(d Diff) error

// Compare the differences between two ajfs database files.
// fn Will be called for each difference that is found.
// If fn returns [SkipAll] then the process will be stopped and nil will be returned as the error.
func Compare(lhsPath string, rhsPath string, onlyLHS bool, fn CompareFn) error {
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

	if lhs.Features().HasHashTable() && rhs.Features().HasHashTable() {
		err = compareWithHashes(lhs, rhs, onlyLHS, fn)
		if err != nil {
			if err != SkipAll {
				return err
			}
			return nil
		}
	} else {
		err = CompareDatabases(lhs, rhs, onlyLHS, fn)
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
