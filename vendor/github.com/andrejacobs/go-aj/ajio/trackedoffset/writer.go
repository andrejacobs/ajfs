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

package trackedoffset

import (
	"io"

	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// Writer keeps track of the offset within an io.Writer source.
type Writer struct {
	wd     io.Writer
	offset uint64
}

// Create a new Writer that will keep track of the offset within the source io.Writer.
func NewWriter(wd io.Writer, baseOffset uint64) *Writer {
	w := &Writer{
		wd:     wd,
		offset: baseOffset,
	}
	return w
}

// Writer implementation.
func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.wd.Write(p)
	if err != nil {
		return n, err
	}

	newOffset, err := safe.Add64(w.offset, uint64(n))
	if err != nil {
		return 0, err
	}
	w.offset = newOffset

	return n, nil
}

// Return the current offset in bytes.
func (w *Writer) Offset() uint64 {
	return w.offset
}

// Set the known offset in bytes.
func (w *Writer) ResetOffset(offset uint64) {
	w.offset = offset
}
