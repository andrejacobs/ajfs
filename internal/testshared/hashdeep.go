package testshared

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// A single entry in the hashdeep file.
type HashDeepEntry struct {
	FileSize int
	Hash     string
	Path     string
}

// Parse a hashdeep file.
func ReadHashDeepFile(path string) ([]HashDeepEntry, error) {
	result := make([]HashDeepEntry, 0, 32)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			continue
		}
		if strings.HasPrefix(text, "%%") {
			continue
		}
		if strings.HasPrefix(text, "##") {
			continue
		}

		entry := HashDeepEntry{}

		parts := strings.Split(text, ",")
		if len(parts) != 3 {
			return nil, fmt.Errorf("failed to parse the line: %s", text)
		}
		entry.FileSize, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, err
		}
		entry.Hash = parts[1]
		entry.Path = strings.TrimPrefix(parts[2], "./")
		result = append(result, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
