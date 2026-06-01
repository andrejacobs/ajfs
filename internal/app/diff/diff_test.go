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

package diff_test

import (
	"io"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/diff"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiffString(t *testing.T) {
	testCases := []struct {
		typ   diff.Type
		path  string
		isDir bool
		flags diff.ChangedFlags
		exp   string
	}{
		{
			typ:   diff.TypeLeftOnly,
			path:  "a.txt",
			isDir: false,
			flags: 0,
			exp:   "f---- a.txt",
		},
		{
			typ:   diff.TypeRightOnly,
			path:  "a.txt",
			isDir: false,
			flags: 0,
			exp:   "f++++ a.txt",
		},
		{
			typ:   diff.TypeLeftOnly,
			path:  "dirA",
			isDir: true,
			flags: 0,
			exp:   "d---- dirA",
		},
		{
			typ:   diff.TypeRightOnly,
			path:  "dirA",
			isDir: true,
			flags: 0,
			exp:   "d++++ dirA",
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedSize,
			exp:   "f~s~~ a.txt",
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedMode,
			exp:   "fm~~~ a.txt",
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedModTime,
			exp:   "f~~l~ a.txt",
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedHash,
			exp:   "f~~~x a.txt",
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedSize | diff.ChangedMode | diff.ChangedModTime | diff.ChangedHash,
			exp:   "fmslx a.txt",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.exp, func(t *testing.T) {
			d := diff.Diff{
				Type:    tC.typ,
				Id:      path.IdFromPath(tC.path),
				Path:    tC.path,
				IsDir:   tC.isDir,
				Changed: tC.flags,
			}
			assert.Equal(t, tC.exp, d.String())
		})
	}
}

func TestDiffCompare(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompare test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	lhs := make([]string, 0, 10)
	rhs := make([]string, 0, 10)
	changed := make([]string, 0, 10)

	err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{}, []diff.FilterFlags{}, func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeLeftOnly:
			lhs = append(lhs, d.String())
		case diff.TypeRightOnly:
			rhs = append(rhs, d.String())
		case diff.TypeChanged:
			changed = append(changed, d.String())
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "invalid type")
		}

		return nil
	})
	require.NoError(t, err)

	expectedLHSOnly := []string{
		"d---- quick",
		"f---- quick/1.txt",
		"f---- quick/2.txt",
		"d---- dir1",
		"f---- dir1/lhs-only",
	}

	expectedRHSOnly := []string{
		"d++++ fox",
		"f++++ fox/3.txt",
		"d++++ hole",
		"f++++ hole/4.txt",
		"d++++ dir2",
		"f++++ dir2/rhs-only",
	}
	expectedChanged := []string{
		"f~s~~ both/6.txt",
		"fm~~~ both/7.txt",
		"f~~l~ both/8.txt",
	}

	slices.Sort(expectedLHSOnly)
	slices.Sort(expectedRHSOnly)
	slices.Sort(expectedChanged)
	slices.Sort(lhs)
	slices.Sort(rhs)
	slices.Sort(changed)

	assert.Equal(t, expectedLHSOnly, lhs)
	assert.Equal(t, expectedRHSOnly, rhs)
	assert.Equal(t, expectedChanged, changed)
}

func TestDiffCompareWithHashes(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root:            "../../testdata/diff/c",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/d"
	require.NoError(t, scan.Run(cfg))

	changed := make([]string, 0, 10)

	err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{}, []diff.FilterFlags{}, func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeChanged:
			changed = append(changed, d.String())
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "invalid type")
		}

		return nil
	})
	require.NoError(t, err)

	expectedChanged := []string{
		"f~~~x changed.txt",
	}
	slices.Sort(expectedChanged)
	slices.Sort(changed)

	assert.Equal(t, expectedChanged, changed)
}

func TestDiffCompareSame(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	err := diff.Compare(lhsPath, lhsPath, []diff.FilterFlags{}, []diff.FilterFlags{}, func(d diff.Diff) error {
		switch d.Type {
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "there should have been no differences")
		}

		return nil
	})
	require.NoError(t, err)
}

func TestDiffCompareOrder(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping CompareDiffOrder test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	// We are testing that the order of diffs are always, LHS only, followed by RHS only, lastly followed by Changed.
	// 0 = LHS, 1 = RHS, 2 == Changed
	state := 0

	err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{}, []diff.FilterFlags{}, func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeLeftOnly:
			require.LessOrEqual(t, state, 0)
			state = 0
		case diff.TypeRightOnly:
			require.LessOrEqual(t, state, 1)
			state = 1
		case diff.TypeChanged:
			require.LessOrEqual(t, state, 2)
			state = 2
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "invalid type")
		}

		return nil
	})
	require.NoError(t, err)
}

//-----------------------------------------------------------------------------
// Filtering

func TestFilterFlagsValidate(t *testing.T) {
	assert.ErrorContains(t, diff.FilterFlags(diff.FilterTypeLeft|diff.FilterTypeRight).Validate(), "filtering on left hand side only or right hand side only is mutually exclusive")
	assert.ErrorContains(t, diff.FilterFlags(diff.FilterTypeLeft|diff.FilterChangedSize).Validate(), "can't filter on left hand side only and changes")
	assert.ErrorContains(t, diff.FilterFlags(diff.FilterTypeRight|diff.FilterChangedSize).Validate(), "can't filter on right hand side only and changes")
}

func TestFilterFlagsString(t *testing.T) {
	testCases := []struct {
		exp   string
		flags diff.FilterFlags
	}{
		{exp: "", flags: diff.FilterNoOp},
		{exp: "-", flags: diff.FilterTypeLeft},
		{exp: "+", flags: diff.FilterTypeRight},
		{exp: "~", flags: diff.FilterTypeChanged},
		{exp: "d", flags: diff.FilterDirs},
		{exp: "f", flags: diff.FilterFiles},
		{exp: "m", flags: diff.FilterChangedMode},
		{exp: "s", flags: diff.FilterChangedSize},
		{exp: "l", flags: diff.FilterChangedModTime},
		{exp: "x", flags: diff.FilterChangedHash},
		{exp: "fmslx", flags: diff.FilterFiles | diff.FilterChangedMode | diff.FilterChangedSize | diff.FilterChangedModTime | diff.FilterChangedHash},
		{exp: "~fmslx", flags: diff.FilterTypeChanged | diff.FilterFiles | diff.FilterChangedMode | diff.FilterChangedSize | diff.FilterChangedModTime | diff.FilterChangedHash},
	}
	for _, tC := range testCases {
		t.Run(tC.exp, func(t *testing.T) {
			assert.Equal(t, tC.exp, tC.flags.String())
		})
	}
}

func TestParseFilterFlags(t *testing.T) {
	testCases := []struct {
		exp   diff.FilterFlags
		input string
	}{
		{
			exp:   diff.FilterNoOp,
			input: "",
		},
		{
			exp:   diff.FilterTypeLeft,
			input: "-",
		},
		{
			exp:   diff.FilterTypeRight,
			input: "+",
		},
		{
			exp:   diff.FilterTypeChanged,
			input: "~",
		},
		{
			exp:   diff.FilterDirs,
			input: "d",
		},
		{
			exp:   diff.FilterFiles,
			input: "f",
		},
		{
			exp:   diff.FilterChangedMode,
			input: "m",
		},
		{
			exp:   diff.FilterChangedSize,
			input: "s",
		},
		{
			exp:   diff.FilterChangedModTime,
			input: "l",
		},
		{
			exp:   diff.FilterChangedHash,
			input: "x",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.exp.String(), func(t *testing.T) {
			result, err := diff.ParseFilterFlags(tC.input)
			require.NoError(t, err)
			assert.Equal(t, tC.exp, result)
		})
	}

	_, err := diff.ParseFilterFlags("abcd")
	require.ErrorContains(t, err, "invalid filter: abcd")
}

func TestParseFilterFlagsArray(t *testing.T) {
	testCases := []struct {
		desc  string
		exp   []diff.FilterFlags
		input []string
	}{
		{
			desc:  "nil",
			exp:   []diff.FilterFlags{},
			input: nil,
		},
		{
			desc:  "empty array",
			exp:   []diff.FilterFlags{},
			input: []string{""},
		},
		{
			desc:  "lhs or rhs",
			exp:   []diff.FilterFlags{diff.FilterTypeLeft, diff.FilterTypeRight},
			input: []string{"-", "+"},
		},
		{
			desc:  "fs or -fl",
			exp:   []diff.FilterFlags{diff.FilterFiles | diff.FilterChangedSize, diff.FilterFiles | diff.FilterTypeLeft | diff.FilterChangedModTime},
			input: []string{"fs", "f-l"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result, err := diff.ParseFilterFlagsArray(tC.input)
			require.NoError(t, err)
			assert.Equal(t, tC.exp, result)
		})
	}

	_, err := diff.ParseFilterFlags("abcd")
	require.ErrorContains(t, err, "invalid filter: abcd")
}

func TestFilterFlagsChangedFlagsMask(t *testing.T) {
	testCases := []struct {
		exp   diff.ChangedFlags
		flags diff.FilterFlags
	}{
		{exp: diff.ChangedNothing, flags: diff.FilterNoOp},
		{exp: diff.ChangedNothing, flags: diff.FilterTypeLeft | diff.FilterFiles},
		{exp: diff.ChangedMode, flags: diff.FilterFiles | diff.FilterChangedMode},
		{exp: diff.ChangedSize, flags: diff.FilterFiles | diff.FilterChangedSize},
		{exp: diff.ChangedModTime, flags: diff.FilterFiles | diff.FilterChangedModTime},
		{exp: diff.ChangedHash, flags: diff.FilterFiles | diff.FilterChangedHash},
	}
	for _, tC := range testCases {
		t.Run(tC.flags.String(), func(t *testing.T) {
			assert.Equal(t, tC.exp, tC.flags.ChangedFlagsMask())
		})
	}
}

func TestDiffFilterFlagsMask(t *testing.T) {
	testCases := []struct {
		typ   diff.Type
		path  string
		isDir bool
		flags diff.ChangedFlags
		exp   diff.FilterFlags
	}{
		{
			typ:   diff.TypeLeftOnly,
			path:  "a.txt",
			isDir: false,
			flags: 0,
			exp:   diff.FilterTypeLeft | diff.FilterFiles,
		},
		{
			typ:   diff.TypeRightOnly,
			path:  "a.txt",
			isDir: false,
			flags: 0,
			exp:   diff.FilterTypeRight | diff.FilterFiles,
		},
		{
			typ:   diff.TypeLeftOnly,
			path:  "dirA",
			isDir: true,
			flags: 0,
			exp:   diff.FilterTypeLeft | diff.FilterDirs,
		},
		{
			typ:   diff.TypeRightOnly,
			path:  "dirA",
			isDir: true,
			flags: 0,
			exp:   diff.FilterTypeRight | diff.FilterDirs,
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedSize,
			exp:   diff.FilterFiles | diff.FilterTypeChanged | diff.FilterChangedSize,
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedMode,
			exp:   diff.FilterFiles | diff.FilterTypeChanged | diff.FilterChangedMode,
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedModTime,
			exp:   diff.FilterFiles | diff.FilterTypeChanged | diff.FilterChangedModTime,
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedHash,
			exp:   diff.FilterFiles | diff.FilterTypeChanged | diff.FilterChangedHash,
		},
		{
			typ:   diff.TypeChanged,
			path:  "a.txt",
			isDir: false,
			flags: diff.ChangedSize | diff.ChangedMode | diff.ChangedModTime | diff.ChangedHash,
			exp:   diff.FilterFiles | diff.FilterTypeChanged | diff.FilterChangedSize | diff.FilterChangedMode | diff.FilterChangedModTime | diff.FilterChangedHash,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.exp.String(), func(t *testing.T) {
			d := diff.Diff{
				Type:    tC.typ,
				Id:      path.IdFromPath(tC.path),
				Path:    tC.path,
				IsDir:   tC.isDir,
				Changed: tC.flags,
			}
			assert.Equal(t, tC.exp, d.FilterFlagsMask())
		})
	}
}

func TestDiffCompareIncludeFilter(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompareIncludeFilter test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	testCases := []struct {
		desc    string
		filters []diff.FilterFlags
		exp     []string
	}{
		{
			desc:    "lhs",
			filters: []diff.FilterFlags{diff.FilterTypeLeft},
			exp: []string{
				"d---- dir1",
				"f---- dir1/lhs-only",
				"d---- quick",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
			},
		},
		{
			desc:    "rhs",
			filters: []diff.FilterFlags{diff.FilterTypeRight},
			exp: []string{
				"d++++ dir2",
				"f++++ dir2/rhs-only",
				"d++++ fox",
				"f++++ fox/3.txt",
				"d++++ hole",
				"f++++ hole/4.txt"},
		},
		{
			desc:    "changed",
			filters: []diff.FilterFlags{diff.FilterTypeChanged},
			exp: []string{
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "files",
			filters: []diff.FilterFlags{diff.FilterFiles},
			exp: []string{
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "files lhs",
			filters: []diff.FilterFlags{diff.FilterTypeLeft | diff.FilterFiles},
			exp: []string{
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
			},
		},
		{
			desc:    "mode",
			filters: []diff.FilterFlags{diff.FilterChangedMode},
			exp: []string{
				"fm~~~ both/7.txt",
			},
		},
		{
			desc:    "size",
			filters: []diff.FilterFlags{diff.FilterChangedSize},
			exp: []string{
				"f~s~~ both/6.txt",
			},
		},
		{
			desc:    "last mod",
			filters: []diff.FilterFlags{diff.FilterChangedModTime},
			exp: []string{
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "file && size or mode",
			filters: []diff.FilterFlags{diff.FilterFiles | diff.FilterChangedSize, diff.FilterFiles | diff.FilterChangedMode},
			exp: []string{
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := make([]string, 0, 10)

			err := diff.Compare(lhsPath, rhsPath, tC.filters, []diff.FilterFlags{}, func(d diff.Diff) error {
				if d.Path == "." {
					return nil
				}
				result = append(result, d.String())
				return nil
			})
			require.NoError(t, err)

			slices.Sort(tC.exp)
			slices.Sort(result)

			assert.Equal(t, tC.exp, result)
		})
	}
}

func TestDiffCompareExcludeFilter(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompareExcludeFilter test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	testCases := []struct {
		desc    string
		filters []diff.FilterFlags
		exp     []string
	}{
		{
			desc:    "lhs",
			filters: []diff.FilterFlags{diff.FilterTypeLeft},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "rhs",
			filters: []diff.FilterFlags{diff.FilterTypeRight},
			exp: []string{
				"d---- dir1",
				"d---- quick",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "changed",
			filters: []diff.FilterFlags{diff.FilterTypeChanged},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"d---- dir1",
				"d---- quick",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
			},
		},
		{
			desc:    "files",
			filters: []diff.FilterFlags{diff.FilterFiles},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"d---- dir1",
				"d---- quick",
			},
		},
		{
			desc:    "dir",
			filters: []diff.FilterFlags{diff.FilterDirs},
			exp: []string{
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "mode",
			filters: []diff.FilterFlags{diff.FilterChangedMode},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"d---- dir1",
				"d---- quick",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "size",
			filters: []diff.FilterFlags{diff.FilterChangedSize},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"d---- dir1",
				"d---- quick",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"fm~~~ both/7.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:    "last mod",
			filters: []diff.FilterFlags{diff.FilterChangedModTime},
			exp: []string{
				"d++++ dir2",
				"d++++ fox",
				"d++++ hole",
				"d---- dir1",
				"d---- quick",
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"fm~~~ both/7.txt",
				"f~s~~ both/6.txt",
			},
		},
		{
			desc:    "exclude a lot",
			filters: []diff.FilterFlags{diff.FilterTypeLeft, diff.FilterTypeRight, diff.FilterDirs, diff.FilterChangedModTime},
			exp: []string{
				"f~s~~ both/6.txt",
				"fm~~~ both/7.txt",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := make([]string, 0, 10)

			err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{}, tC.filters, func(d diff.Diff) error {
				if d.Path == "." {
					return nil
				}
				result = append(result, d.String())
				return nil
			})
			require.NoError(t, err)

			slices.Sort(tC.exp)
			slices.Sort(result)

			assert.Equal(t, tC.exp, result)
		})
	}
}

func TestDiffCompareIncludeFilterWithHashes(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root:            "../../testdata/diff/c",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/d"
	require.NoError(t, scan.Run(cfg))

	changed := make([]string, 0, 10)

	err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{diff.FilterChangedHash}, []diff.FilterFlags{}, func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeChanged:
			changed = append(changed, d.String())
		}

		return nil
	})
	require.NoError(t, err)

	expectedChanged := []string{
		"f~~~x changed.txt",
	}
	slices.Sort(expectedChanged)
	slices.Sort(changed)

	assert.Equal(t, expectedChanged, changed)
}

func TestDiffCompareExcludeFilterWithHashes(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root:            "../../testdata/diff/c",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/d"
	require.NoError(t, scan.Run(cfg))

	changed := make([]string, 0, 10)

	err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{diff.FilterNoOp}, []diff.FilterFlags{diff.FilterChangedHash}, func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeChanged:
			changed = append(changed, d.String())
		}

		return nil
	})
	require.NoError(t, err)

	assert.Len(t, changed, 0)
}

func TestDiffCompareIncludeAndExcludeFilter(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompareIncludeAndExcludeFilter test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	testCases := []struct {
		desc    string
		include diff.FilterFlags
		exclude diff.FilterFlags
		exp     []string
	}{
		{
			desc:    "lhs",
			include: diff.FilterFiles,
			exclude: diff.FilterChangedSize,
			exp: []string{
				"f++++ dir2/rhs-only",
				"f++++ fox/3.txt",
				"f++++ hole/4.txt",
				"f---- dir1/lhs-only",
				"f---- quick/1.txt",
				"f---- quick/2.txt",
				"fm~~~ both/7.txt",
				"f~~l~ both/8.txt"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := make([]string, 0, 10)

			err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{tC.include}, []diff.FilterFlags{tC.exclude}, func(d diff.Diff) error {
				if d.Path == "." {
					return nil
				}
				result = append(result, d.String())
				return nil
			})
			require.NoError(t, err)

			slices.Sort(tC.exp)
			slices.Sort(result)

			assert.Equal(t, tC.exp, result)
		})
	}
}

//-----------------------------------------------------------------------------
// Stats tests

func TestDiffStats(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompareIncludeFilter test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	cfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(cfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	cfg.DbPath = rhsPath
	cfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(cfg))

	testCases := []struct {
		desc    string
		include diff.FilterFlags
		exclude diff.FilterFlags
		exp     diff.DiffStats
	}{
		{
			desc:    "no filter",
			include: diff.FilterNoOp,
			exclude: diff.FilterNoOp,
			exp: diff.DiffStats{
				LeftOnly:       5,
				RightOnly:      6,
				Changed:        4,
				NotChanged:     2,
				Files:          9,
				Dirs:           6,
				ModeChanged:    1,
				SizeChanged:    2,
				ModTimeChanged: 2,
			},
		},
		{
			desc:    "lhs",
			include: diff.FilterTypeLeft,
			exclude: diff.FilterNoOp,
			exp: diff.DiffStats{
				LeftOnly: 5,
				Files:    3,
				Dirs:     2,
			},
		},
		{
			desc:    "rhs",
			include: diff.FilterTypeRight,
			exclude: diff.FilterNoOp,
			exp: diff.DiffStats{
				RightOnly: 6,
				Files:     3,
				Dirs:      3,
			},
		},
		{
			desc:    "changed",
			include: diff.FilterTypeChanged,
			exclude: diff.FilterNoOp,
			exp: diff.DiffStats{
				Changed:        4,
				Files:          3,
				Dirs:           1,
				ModeChanged:    1,
				SizeChanged:    2,
				ModTimeChanged: 2,
			},
		},
		{
			desc:    "lhs files",
			include: diff.FilterTypeLeft | diff.FilterFiles,
			exclude: diff.FilterNoOp,
			exp: diff.DiffStats{
				LeftOnly: 3,
				Files:    3,
			},
		},
		{
			desc:    "changed files excl size",
			include: diff.FilterFiles | diff.FilterTypeChanged,
			exclude: diff.FilterChangedSize,
			exp: diff.DiffStats{
				Changed:        2,
				Files:          2,
				ModeChanged:    1,
				SizeChanged:    0,
				ModTimeChanged: 1,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			result := diff.DiffStats{
				Fn: func(d diff.Diff) error {
					if d.Path == "." {
						return nil
					}
					return nil
				},
			}

			err := diff.Compare(lhsPath, rhsPath, []diff.FilterFlags{tC.include}, []diff.FilterFlags{tC.exclude}, result.Compare)
			require.NoError(t, err)

			result.Fn = nil
			assert.Equal(t, tC.exp, result)
		})
	}
}

//-----------------------------------------------------------------------------
// Run tests

func TestRunTwoDirs(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompare test")
		return
	}

	lhs := make([]string, 0, 10)
	rhs := make([]string, 0, 10)
	changed := make([]string, 0, 10)

	fn := func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeLeftOnly:
			lhs = append(lhs, d.String())
		case diff.TypeRightOnly:
			rhs = append(rhs, d.String())
		case diff.TypeChanged:
			changed = append(changed, d.String())
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "invalid type")
		}

		return nil
	}

	cfg := diff.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: "../../testdata/diff/a",
		RhsPath: "../../testdata/diff/b",
		Fn:      fn,
	}

	err := diff.Run(cfg)
	require.NoError(t, err)

	expectedLHSOnly := []string{
		"d---- quick",
		"f---- quick/1.txt",
		"f---- quick/2.txt",
		"d---- dir1",
		"f---- dir1/lhs-only",
	}

	expectedRHSOnly := []string{
		"d++++ fox",
		"f++++ fox/3.txt",
		"d++++ hole",
		"f++++ hole/4.txt",
		"d++++ dir2",
		"f++++ dir2/rhs-only",
	}
	expectedChanged := []string{
		"f~s~~ both/6.txt",
		"fm~~~ both/7.txt",
		"f~~l~ both/8.txt",
	}

	slices.Sort(expectedLHSOnly)
	slices.Sort(expectedRHSOnly)
	slices.Sort(expectedChanged)
	slices.Sort(lhs)
	slices.Sort(rhs)
	slices.Sort(changed)

	assert.Equal(t, expectedLHSOnly, lhs)
	assert.Equal(t, expectedRHSOnly, rhs)
	assert.Equal(t, expectedChanged, changed)
}

func TestRunTwoDatabases(t *testing.T) {
	if os.Getenv("SKIP_TEST") == "1" {
		t.Skip("Skipping DiffCompare test")
		return
	}

	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(scanCfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	scanCfg.DbPath = rhsPath
	scanCfg.Root = "../../testdata/diff/b"
	require.NoError(t, scan.Run(scanCfg))

	lhs := make([]string, 0, 10)
	rhs := make([]string, 0, 10)
	changed := make([]string, 0, 10)

	fn := func(d diff.Diff) error {
		if d.Path == "." {
			return nil
		}
		switch d.Type {
		case diff.TypeLeftOnly:
			lhs = append(lhs, d.String())
		case diff.TypeRightOnly:
			rhs = append(rhs, d.String())
		case diff.TypeChanged:
			changed = append(changed, d.String())
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "invalid type")
		}

		return nil
	}

	cfg := diff.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		RhsPath: rhsPath,
		Fn:      fn,
	}

	err := diff.Run(cfg)
	require.NoError(t, err)

	expectedLHSOnly := []string{
		"d---- quick",
		"f---- quick/1.txt",
		"f---- quick/2.txt",
		"d---- dir1",
		"f---- dir1/lhs-only",
	}

	expectedRHSOnly := []string{
		"d++++ fox",
		"f++++ fox/3.txt",
		"d++++ hole",
		"f++++ hole/4.txt",
		"d++++ dir2",
		"f++++ dir2/rhs-only",
	}
	expectedChanged := []string{
		"f~s~~ both/6.txt",
		"fm~~~ both/7.txt",
		"f~~l~ both/8.txt",
	}

	slices.Sort(expectedLHSOnly)
	slices.Sort(expectedRHSOnly)
	slices.Sort(expectedChanged)
	slices.Sort(lhs)
	slices.Sort(rhs)
	slices.Sort(changed)

	assert.Equal(t, expectedLHSOnly, lhs)
	assert.Equal(t, expectedRHSOnly, rhs)
	assert.Equal(t, expectedChanged, changed)
}

func TestRunSingleDatabases(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(scanCfg))

	fn := func(d diff.Diff) error {
		switch d.Type {
		case diff.TypeNothing:
			// nothing changed
		default:
			require.Fail(t, "there should have been no differences")
		}

		return nil
	}

	cfg := diff.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		Fn:      fn,
	}

	err := diff.Run(cfg)
	require.NoError(t, err)
}

func TestSkipAll(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root: "../../testdata/diff/a",
	}
	require.NoError(t, scan.Run(scanCfg))

	count := 0
	fn := func(d diff.Diff) error {
		count++
		return diff.SkipAll
	}

	cfg := diff.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		Fn:      fn,
	}

	err := diff.Run(cfg)
	require.NoError(t, err)

	assert.Equal(t, 1, count)
}

func TestRunTwoDatabasesWithDifferentHashAlgos(t *testing.T) {
	lhsPath := filepath.Join(t.TempDir(), "unit-testing-lhs")
	_ = os.Remove(lhsPath)
	defer os.Remove(lhsPath)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: lhsPath,
		},
		Root:            "../../testdata/diff/a",
		CalculateHashes: true,
		Algo:            ajhash.AlgoSHA1,
	}
	require.NoError(t, scan.Run(scanCfg))

	rhsPath := filepath.Join(t.TempDir(), "unit-testing-rhs")
	_ = os.Remove(rhsPath)
	defer os.Remove(rhsPath)

	scanCfg.DbPath = rhsPath
	scanCfg.Root = "../../testdata/diff/b"
	scanCfg.Algo = ajhash.AlgoSHA256
	require.NoError(t, scan.Run(scanCfg))

	fn := func(d diff.Diff) error {
		require.False(t, d.Changed.HashChanged())
		return nil
	}

	cfg := diff.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		LhsPath: lhsPath,
		RhsPath: rhsPath,
		Fn:      fn,
	}

	err := diff.Run(cfg)
	require.NoError(t, err)
}
