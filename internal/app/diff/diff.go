package diff

import (
	"errors"
	"fmt"
	"strings"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-collection/collection"
)

// LHS = Left Hand Side (a)
// RHS = Right Hand Side (b)

// Config for the ajfs diff command.
type Config struct {
	config.CommonConfig

	RhsPath string
}

// Process the ajfs diff command.
func Run(cfg Config) error {
	return fmt.Errorf("TODO")
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
	Type    Type
	Id      path.Id
	Path    string
	IsDir   bool
	Changed ChangedFlags
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
		err = compare(lhs, rhs, onlyLHS, fn)
		if err != nil {
			if err != SkipAll {
				return err
			}
			return nil
		}
	}

	return nil
}

func compare(lhs *db.DatabaseFile, rhs *db.DatabaseFile, onlyLHS bool, fn CompareFn) error {
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
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func compareWithHashes(lhs *db.DatabaseFile, rhs *db.DatabaseFile, onlyLHS bool, fn CompareFn) error {
	return nil
}
