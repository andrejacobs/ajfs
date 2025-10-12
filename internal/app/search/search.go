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

// Package search provides the functionality for ajfs search command.
package search

import (
	"encoding/hex"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
)

// Config for the ajfs info command.
type Config struct {
	config.CommonConfig
	Expresion        Expression // The search expression used to match path entries against.
	AlsoHashes       bool       // If the hashes need to also be checked, because we know one of the expressions require this.
	DisplayFullPaths bool       // If true then each path entry will be prefixed with the root path of the database.
	DisplayMinimal   bool       // Display only the paths.
}

// Process the ajfs info command.
func Run(cfg Config) error {

	if cfg.Expresion == nil {
		return fmt.Errorf("expected a search expression")
	}

	dbf, err := db.OpenDatabase(cfg.DbPath)
	if err != nil {
		return err
	}
	defer dbf.Close()

	// Header
	if cfg.CommonConfig.Verbose {
		if cfg.AlsoHashes && dbf.Features().HasHashTable() {
			if cfg.DisplayMinimal {
				cfg.Println("Hash, Path")
			} else {
				cfg.Println(path.HeaderWithHash())
			}
		} else {
			if cfg.DisplayMinimal {
				cfg.Println("Path")
			} else {
				cfg.Println(path.Header())
			}
		}
	}

	// Hashes?
	if cfg.AlsoHashes && dbf.Features().HasHashTable() {
		err = dbf.ReadAllEntriesWithHashes(func(idx int, pi path.Info, hash []byte) error {
			matched, err := cfg.Expresion.Match(pi, hash)
			if err != nil {
				return err
			}

			if !matched {
				return nil
			}

			if cfg.DisplayFullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			hashStr := hex.EncodeToString(hash)

			if cfg.DisplayMinimal {
				cfg.Println(fmt.Sprintf("%s, %q", hashStr, pi.Path))
			} else {
				cfg.Println(fmt.Sprintf("{%x}, %s, %v, %q, %v, %v", pi.Id, hashStr, pi.Size, pi.Path, pi.Mode, pi.ModTime.Format(time.RFC3339Nano)))
			}
			return nil
		})
		return err
	} else {
		// Without hashes
		err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
			matched, err := cfg.Expresion.Match(pi, nil)
			if err != nil {
				return err
			}

			if !matched {
				return nil
			}

			if cfg.DisplayFullPaths {
				pi.Path = filepath.Join(dbf.RootPath(), pi.Path)
			}

			if cfg.DisplayMinimal {
				cfg.Println(pi.Path)
			} else {
				cfg.Println(pi)
			}
			return nil
		})
		return err
	}
}

//-----------------------------------------------------------------------------

// Expression is used to form an expression that will be used to see if a path entry matches.
type Expression interface {
	// Match checks if the path info matches and returns true if it does.
	// The hash is optional and will be nil if that database does not have a file signature hash for the path.
	Match(pi path.Info, hash []byte) (bool, error)
}

// A matching path entry.
type Result struct {
	Info path.Info // The path entry that was matched.
	Hash []byte    // [optional] The file signature hash.
}

//-----------------------------------------------------------------------------
// AND

type searchAnd struct {
	lhs Expression
	rhs Expression
}

// Matches if both expressions matches.
func NewAnd(lhs Expression, rhs Expression) *searchAnd {
	return &searchAnd{lhs: lhs, rhs: rhs}
}

func (s *searchAnd) Match(pi path.Info, hash []byte) (bool, error) {
	lhsMatch, err := s.lhs.Match(pi, hash)
	if err != nil {
		return lhsMatch, err
	}

	rhsMatch, err := s.rhs.Match(pi, hash)
	if err != nil {
		return rhsMatch, err
	}

	return lhsMatch && rhsMatch, nil
}

//-----------------------------------------------------------------------------
// OR

type searchOr struct {
	lhs Expression
	rhs Expression
}

// Matches if any expressions matches.
func NewOr(lhs Expression, rhs Expression) *searchOr {
	return &searchOr{lhs: lhs, rhs: rhs}
}

func (s *searchOr) Match(pi path.Info, hash []byte) (bool, error) {
	lhsMatch, err := s.lhs.Match(pi, hash)
	if err != nil {
		return lhsMatch, err
	}

	if lhsMatch {
		return true, nil
	}

	rhsMatch, err := s.rhs.Match(pi, hash)
	if err != nil {
		return rhsMatch, err
	}

	return rhsMatch, nil
}

//-----------------------------------------------------------------------------
// NOT

type searchNot struct {
	exp Expression
}

// Inverts a match
func NewNot(exp Expression) *searchNot {
	return &searchNot{exp: exp}
}

func (s *searchNot) Match(pi path.Info, hash []byte) (bool, error) {
	match, err := s.exp.Match(pi, hash)
	return !match, err
}

//-----------------------------------------------------------------------------
// Func

// Call back function to check if a path matches.
type MatchFn func(pi path.Info, hash []byte) (bool, error)

type searchFunc struct {
	fn MatchFn
}

// Call a user provided funcion to check if a path matches.
func NewFunc(fn MatchFn) *searchFunc {
	return &searchFunc{fn: fn}
}

func (s *searchFunc) Match(pi path.Info, hash []byte) (bool, error) {
	return s.fn(pi, hash)
}

//-----------------------------------------------------------------------------
// Always

// Always matches.
type Always struct {
}

func (s *Always) Match(pi path.Info, hash []byte) (bool, error) {
	return true, nil
}

//-----------------------------------------------------------------------------
// Never

// Never matches.
type Never struct {
}

func (s *Never) Match(pi path.Info, hash []byte) (bool, error) {
	return false, nil
}

//-----------------------------------------------------------------------------
// Regex

type searchRegex struct {
	regex *regexp.Regexp
}

// Match a path against a regular expression.
func NewRegex(expression string) (*searchRegex, error) {
	s := &searchRegex{}
	var err error
	s.regex, err = regexp.Compile(expression)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *searchRegex) Match(pi path.Info, hash []byte) (bool, error) {
	if s.regex == nil {
		return false, nil
	}

	return s.regex.MatchString(pi.Path), nil
}

//-----------------------------------------------------------------------------
// Shell pattern

type searchShellPattern struct {
	pattern     string
	baseOnly    bool
	insensitive bool
}

// Match path against a shell pattern.
// pattern As supported by [filepath.Match].
// baseOnly If true then only the  last part of the path will be checked.
// insensitive If true then a case insensitive compare will be done.
func NewShellPattern(pattern string, baseOnly bool, insensitive bool) (*searchShellPattern, error) {
	if insensitive {
		pattern = strings.ToLower(pattern)
	}
	// Validate pattern
	_, err := filepath.Match(pattern, "")
	if err != nil {
		return nil, err
	}

	s := &searchShellPattern{
		pattern:     pattern,
		baseOnly:    baseOnly,
		insensitive: insensitive,
	}
	return s, nil
}

func (s *searchShellPattern) Match(pi path.Info, hash []byte) (bool, error) {
	path := pi.Path
	if s.baseOnly {
		path = filepath.Base(path)
	}
	if s.insensitive {
		path = strings.ToLower(path)
	}

	matched, err := filepath.Match(s.pattern, path)
	if err != nil {
		return false, fmt.Errorf("failed to match the shell pattern %q. %v", s.pattern, err)
	}

	return matched, nil
}

//-----------------------------------------------------------------------------
// Type

type searchType struct {
	flags    fs.FileMode
	andFiles bool
}

// Match if path is of a specific type as compared against the fs.FileMode flags.
func NewTypeFlags(flags fs.FileMode) *searchType {
	s := &searchType{flags: flags}
	return s
}

// Match if path is of a specific type.
// t Is a string containing a combination of [d, f, l, p, s] to form a matching flag set.
// d: Directory
// f: Regular file
// l: Symbolic link
// p: Named pipe
// s: Socket
func NewType(t string) (*searchType, error) {
	s := &searchType{}

	for _, c := range strings.ToLower(t) {
		switch c {
		case 'd':
			s.flags |= fs.ModeDir
		case 'f':
			s.andFiles = true
		case 'l':
			s.flags |= fs.ModeSymlink
		case 'p':
			s.flags |= fs.ModeNamedPipe
		case 's':
			s.flags |= fs.ModeSocket
		default:
			return nil, fmt.Errorf("unknown type %q in %q", c, t)
		}
	}

	return s, nil
}

func (s *searchType) Match(pi path.Info, hash []byte) (bool, error) {
	if s.andFiles || (s.flags == 0) { // Regular file
		return pi.Mode.IsRegular(), nil
	}
	return (pi.Mode & s.flags) != 0, nil
}

//-----------------------------------------------------------------------------
// Size

type searchSize struct {
	size uint64
	op   searchSizeOp
}

type searchSizeOp int

const (
	searchSizeOpEqual searchSizeOp = iota
	searchSizeOpLess
	searchSizeOpGreater
)

// Match path based on a file size expression.
// NOTE: This does not work on the BLOCKSIZE and rouding up concept used by find.
// Expresion can be in the format of: [+/-]<n>[suffix]
// No suffix means exactly n bytes.
// Valid suffixes are:
// k/K for Kilobytes (1 KB = 1000 bytes). e.g. 1k
// m/M for Megabytes (1 MB = 1000 KB). e.g. 1m
// g/G for Gigabytes (1 GB = 1000 MB). e.g. 1g
// t/T for Terrabytes (1 TB = 1000 GB). e.g. 1t
// p/P for Petabytes (1 PB = 1000 TB). e.g. 1p
// Valid prefixes are:
// + means Greater than. e.g. +1k
// - means Less than. e.g. -1k
func NewSize(expression string) (*searchSize, error) {
	s := &searchSize{}
	err := s.parse(expression)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *searchSize) parse(expression string) error {
	lenExp := len(expression)
	if lenExp == 0 {
		s.size = 0
		s.op = searchSizeOpEqual
		return nil
	}

	// Default is exact size in bytes
	blocksize := uint64(1)

	// Parse scale suffix
	suffix := expression[lenExp-1:]
	switch strings.ToLower(suffix) {
	case "k":
		// Kilobytes
		blocksize = 1000
		expression = expression[:lenExp-1]
	case "m":
		// Megabytes
		blocksize = 1000 * 1000
		expression = expression[:lenExp-1]
	case "g":
		// Gigabytes
		blocksize = 1000 * 1000 * 1000
		expression = expression[:lenExp-1]
	case "t":
		// Terrabytes
		blocksize = 1000 * 1000 * 1000 * 1000
		expression = expression[:lenExp-1]
	case "p":
		// Petabytes
		blocksize = 1000 * 1000 * 1000 * 1000 * 1000
		expression = expression[:lenExp-1]
	}

	lenExp = len(expression)
	if lenExp == 0 {
		return fmt.Errorf("failed to parse the size expression %q after removing scale suffix", expression)
	}

	prefix := expression[:1]
	switch prefix {
	case "+":
		s.op = searchSizeOpGreater
		expression = expression[1:lenExp]
	case "-":
		s.op = searchSizeOpLess
		expression = expression[1:lenExp]
	default:
		s.op = searchSizeOpEqual
	}

	value, err := strconv.Atoi(expression)
	if err != nil {
		return fmt.Errorf("failed to parse the size expression %q. %v", expression, err)
	}

	s.size = uint64(value) * blocksize

	return nil
}

func (s *searchSize) Match(pi path.Info, hash []byte) (bool, error) {
	matched := false
	switch s.op {
	case searchSizeOpEqual:
		matched = (pi.Size == s.size)
	case searchSizeOpGreater:
		matched = (pi.Size > s.size)
	case searchSizeOpLess:
		matched = (pi.Size < s.size)
	}

	return matched, nil
}

//-----------------------------------------------------------------------------
// Hash

type Hash struct {
	Prefix string
}

// Match if the entry's hash starts with the specified hash prefix. case insensitive.
func (s *Hash) Match(pi path.Info, hash []byte) (bool, error) {
	hashStr := hex.EncodeToString(hash)
	matched := strings.HasPrefix(strings.ToLower(hashStr), strings.ToLower(s.Prefix))
	return matched, nil
}

//-----------------------------------------------------------------------------
// Last modification time

type searchModTime struct {
	reference time.Time
	after     bool
}

// Match if the entry's last modification time is before the specified date.
// The following formats are allowed:
// YYYY-MM-DD
// YYYY-MM-DD HH:mm:ss
// YYYY-MM-DDTHH:mm:ss
// <n>D n Days before now. e.g. 10D
// <n>M n Months before now. .e.g. 2M
// <n>Y n Years before now. e.g. 5Y
func NewModTimeBefore(expression string) (*searchModTime, error) {
	s := &searchModTime{}
	err := s.parse(expression, false)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Match if the entry's last modification time is after the specified date.
// The following formats are allowed:
// YYYY-MM-DD
// YYYY-MM-DD HH:mm:ss
// YYYY-MM-DDTHH:mm:ss
func NewModTimeAfter(expression string) (*searchModTime, error) {
	s := &searchModTime{}
	err := s.parse(expression, true)
	if err != nil {
		return nil, err
	}
	return s, nil
}

const (
	searchModTimeSuffixNone = iota
	searchModTimeSuffixSeconds
	searchModTimeSuffixMinutes
	searchModTimeSuffixHours
	searchModTimeSuffixDays
	searchModTimeSuffixMonths
	searchModTimeSuffixYears
)

func (s *searchModTime) parse(expression string, after bool) error {
	from := time.Now()

	lenExp := len(expression)
	if lenExp < 2 {
		return fmt.Errorf("failed to parse the date/time expression %q", expression)
	}

	suffixOp := searchModTimeSuffixNone
	suffix := expression[lenExp-1:]
	switch suffix {
	case "s":
		// Seconds ago
		suffixOp = searchModTimeSuffixSeconds
		expression = expression[:lenExp-1]
	case "m":
		// Minutes ago
		suffixOp = searchModTimeSuffixMinutes
		expression = expression[:lenExp-1]
	case "h":
		// Hours ago
		suffixOp = searchModTimeSuffixHours
		expression = expression[:lenExp-1]
	case "D":
		// Days ago
		suffixOp = searchModTimeSuffixDays
		expression = expression[:lenExp-1]
	case "M":
		// Months ago
		suffixOp = searchModTimeSuffixMonths
		expression = expression[:lenExp-1]
	case "Y":
		// Years ago
		suffixOp = searchModTimeSuffixYears
		expression = expression[:lenExp-1]
	}

	lenExp = len(expression)
	if lenExp == 0 {
		return fmt.Errorf("failed to parse the date/time expression %q after removing suffix", expression)
	}

	if suffixOp == searchModTimeSuffixNone {
		parseDate := strings.Contains(expression, "-")
		parseTime := strings.Contains(expression, ":")

		format := ""
		if parseDate && parseTime {
			if strings.Contains(expression, "T") {
				format = "2006-01-02T15:04:05"
			} else {
				format = "2006-01-02 15:04:05"
			}
		} else if parseDate {
			format = "2006-01-02"
		} else if parseTime {
			format = "15:04:05"
		} else {
			return fmt.Errorf("failed to parse the date/time expression %q. unknown format", expression)
		}

		parsedDateTime, err := time.Parse(format, expression)
		if err != nil {
			return fmt.Errorf("failed to parse the date/time expression %q. %v", expression, err)
		}

		if parseTime && !parseDate {
			s.reference = time.Date(from.Year(),
				from.Month(),
				from.Day(),
				parsedDateTime.Hour(),
				parsedDateTime.Minute(),
				parsedDateTime.Second(),
				0,
				time.UTC)
		} else {
			s.reference = parsedDateTime
		}

	} else {
		// Suffix not allowed when using "after date" option
		if after {
			return fmt.Errorf("date/time search does not allow shorthand suffixes when using 'after' option. %q", expression)
		}

		value, err := strconv.Atoi(expression)
		if err != nil {
			return fmt.Errorf("failed to parse the date/time expression %q. %v", expression, err)
		}

		switch suffixOp {
		case searchModTimeSuffixSeconds:
			s.reference = from.Add(time.Second * -time.Duration(value))
		case searchModTimeSuffixMinutes:
			s.reference = from.Add(time.Minute * -time.Duration(value))
		case searchModTimeSuffixHours:
			s.reference = from.Add(time.Hour * -time.Duration(value))
		case searchModTimeSuffixDays:
			s.reference = from.AddDate(0, 0, -value)
		case searchModTimeSuffixMonths:
			s.reference = from.AddDate(0, -value, 0)
		case searchModTimeSuffixYears:
			s.reference = from.AddDate(-value, 0, 0)
		default:
			return fmt.Errorf("failed to parse the date/time expression %q. unknown suffix type", expression)
		}
	}

	s.reference = s.reference.Round(time.Second)
	s.after = after
	return nil
}

func (s *searchModTime) Match(pi path.Info, hash []byte) (bool, error) {
	compare := pi.ModTime.Compare(s.reference)
	if s.after {
		return compare == 1, nil
	}
	return compare == -1, nil
}

//-----------------------------------------------------------------------------
// Id

type Id struct {
	Prefix string
}

// Match if the entry's identifier starts with the specified prefix. case insensitive.
func (s *Id) Match(pi path.Info, hash []byte) (bool, error) {
	str := hex.EncodeToString(pi.Id[:])
	matched := strings.HasPrefix(strings.ToLower(str), strings.ToLower(s.Prefix))
	return matched, nil
}
