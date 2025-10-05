package export_test

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrejacobs/ajfs/internal/app/config"
	"github.com/andrejacobs/ajfs/internal/app/export"
	"github.com/andrejacobs/ajfs/internal/db"
	"github.com/andrejacobs/ajfs/internal/path"
	"github.com/andrejacobs/go-aj/ajhash"
	"github.com/andrejacobs/go-aj/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportCSV(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.csv")
	_ = os.Remove(tempExportFile)
	defer os.Remove(tempExportFile)

	expected := expectedDatabase(t, tempFile, false)
	expectedF, err := os.CreateTemp("", "unit-test.ajfs.expected.csv")
	require.NoError(t, err)
	defer os.Remove(expectedF.Name())

	csvWriter := csv.NewWriter(expectedF)
	csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "IsDir", "Path"})

	for _, exp := range expected {
		csvWriter.Write([]string{
			fmt.Sprintf("%x", exp.pi.Id),
			fmt.Sprintf("%d", exp.pi.Size),
			exp.pi.Mode.String(),
			exp.pi.ModTime.Format(time.RFC3339Nano),
			fmt.Sprintf("%t", exp.pi.IsDir()),
			exp.pi.Path,
		})
	}

	csvWriter.Flush()
	require.NoError(t, csvWriter.Error())
	require.NoError(t, expectedF.Close())

	cfg := export.Config{
		CommonConfig: config.CommonConfig{
			DbPath: tempFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Format:     export.FormatCSV,
		ExportPath: tempExportFile,
	}

	require.NoError(t, export.Run(cfg))

	simpleDiff(t, expectedF.Name(), tempExportFile)
}

func TestExportWithHashesCSV(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.csv")
	_ = os.Remove(tempExportFile)
	defer os.Remove(tempExportFile)

	expected := expectedDatabase(t, tempFile, true)
	expectedF, err := os.CreateTemp("", "unit-test.ajfs.expected.csv")
	require.NoError(t, err)
	defer os.Remove(expectedF.Name())

	csvWriter := csv.NewWriter(expectedF)
	csvWriter.Write([]string{"Id", "Size", "Mode", "ModTime", "IsDir", "Hash (" + ajhash.AlgoSHA1.String() + ")", "Path"})

	for _, exp := range expected {
		hashStr := hex.EncodeToString(exp.hash)

		csvWriter.Write([]string{
			fmt.Sprintf("%x", exp.pi.Id),
			fmt.Sprintf("%d", exp.pi.Size),
			exp.pi.Mode.String(),
			exp.pi.ModTime.Format(time.RFC3339Nano),
			fmt.Sprintf("%t", exp.pi.IsDir()),
			hashStr,
			exp.pi.Path,
		})
	}

	csvWriter.Flush()
	require.NoError(t, csvWriter.Error())
	require.NoError(t, expectedF.Close())

	cfg := export.Config{
		CommonConfig: config.CommonConfig{
			DbPath: tempFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Format:     export.FormatCSV,
		ExportPath: tempExportFile,
	}

	require.NoError(t, export.Run(cfg))

	simpleDiff(t, expectedF.Name(), tempExportFile)
}

//-----------------------------------------------------------------------------

type JsonEntry struct {
	Id   string `json:"id"`
	Path string `json:"path"`

	Size    uint64      `json:"size"`
	Mode    fs.FileMode `json:"mode"`
	ModeStr string      `json:"modeStr"`
	ModTime time.Time   `json:"modTime"`

	Hash string `json:"hash,omitempty"`
}

type JsonDatabase struct {
	Version          int             `json:"version"`
	DbPath           string          `json:"dbPath"`
	Root             string          `json:"root"`
	Features         db.FeatureFlags `json:"features"`
	EntriesCount     int             `json:"entriesCount"`
	FileEntriesCount int             `json:"fileCount"`
	Meta             db.MetaEntry    `json:"meta"`
	HashTableAlgo    string          `json:"hashTableAlgo,omitempty"`
}

func TestExportJSON(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.json")
	_ = os.Remove(tempExportFile)
	defer os.Remove(tempExportFile)

	expectedDatabase(t, tempFile, false)

	dbf, err := db.OpenDatabase(tempFile)
	require.NoError(t, err)

	jsonDb := JsonDatabase{
		Version:          dbf.Version(),
		DbPath:           dbf.Path(),
		Root:             dbf.RootPath(),
		Features:         dbf.Features(),
		EntriesCount:     dbf.EntriesCount(),
		FileEntriesCount: dbf.FileEntriesCount(),
		Meta:             dbf.Meta(),
	}

	entries := make([]JsonEntry, 0, dbf.EntriesCount())

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		entry := JsonEntry{
			Id:      hex.EncodeToString(pi.Id[:]),
			Path:    pi.Path,
			Size:    pi.Size,
			Mode:    pi.Mode,
			ModeStr: pi.Mode.String(),
			ModTime: pi.ModTime,
		}
		entries = append(entries, entry)
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	expectedF, err := os.CreateTemp("", "unit-test.ajfs.expected.json")
	require.NoError(t, err)
	defer os.Remove(expectedF.Name())

	encoder := json.NewEncoder(expectedF)
	encoder.SetIndent("", "\t")

	actual := struct {
		Database JsonDatabase `json:"database"`
		Entries  []JsonEntry  `json:"entries"`
	}{
		Database: jsonDb,
		Entries:  entries,
	}

	require.NoError(t, encoder.Encode(&actual))
	require.NoError(t, expectedF.Close())

	cfg := export.Config{
		CommonConfig: config.CommonConfig{
			DbPath: tempFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Format:     export.FormatJSON,
		ExportPath: tempExportFile,
	}

	require.NoError(t, export.Run(cfg))

	simpleDiff(t, expectedF.Name(), tempExportFile)
}

func TestExportWithHashesJSON(t *testing.T) {
	tempFile := filepath.Join(os.TempDir(), "unit-test.ajfs")
	_ = os.Remove(tempFile)
	defer os.Remove(tempFile)

	tempExportFile := filepath.Join(os.TempDir(), "unit-test.ajfs.json")
	_ = os.Remove(tempExportFile)
	defer os.Remove(tempExportFile)

	expectedDatabase(t, tempFile, true)

	dbf, err := db.OpenDatabase(tempFile)
	require.NoError(t, err)

	algo, err := dbf.HashTableAlgo()
	require.NoError(t, err)

	jsonDb := JsonDatabase{
		Version:          dbf.Version(),
		DbPath:           dbf.Path(),
		Root:             dbf.RootPath(),
		Features:         dbf.Features(),
		EntriesCount:     dbf.EntriesCount(),
		FileEntriesCount: dbf.FileEntriesCount(),
		Meta:             dbf.Meta(),
		HashTableAlgo:    algo.String(),
	}

	entries := make([]JsonEntry, 0, dbf.EntriesCount())

	hashTable, err := dbf.ReadHashTable()
	require.NoError(t, err)

	err = dbf.ReadAllEntries(func(idx int, pi path.Info) error {
		var hashStr string
		if !pi.IsDir() {
			hash, ok := hashTable[idx]

			if ok {
				hashStr = hex.EncodeToString(hash)
			}
		}

		entry := JsonEntry{
			Id:      hex.EncodeToString(pi.Id[:]),
			Path:    pi.Path,
			Size:    pi.Size,
			Mode:    pi.Mode,
			ModeStr: pi.Mode.String(),
			ModTime: pi.ModTime,
			Hash:    hashStr,
		}
		entries = append(entries, entry)
		return nil
	})
	require.NoError(t, err)
	require.NoError(t, dbf.Close())

	expectedF, err := os.CreateTemp("", "unit-test.ajfs.expected.json")
	require.NoError(t, err)
	defer os.Remove(expectedF.Name())

	encoder := json.NewEncoder(expectedF)
	encoder.SetIndent("", "\t")

	actual := struct {
		Database JsonDatabase `json:"database"`
		Entries  []JsonEntry  `json:"entries"`
	}{
		Database: jsonDb,
		Entries:  entries,
	}

	require.NoError(t, encoder.Encode(&actual))
	require.NoError(t, expectedF.Close())

	cfg := export.Config{
		CommonConfig: config.CommonConfig{
			DbPath: tempFile,
			Stdout: io.Discard,
			Stderr: io.Discard,
		},
		Format:     export.FormatJSON,
		ExportPath: tempExportFile,
	}

	require.NoError(t, export.Run(cfg))

	simpleDiff(t, expectedF.Name(), tempExportFile)
}

//-----------------------------------------------------------------------------

type expectedEntry struct {
	pi   path.Info
	hash []byte
}

func expectedDatabase(t *testing.T, dbPath string, hashes bool) []expectedEntry {
	algo := ajhash.AlgoSHA1

	features := db.FeatureJustEntries
	if hashes {
		features |= db.FeatureHashTable
	}

	dbf, err := db.CreateDatabase(dbPath, "/test/", db.FeatureFlags(features))
	require.NoError(t, err)

	p1 := path.Info{
		Id:      path.IdFromPath("a.txt"),
		Path:    "a.txt",
		Size:    uint64(42),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p1))

	p2 := path.Info{
		Id:      path.IdFromPath("some/dir"),
		Path:    "some/dir",
		Size:    uint64(142),
		Mode:    0644 | fs.ModeDir,
		ModTime: time.Now().Add(-20 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p2))

	p3 := path.Info{
		Id:      path.IdFromPath("c.txt"),
		Path:    "c.txt",
		Size:    uint64(442),
		Mode:    0740,
		ModTime: time.Now().Add(-10 * time.Minute),
	}
	require.NoError(t, dbf.WriteEntry(&p3))

	require.NoError(t, dbf.FinishEntries())

	var (
		h1 []byte
		h3 []byte
	)

	if hashes {
		require.NoError(t, dbf.StartHashTable(algo))
		require.NoError(t, dbf.FinishHashTable())

		h1 = algo.Buffer()
		require.NoError(t, random.SecureBytes(h1))
		dbf.WriteHashEntry(0, h1)

		h3 = algo.Buffer()
		require.NoError(t, random.SecureBytes(h3))
		dbf.WriteHashEntry(2, h3)
	}

	require.NoError(t, dbf.Close())

	if hashes {
		return []expectedEntry{
			{
				pi:   p1,
				hash: h1,
			},
			{
				pi: p2,
			},
			{
				pi:   p3,
				hash: h3,
			},
		}
	}
	return []expectedEntry{
		{
			pi: p1,
		},
		{
			pi: p2,
		},
		{
			pi: p3,
		},
	}
}

func simpleDiff(t *testing.T, fileA string, fileB string) {
	li, err := os.Stat(fileA)
	require.NoError(t, err)

	ri, err := os.Stat(fileB)
	require.NoError(t, err)

	require.Equal(t, li.Size(), ri.Size())

	l, err := os.Open(fileA)
	require.NoError(t, err)
	defer l.Close()

	r, err := os.Open(fileB)
	require.NoError(t, err)
	defer r.Close()

	ls := bufio.NewScanner(l)
	rs := bufio.NewScanner(r)

	line := 0
	for {
		if ls.Scan() && rs.Scan() {
			line++
			require.NoError(t, ls.Err())
			require.NoError(t, rs.Err())

			assert.Equal(t, ls.Text(), rs.Text(), fmt.Sprintf("line: %d", line))

		} else {
			if ls.Err() != nil || rs.Err() != nil {
				require.Fail(t, fmt.Sprintf("failed to read from both left and right side. line: %d", line))
			}
			break
		}
	}
}
