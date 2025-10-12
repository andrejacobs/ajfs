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
