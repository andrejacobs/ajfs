package testshared

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func SimpleDiff(t *testing.T, fileA string, fileB string) {
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
