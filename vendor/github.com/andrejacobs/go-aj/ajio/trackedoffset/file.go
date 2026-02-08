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
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// File wraps an os.File and keeps track of the current offset without requiring constant calls to Seek which involves syscall Lseek to be made.
// Reading and Writing is buffered by using the bufio package.
// Implements the following interfaces: io.Reader, io.Writer, io.Seeker, io.ByteReader.
type File struct {
	of     *os.File
	reader *bufio.Reader
	writer *bufio.Writer
	offset uint64
}

// Create a new File.
func NewFile(of *os.File) (*File, error) {
	f := &File{
		of:     of,
		reader: bufio.NewReader(of),
		writer: bufio.NewWriter(of),
	}

	if err := f.SyncOffset(); err != nil {
		return nil, err
	}

	return f, nil
}

// Close the file and release resources.
func (f *File) Close() error {
	err := f.of.Close()
	f.of = nil
	f.reader = nil
	f.writer = nil
	return err
}

// Path of the file.
func (f *File) Name() string {
	return f.of.Name()
}

// Return information describing the file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.of.Stat()
}

// io.Reader.
func (f *File) Read(p []byte) (int, error) {
	n, err := f.reader.Read(p)
	if err != nil {
		return n, err
	}

	newOffset, err := safe.Add64(f.offset, uint64(n))
	if err != nil {
		return 0, err
	}
	f.offset = newOffset

	return n, nil
}

// io.ByteReader.
func (f *File) ReadByte() (byte, error) {
	b, err := f.reader.ReadByte()
	if err != nil {
		return 0, err
	}

	newOffset, err := safe.Add64(f.offset, 1)
	if err != nil {
		return 0, err
	}
	f.offset = newOffset
	return b, nil
}

// Unreads the last byte. Only the most recently read byte can be unread.
// See [bufio.Reader. UnreadByte].
func (f *File) UnreadByte() error {
	if err := f.reader.UnreadByte(); err != nil {
		return err
	}

	newOffset, err := safe.Sub64(f.offset, 1)
	if err != nil {
		return err
	}
	f.offset = newOffset
	return nil
}

// Peek returns the next n bytes without advancing the offset or the reader.
// See [bufio.Reader.Peek].
func (f *File) Peek(n int) ([]byte, error) {
	return f.reader.Peek(n)
}

// Skips the next n bytes, returning the number of bytes discarded.
func (f *File) Discard(n int) (int, error) {
	rn, err := f.reader.Discard(n)
	if err != nil {
		return rn, err
	}

	newOffset, err := safe.Add64(f.offset, uint64(rn))
	if err != nil {
		return 0, err
	}
	f.offset = newOffset

	return rn, nil
}

// io.Writer.
func (f *File) Write(p []byte) (int, error) {
	n, err := f.writer.Write(p)
	if err != nil {
		return n, err
	}

	newOffset, err := safe.Add64(f.offset, uint64(n))
	if err != nil {
		return 0, err
	}
	f.offset = newOffset

	return n, nil
}

// io.ByteWriter.
func (f *File) WriteByte(c byte) error {
	if err := f.writer.WriteByte(c); err != nil {
		return err
	}

	newOffset, err := safe.Add64(f.offset, 1)
	if err != nil {
		return err
	}
	f.offset = newOffset
	return nil
}

// Reads a single UTF-8 encoded Unicode character and returns the
// rune and its size in bytes. If the encoded rune is invalid, it consumes one byte
// and returns unicode.ReplacementChar (U+FFFD) with a size of 1.
func (f *File) ReadRune() (rune, int, error) {
	r, s, err := f.reader.ReadRune()
	if err != nil {
		return r, s, err
	}

	newOffset, err := safe.Add64(f.offset, uint64(s))
	if err != nil {
		return r, s, err
	}
	f.offset = newOffset

	return r, s, nil
}

// Writes a single Unicode code point, returning
// the number of bytes written and any error.
func (f *File) WriteRune(r rune) (int, error) {
	n, err := f.writer.WriteRune(r)
	if err != nil {
		return n, err
	}

	newOffset, err := safe.Add64(f.offset, uint64(n))
	if err != nil {
		return 0, err
	}
	f.offset = newOffset

	return n, nil
}

// io.Seeker.
// It is recommended that you ResetReadBuffer or ResetWriteBuffer.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	newOffset, err := f.of.Seek(offset, whence)
	if err != nil {
		return newOffset, err
	}

	f.offset, err = safe.Int64ToUint64(newOffset)
	if err != nil {
		return newOffset, err
	}

	return newOffset, nil
}

// Return the current offset in bytes.
func (f *File) Offset() uint64 {
	return f.offset
}

// Set the current offset in bytes.
func (f *File) SetOffset(newOffset uint64) {
	f.offset = newOffset
}

// Ensure the File's offset and the underlying os.File's actual offsets are the same.
// This will make a call to file.Seek.
func (f *File) SyncOffset() error {
	offset, err := f.of.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}

	f.offset, err = safe.Int64ToUint64(offset)
	if err != nil {
		return err
	}

	return nil
}

// Ensure any unwritten data in the write buffer is written to the actual file.
func (f *File) Flush() error {
	return f.writer.Flush()
}

// Sync commits the current contents of the file to stable storage.
func (f *File) Sync() error {
	return f.of.Sync()
}

// Discard any buffered data that has not yet been read.
// Ensure you call this if you have changed the current offset using Seek.
func (f *File) ResetReadBuffer() {
	f.reader.Reset(f.of)
}

// Discard any buffered data that has not yet been written.
// Ensure you call this if you have changed the current offset using Seek.
func (f *File) ResetWriteBuffer() {
	f.writer.Reset(f.of)
}

// Guard rails removed ------

// Access the underlying os.File.
func (f *File) File() *os.File {
	return f.of
}

// Access the underlying bufio.Reader.
func (f *File) Reader() *bufio.Reader {
	return f.reader
}

// Access the underlying bufio.Writer.
func (f *File) Writer() *bufio.Writer {
	return f.writer
}

//-----------------------------------------------------------------------------
// Helpers

// Wraps [os.OpenFile] to allow for more options.
func OpenFile(path string, flag int, perm os.FileMode) (*File, error) {
	f, err := os.OpenFile(path, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	file, err := NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	return file, nil
}

// Open a file for reading.
// Wraps [os.Open].
func Open(path string) (*File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	file, err := NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	return file, nil
}

// Create a file for reading and writing.
// Wraps [os.Create].
func Create(path string) (*File, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	file, err := NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("failed to open the file %q. %w", path, err)
	}

	return file, nil
}
