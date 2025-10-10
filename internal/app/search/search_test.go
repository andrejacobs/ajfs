package search_test

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/scan"
	"github.com/andrejacobs/ajfs/internal/app/search"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFunc(t *testing.T) {
	wasCalled := false
	var withPi path.Info
	var withHash []byte

	s := search.NewFunc(func(pi path.Info, hash []byte) (bool, error) {
		wasCalled = true
		withPi = pi
		withHash = hash
		return true, nil
	})

	p1 := path.Info{
		Id:      path.IdFromPath("a.txt"),
		Path:    "a.txt",
		Size:    42,
		Mode:    0,
		ModTime: time.Now(),
	}

	m, err := s.Match(p1, nil)
	require.NoError(t, err)
	assert.True(t, m)

	assert.True(t, wasCalled)
	assert.True(t, p1.Equals(&withPi))
	assert.Nil(t, withHash)

	wasCalled = false
	withPi = path.Info{}

	m, err = s.Match(p1, []byte("The quick brown fox"))
	require.NoError(t, err)
	assert.True(t, m)

	assert.True(t, wasCalled)
	assert.True(t, p1.Equals(&withPi))
	assert.Equal(t, []byte("The quick brown fox"), withHash)
}

func TestAlways(t *testing.T) {
	s := search.Always{}
	m, err := s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.True(t, m)
}

func TestNever(t *testing.T) {
	s := search.Never{}
	m, err := s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.False(t, m)
}

func TestAnd(t *testing.T) {
	f1Called := false
	f2Called := false

	f1 := func(pi path.Info, hash []byte) (bool, error) {
		f1Called = true
		return true, nil
	}

	f2 := func(pi path.Info, hash []byte) (bool, error) {
		f2Called = true
		return true, nil
	}

	s := search.NewAnd(search.NewFunc(f1), search.NewFunc(f2))
	m, err := s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.True(t, m)
	assert.True(t, f1Called)
	assert.True(t, f2Called)
}

func TestOr(t *testing.T) {
	f1Called := false
	f2Called := false

	f1 := func(pi path.Info, hash []byte) (bool, error) {
		f1Called = true
		return true, nil
	}

	f2 := func(pi path.Info, hash []byte) (bool, error) {
		f2Called = true
		return true, nil
	}

	s := search.NewOr(search.NewFunc(f1), search.NewFunc(f2))
	m, err := s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.True(t, m)
	assert.True(t, f1Called)
	assert.False(t, f2Called)

	f2Called = false
	s = search.NewOr(&search.Never{}, search.NewFunc(f2))
	m, err = s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.True(t, m)
	assert.True(t, f2Called)
}

func TestNot(t *testing.T) {
	s := search.NewNot(&search.Always{})
	m, err := s.Match(path.Info{}, nil)
	require.NoError(t, err)
	assert.False(t, m)
}

func TestRegex(t *testing.T) {
	s, err := search.NewRegex("qu")
	require.NoError(t, err)

	m, err := s.Match(path.Info{Path: "/the/quick/brown"}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	m, err = s.Match(path.Info{Path: "/the/slow/brown"}, nil)
	require.NoError(t, err)
	assert.False(t, m)

	// Case insensitive (just for documentation)
	s, err = search.NewRegex("(?i)QU")
	require.NoError(t, err)

	m, err = s.Match(path.Info{Path: "/the/queen/bee"}, nil)
	require.NoError(t, err)
	assert.True(t, m)
}

func TestShellPattern(t *testing.T) {
	testCases := []struct {
		desc        string
		pattern     string
		baseOnly    bool
		insensitive bool
		input       string
		expected    bool
	}{
		{pattern: "*.txt", baseOnly: true, input: "/etc/test.txt", expected: true, desc: "*.txt"},
		{pattern: "*.txt", baseOnly: true, input: "/etc/test.zip", expected: false, desc: "*.txt - false"},
		{pattern: "*.txt", baseOnly: true, input: "/etc/test.TxT", expected: false, desc: "*.txt - insensitive - false"},
		{pattern: "*.txt", baseOnly: true, insensitive: true, input: "/etc/test.TxT", expected: true, desc: "*.txt"},
		{pattern: "*.a?", baseOnly: true, input: "/etc/test.a2", expected: true, desc: "*.a?"},

		{pattern: "/etc/*.txt", input: "/etc/test.txt", expected: true, desc: "/etc/*.txt"},
		{pattern: "/etc/*.txt", input: "/etcd/test.txt", expected: false, desc: "/etc/*.txt - false"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, err := search.NewShellPattern(tC.pattern, tC.baseOnly, tC.insensitive)
			require.NoError(t, err)

			m, err := s.Match(path.Info{Path: tC.input}, nil)
			require.NoError(t, err)
			assert.Equal(t, tC.expected, m)
		})
	}
}

func TestType(t *testing.T) {
	s, err := search.NewType("d")
	require.NoError(t, err)

	m, err := s.Match(path.Info{Mode: fs.ModeDir}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	m, err = s.Match(path.Info{Mode: fs.ModeSymlink}, nil)
	require.NoError(t, err)
	assert.False(t, m)

	s, err = search.NewType("f")
	require.NoError(t, err)

	m, err = s.Match(path.Info{Mode: 0}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	m, err = s.Match(path.Info{Mode: fs.ModeDir}, nil)
	require.NoError(t, err)
	assert.False(t, m)

	s, err = search.NewType("l")
	require.NoError(t, err)

	m, err = s.Match(path.Info{Mode: fs.ModeSymlink}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	s, err = search.NewType("p")
	require.NoError(t, err)

	m, err = s.Match(path.Info{Mode: fs.ModeNamedPipe}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	s, err = search.NewType("s")
	require.NoError(t, err)

	m, err = s.Match(path.Info{Mode: fs.ModeSocket}, nil)
	require.NoError(t, err)
	assert.True(t, m)

	_, err = search.NewType("x")
	require.Error(t, err)
}

func TestSize(t *testing.T) {
	testCases := []struct {
		desc          string
		exp           string
		size          uint64
		expected      bool
		expectedError string
	}{
		{desc: "Not a valid expression", exp: "100z", expectedError: "failed to parse the size expression"},
		{desc: "Not a number", exp: "zebra", expectedError: "failed to parse the size expression"},
		{desc: "Not valid after suffix cut", exp: "k", expectedError: "after removing scale suffix"},
		{desc: "Empty expression", exp: "", size: 0, expected: true},
		{desc: "Exact size", exp: "100", size: 100, expected: true},
		{desc: "Kilobytes", exp: "1k", size: 1000, expected: true},
		{desc: "Megabytes", exp: "100M", size: 100 * 1000 * 1000, expected: true},
		{desc: "Gigabytes", exp: "100G", size: 100 * 1000 * 1000 * 1000, expected: true},
		{desc: "Terrabytes", exp: "100T", size: 100 * 1000 * 1000 * 1000 * 1000, expected: true},
		{desc: "Petabytes", exp: "100p", size: 100 * 1000 * 1000 * 1000 * 1000 * 1000, expected: true},
		{desc: "Greater than - exact size - pass", exp: "+400", size: 432, expected: true},
		{desc: "Greater than - exact size - false", exp: "+400", size: 399, expected: false},
		{desc: "Greater than - pass", exp: "+1k", size: 1100, expected: true},
		{desc: "Greater than - fail", exp: "+1k", size: 1 * 1000, expected: false},
		{desc: "Less than - exact size - pass", exp: "-400", size: 399, expected: true},
		{desc: "Less than - exact size - false", exp: "-400", size: 400, expected: false},
		{desc: "Less than - pass", exp: "-2k", size: 1890, expected: true},
		{desc: "Less than - fail", exp: "-2k", size: 2 * 1000, expected: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			s, err := search.NewSize(tC.exp)
			if tC.expectedError != "" {
				assert.ErrorContains(t, err, tC.expectedError)
				return
			} else {
				assert.NoError(t, err)
			}

			m, err := s.Match(path.Info{Size: tC.size}, nil)
			require.NoError(t, err)
			assert.Equal(t, tC.expected, m)
		})
	}
}

func TestHash(t *testing.T) {
	e := search.Hash{Prefix: "abc"}

	hs, _ := hex.DecodeString("abcdef")
	m, err := e.Match(path.Info{}, hs)
	require.NoError(t, err)
	assert.True(t, m)

	hs, _ = hex.DecodeString("AbCdef")
	m, err = e.Match(path.Info{}, hs) // case insensitive
	require.NoError(t, err)
	assert.True(t, m)

	hs, _ = hex.DecodeString("abxyz")
	m, err = e.Match(path.Info{}, hs)
	require.NoError(t, err)
	assert.False(t, m)
}

func TestModTimeExpression(t *testing.T) {
	now := time.Now().Round(time.Second)

	testCases := []struct {
		desc     string
		exp      string
		after    bool
		modTime  time.Time
		expected bool
	}{
		{exp: "1999-08-27", modTime: time.Date(1985, 1, 20, 0, 0, 0, 0, time.UTC), expected: true, desc: "before YYYY-MM-DD"},
		{exp: "1999-08-27", modTime: time.Date(1999, 9, 11, 0, 0, 0, 0, time.UTC), expected: false, desc: "before YYYY-MM-DD - false"},
		{exp: "1999-08-27 14:32:42", modTime: time.Date(1999, 8, 27, 13, 0, 0, 0, time.UTC), expected: true, desc: "before YYYY-MM-DD HH:mm:ss"},
		{exp: "1999-08-27 14:32:42", modTime: time.Date(1999, 8, 27, 15, 0, 0, 0, time.UTC), expected: false, desc: "before YYYY-MM-DD HH:mm:ss - false"},
		{exp: "14:32:42", modTime: time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, time.UTC), expected: true, desc: "before HH:mm:ss"},
		{exp: "14:32:42", modTime: time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC), expected: false, desc: "before HH:mm:ss - false"},

		{exp: "1999-08-27", after: true, modTime: time.Date(1985, 1, 20, 0, 0, 0, 0, time.UTC), expected: false, desc: "after YYYY-MM-DD"},
		{exp: "1999-08-27", after: true, modTime: time.Date(1999, 9, 11, 0, 0, 0, 0, time.UTC), expected: true, desc: "after YYYY-MM-DD - false"},
		{exp: "1999-08-27 14:32:42", after: true, modTime: time.Date(1999, 8, 27, 13, 0, 0, 0, time.UTC), expected: false, desc: "after YYYY-MM-DD HH:mm:ss"},
		{exp: "1999-08-27 14:32:42", after: true, modTime: time.Date(1999, 8, 27, 15, 0, 0, 0, time.UTC), expected: true, desc: "after YYYY-MM-DD HH:mm:ss - false"},
		{exp: "14:32:42", after: true, modTime: time.Date(now.Year(), now.Month(), now.Day(), 13, 0, 0, 0, time.UTC), expected: false, desc: "after HH:mm:ss"},
		{exp: "14:32:42", after: true, modTime: time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, time.UTC), expected: true, desc: "after HH:mm:ss - false"},

		{exp: "160s", modTime: now.Add(time.Second * -170), expected: true, desc: "before 160s"},
		{exp: "160s", modTime: now.Add(time.Second * -42), expected: false, desc: "before 160s - false"},
		{exp: "160m", modTime: now.Add(time.Minute * -170), expected: true, desc: "before 160m"},
		{exp: "160m", modTime: now.Add(time.Minute * -42), expected: false, desc: "before 160m - false"},
		{exp: "160h", modTime: now.Add(time.Hour * -170), expected: true, desc: "before 160h"},
		{exp: "160h", modTime: now.Add(time.Hour * -42), expected: false, desc: "before 160h - false"},

		{exp: "160D", modTime: now.AddDate(0, 0, -170), expected: true, desc: "before 160D"},
		{exp: "160D", modTime: now.AddDate(0, 0, -42), expected: false, desc: "before 160D - false"},
		{exp: "160M", modTime: now.AddDate(0, -170, 0), expected: true, desc: "before 160M"},
		{exp: "160M", modTime: now.AddDate(0, -42, 0), expected: false, desc: "before 160M - false"},
		{exp: "160Y", modTime: now.AddDate(-170, 0, 0), expected: true, desc: "before 160Y"},
		{exp: "160Y", modTime: now.AddDate(-42, 0, 0), expected: false, desc: "before 160Y - false"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var s search.Expression
			var err error
			if tC.after {
				s, err = search.NewModTimeAfter(tC.exp)
				require.NoError(t, err)
			} else {
				s, err = search.NewModTimeBefore(tC.exp)
				require.NoError(t, err)
			}
			m, err := s.Match(path.Info{ModTime: tC.modTime}, nil)
			require.NoError(t, err)
			assert.Equal(t, tC.expected, m)
		})
	}
}

func TestScanAndSearch(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-testing")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	scanCfg := scan.Config{
		CommonConfig: config.CommonConfig{
			Stdout: io.Discard,
			Stderr: io.Discard,
			DbPath: tempFile,
		},
		Root: "../../testdata/scan",
	}

	err := scan.Run(scanCfg)
	require.NoError(t, err)

	var outBuffer bytes.Buffer
	var errBuffer bytes.Buffer

	r1, err := search.NewRegex("b1a/blank")
	require.NoError(t, err)
	r2, err := search.NewRegex("^c/c.txt$")
	require.NoError(t, err)

	cfg := search.Config{
		CommonConfig: config.CommonConfig{
			Stdout: &outBuffer,
			Stderr: &errBuffer,
			DbPath: tempFile,
		},
		Expresion:      search.NewOr(r1, r2),
		DisplayMinimal: true,
	}

	err = search.Run(cfg)
	assert.NoError(t, err)

	result := make([]string, 0, 2)

	scanner := bufio.NewScanner(&outBuffer)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	assert.Equal(t, "", errBuffer.String())

	expected := []string{
		fmt.Sprintf("{%x}, %q", path.IdFromPath("b/b1/b1a/blank.txt"), "b/b1/b1a/blank.txt"),
		fmt.Sprintf("{%x}, %q", path.IdFromPath("c/c.txt"), "c/c.txt"),
	}

	slices.Sort(result)
	assert.Equal(t, expected, result)
}
