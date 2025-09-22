package scan_test

import (
	"crypto/sha1"
	"testing"

	"github.com/andrejacobs/ajfs/internal/scan"
	"github.com/stretchr/testify/assert"
)

func TestIdFromPath(t *testing.T) {
	id := scan.IdFromPath("/usr/bin")
	assert.Equal(t, scan.PathId(sha1.Sum([]byte("/usr/bin"))), id)
}
