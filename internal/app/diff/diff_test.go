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

	err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedNothing, diff.ChangedNothing, func(d diff.Diff) error {
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

	err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedNothing, diff.ChangedNothing, func(d diff.Diff) error {
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

	err := diff.Compare(lhsPath, lhsPath, false, diff.ChangedNothing, diff.ChangedNothing, func(d diff.Diff) error {
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

	err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedNothing, diff.ChangedNothing, func(d diff.Diff) error {
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

func TestDiffCompareIncludeAndExcludeIsMutuallyExlusive(t *testing.T) {
	assert.ErrorContains(t, diff.Compare("./", "./", false, diff.ChangedMode, diff.ChangedSize, nil), "the include and exclude filters are mutually exclusive")
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
		desc   string
		filter diff.ChangedFlags
		exp    []string
	}{
		{
			desc:   "mode",
			filter: diff.ChangedMode,
			exp: []string{
				"fm~~~ both/7.txt",
			},
		},
		{
			desc:   "size",
			filter: diff.ChangedSize,
			exp: []string{
				"f~s~~ both/6.txt",
			},
		},
		{
			desc:   "last mod",
			filter: diff.ChangedModTime,
			exp: []string{
				"f~~l~ both/8.txt",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			changed := make([]string, 0, 10)

			err := diff.Compare(lhsPath, rhsPath, false, tC.filter, diff.ChangedNothing, func(d diff.Diff) error {
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

			slices.Sort(tC.exp)
			slices.Sort(changed)

			assert.Equal(t, tC.exp, changed)
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
		desc   string
		filter diff.ChangedFlags
		exp    []string
	}{
		{
			desc:   "mode",
			filter: diff.ChangedMode,
			exp: []string{
				"f~s~~ both/6.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:   "size",
			filter: diff.ChangedSize,
			exp: []string{
				"fm~~~ both/7.txt",
				"f~~l~ both/8.txt",
			},
		},
		{
			desc:   "last mod",
			filter: diff.ChangedModTime,
			exp: []string{
				"f~s~~ both/6.txt",
				"fm~~~ both/7.txt",
			},
		},
		{
			desc:   "mode & size",
			filter: diff.ChangedMode | diff.ChangedSize,
			exp: []string{
				"f~~l~ both/8.txt",
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			changed := make([]string, 0, 10)

			err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedNothing, tC.filter, func(d diff.Diff) error {
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

			slices.Sort(tC.exp)
			slices.Sort(changed)

			assert.Equal(t, tC.exp, changed)
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

	err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedHash, diff.ChangedNothing, func(d diff.Diff) error {
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

	err := diff.Compare(lhsPath, rhsPath, false, diff.ChangedNothing, diff.ChangedHash, func(d diff.Diff) error {
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
