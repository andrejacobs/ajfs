// Copyright (c) 2025 Andre Jacobs
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package vardata is used to read or write variable length data.
package vardata

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"golang.org/x/exp/constraints"
)

// Need to be able to do io.Reader.Read as well as io.ByteReader.ReadByte
// bufio.NewReader implements this interface.
type Reader interface {
	io.Reader
	io.ByteReader
}

// VariableDataFixedLen is used to read and write variable sized data from an io.Reader or io.Writer.
// Data written to an io.Writer is first prefixed with the size of the data to be written.
// The type parameter S determines which size of integer will be used for the prefix.
// For example: VariableDataFixedLen[uint8] will use 1 byte for the size and thus a maximum of 256 bytes may be read or written.
// The default byte order is little endian.
type VariableDataFixedLen[S constraints.Unsigned] struct {
	size     int // the size in bytes used for the data prefix
	maxValue S   // the maximum size of the integer
	order    binary.ByteOrder
	write    writeFunc
	read     readFunc
}

// Create a new VariableData instance that will use 1 byte for the data prefix size.
func NewVariableDataUint8() VariableDataFixedLen[uint8] {
	return VariableDataFixedLen[uint8]{size: 1, maxValue: math.MaxUint8, order: binary.LittleEndian, write: writeUint8, read: readUint8}
}

// Create a new VariableData instance that will use 2 bytes for the data prefix size.
func NewVariableDataUint16() VariableDataFixedLen[uint16] {
	return VariableDataFixedLen[uint16]{size: 2, maxValue: math.MaxUint16, order: binary.LittleEndian, write: writeUint16, read: readUint16}
}

// Create a new VariableData instance that will use 4 bytes for the data prefix size.
func NewVariableDataUint32() VariableDataFixedLen[uint32] {
	return VariableDataFixedLen[uint32]{size: 4, maxValue: math.MaxUint32, order: binary.LittleEndian, write: writeUint32, read: readUint32}
}

// Create a new VariableData instance that will use 8 bytes for the data prefix size.
func NewVariableDataUint64() VariableDataFixedLen[uint64] {
	return VariableDataFixedLen[uint64]{size: 8, maxValue: math.MaxUint64, order: binary.LittleEndian, write: writeUint64, read: readUint64}
}

// Use little endianess.
func (v VariableDataFixedLen[S]) LittleEndian() VariableDataFixedLen[S] {
	v.order = binary.LittleEndian
	return v
}

// Use big endianess.
func (v VariableDataFixedLen[S]) BigEndian() VariableDataFixedLen[S] {
	v.order = binary.BigEndian
	return v
}

// Use the native platform's endianess. Not recommended.
func (v VariableDataFixedLen[S]) NativeEndian() VariableDataFixedLen[S] {
	v.order = binary.NativeEndian
	return v
}

// Return the maximum number of bytes that can be read or written to for the data prefix.
func (v VariableDataFixedLen[S]) MaxSize() S {
	return v.maxValue
}

// The number of bytes that will be used to write the data prefix. e.g. Uint16 will use 2 bytes.
func (v VariableDataFixedLen[S]) PrefixSize() int {
	return v.size
}

// Return the byte order (endianess) used.
func (v VariableDataFixedLen[S]) ByteOrder() binary.ByteOrder {
	return v.order
}

// Write the size of the data (i.e len(data)) followed by that data itself.
// Returns the number of bytes written including the size of the prefix.
func (v VariableDataFixedLen[S]) Write(w io.Writer, data []byte) (int, error) {
	dataLen := len(data)
	if uint64(dataLen) > uint64(v.maxValue) {
		return 0, fmt.Errorf("failed to write data of size %d. maximum size allowed is %d", dataLen, v.maxValue)
	}

	return v.write(w, data, dataLen, v.order)
}

// Read the size of the data followed by that amount of bytes into the provided buffer.
// A new buffer will be allocated if the provided one is not large enough to hold the data.
// Returns the buffer and the number of bytes read including the size of the prefix.
func (v VariableDataFixedLen[S]) Read(r io.Reader, buffer []byte) ([]byte, int, error) {
	return v.read(r, buffer, v.order)
}

// Write a string using the generic Write method to prefix the length of the string first and reducing allocs.
func (v VariableDataFixedLen[S]) WriteString(w io.Writer, data string) (int, error) {
	dataLen := len(data)
	if uint64(dataLen) > uint64(v.maxValue) {
		return 0, fmt.Errorf("failed to write string of size %d. maximum size allowed is %d", dataLen, v.maxValue)
	}

	var err error
	dt := any(v.maxValue)
	switch dt.(type) {
	case uint8:
		err = binary.Write(w, v.order, uint8(dataLen))
	case uint16:
		err = binary.Write(w, v.order, uint16(dataLen))
	case uint32:
		err = binary.Write(w, v.order, uint32(dataLen))
	case uint64:
		err = binary.Write(w, v.order, uint64(dataLen))
	default:
		panic("unsupported parameter type")
	}
	if err != nil {
		return 0, fmt.Errorf("failed to write string's length. %w", err)
	}

	n, err := io.WriteString(w, data) // more efficient if we know upfront it is a string. This avoids allocs
	return n + v.size, err
}

// Read a string.
// NOTE: If you are going to be reading a lot of strings then it is better to use the generic Read method
// and passing in a pre-allocated []byte.
func (v VariableDataFixedLen[S]) ReadString(r io.Reader) (string, int, error) {
	data, rcount, err := v.Read(r, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read a string. %w", err)
	}

	return string(data), rcount, err
}

//-----------------------------------------------------------------------------

// VariableData is used to read and write variable sized data from an io.Reader or io.Writer and the
// size prefix uses a varint to record the number of bytes stored (1 byte minimum to 10 bytes maximum)
// Data written to an io.Writer is first prefixed with the size of the data to be written.
// The default byte order is little endian.
type VariableData struct {
	order binary.ByteOrder
}

// Create a new VariableDataVarInt instance that will use between 1 and 10 bytes for the data prefix size.
func NewVariableData() VariableData {
	return VariableData{order: binary.LittleEndian}
}

// Return the byte order (endianess) used.
func (v VariableData) ByteOrder() binary.ByteOrder {
	return v.order
}

// Write the size of the data (i.e len(data)) followed by that data itself.
// Returns the number of bytes written including the size of the prefix.
func (v VariableData) Write(w io.Writer, data []byte) (int, error) {
	dataLen := len(data)

	varintBuf := make([]byte, binary.MaxVarintLen64)
	varintSize := binary.PutUvarint(varintBuf[:], uint64(dataLen))
	varintBuf = varintBuf[:varintSize]
	_, err := w.Write(varintBuf)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return n + varintSize, err
}

// Read the size of the data followed by that amount of bytes into the provided buffer.
// A new buffer will be allocated if the provided one is not large enough to hold the data.
// Returns the buffer and the number of bytes read including the size of the prefix.
func (v VariableData) Read(r Reader, buffer []byte) ([]byte, int, error) {
	dataLen, varintSize, err := v.readUvarint(r)
	if err != nil {
		return nil, varintSize, err
	}

	if cap(buffer) < int(dataLen) {
		buffer = make([]byte, dataLen)
	} else {
		buffer = buffer[:dataLen]
	}

	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return nil, n, fmt.Errorf("failed to read the expected size %d of data. %w", dataLen, err)
	}

	return buffer, n + varintSize, nil
}

// Write a string using the generic Write method to prefix the length of the string first and reducing allocs.
func (v VariableData) WriteString(w io.Writer, data string) (int, error) {
	dataLen := len(data)

	varintBuf := make([]byte, binary.MaxVarintLen64)
	varintSize := binary.PutUvarint(varintBuf[:], uint64(dataLen))
	varintBuf = varintBuf[:varintSize]
	_, err := w.Write(varintBuf)
	if err != nil {
		return 0, err
	}

	n, err := io.WriteString(w, data) // more efficient if we know upfront it is a string. This avoids allocs
	return n + varintSize, err
}

// Read a string.
// NOTE: If you are going to be reading a lot of strings then it is better to use the generic Read method
// and passing in a pre-allocated []byte.
func (v VariableData) ReadString(r Reader) (string, int, error) {
	data, rcount, err := v.Read(r, nil)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read a string. %w", err)
	}

	return string(data), rcount, err
}

// NOTE: Taken from encoding/binary/varint.go and modified to return the number of bytes read
// >>> ReadUvarint reads an encoded unsigned integer from r and returns it as a uint64.
// >>> The error is EOF only if no bytes were read.
// >>> If an EOF happens after reading some but not all the bytes,
// >>> ReadUvarint returns io.ErrUnexpectedEOF.
func (v VariableData) readUvarint(r io.ByteReader) (uint64, int, error) {
	var x uint64
	var s uint
	var i int
	for i = 0; i < binary.MaxVarintLen64; i++ {
		b, err := r.ReadByte()
		if err != nil {
			if i > 0 && err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
			return x, i + 1, err
		}
		if b < 0x80 {
			if i == binary.MaxVarintLen64-1 && b > 1 {
				return x, i + 1, errOverflow
			}
			return x | uint64(b)<<s, i + 1, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return x, i + 1, errOverflow
}

var errOverflow = errors.New("binary: varint overflows a 64-bit integer")

//-----------------------------------------------------------------------------

type writeFunc func(w io.Writer, data []byte, count int, order binary.ByteOrder) (int, error)
type readFunc func(r io.Reader, buffer []byte, order binary.ByteOrder) ([]byte, int, error)

func writeUint8(w io.Writer, data []byte, count int, order binary.ByteOrder) (int, error) {
	if err := binary.Write(w, order, uint8(count)); err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return n + 1, err
}

func writeUint16(w io.Writer, data []byte, count int, order binary.ByteOrder) (int, error) {
	if err := binary.Write(w, order, uint16(count)); err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return n + 2, err
}

func writeUint32(w io.Writer, data []byte, count int, order binary.ByteOrder) (int, error) {
	if err := binary.Write(w, order, uint32(count)); err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return n + 4, err
}

func writeUint64(w io.Writer, data []byte, count int, order binary.ByteOrder) (int, error) {
	if err := binary.Write(w, order, uint64(count)); err != nil {
		return 0, err
	}

	n, err := w.Write(data)
	return n + 8, err
}

func readUint8(r io.Reader, buffer []byte, order binary.ByteOrder) ([]byte, int, error) {
	var count uint8
	if err := binary.Read(r, order, &count); err != nil {
		return nil, 0, fmt.Errorf("failed to read the size of the data. %w", err)
	}

	if cap(buffer) < int(count) {
		buffer = make([]byte, count)
	} else {
		buffer = buffer[:count]
	}

	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return nil, n, fmt.Errorf("failed to read the expected size %d of data. %w", count, err)
	}

	return buffer, n + 1, nil
}

func readUint16(r io.Reader, buffer []byte, order binary.ByteOrder) ([]byte, int, error) {
	var count uint16
	if err := binary.Read(r, order, &count); err != nil {
		return nil, 0, fmt.Errorf("failed to read the size of the data. %w", err)
	}

	if cap(buffer) < int(count) {
		buffer = make([]byte, count)
	} else {
		buffer = buffer[:count]
	}

	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return nil, n, fmt.Errorf("failed to read the expected size %d of data. %w", count, err)
	}

	return buffer, n + 2, nil
}

func readUint32(r io.Reader, buffer []byte, order binary.ByteOrder) ([]byte, int, error) {
	var count uint32
	if err := binary.Read(r, order, &count); err != nil {
		return nil, 0, fmt.Errorf("failed to read the size of the data. %w", err)
	}

	if cap(buffer) < int(count) {
		buffer = make([]byte, count)
	} else {
		buffer = buffer[:count]
	}

	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return nil, n, fmt.Errorf("failed to read the expected size %d of data. %w", count, err)
	}

	return buffer, n + 4, nil
}

func readUint64(r io.Reader, buffer []byte, order binary.ByteOrder) ([]byte, int, error) {
	var count uint64
	if err := binary.Read(r, order, &count); err != nil {
		return nil, 0, fmt.Errorf("failed to read the size of the data. %w", err)
	}

	if cap(buffer) < int(count) {
		buffer = make([]byte, count)
	} else {
		buffer = buffer[:count]
	}

	n, err := io.ReadFull(r, buffer)
	if err != nil {
		return nil, n, fmt.Errorf("failed to read the expected size %d of data. %w", count, err)
	}

	return buffer, n + 8, nil
}
