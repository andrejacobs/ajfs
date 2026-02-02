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

// Package safe provides safe casting operations for different data types.
package safe

import (
	"errors"
	"math"
	"math/bits"
)

var (
	// An integer overflow occurred.
	ErrIntegerOverflow = errors.New("integer overflow occurred")

	// An integer underflow occurred.
	ErrIntegerUnderflow = errors.New("integer underflow occurred")
)

// IntSize is the size in bits of an int or uint value on the running platform. Either 32 or 64.
const IntSize = 32 << (^uint(0) >> 63) // taken from strconv.IntSize and math.intSize

// Add two unsigned 32bit integers.
// Returns [ErrIntegerOverflow] if an overflow occurred.
func Add32(x, y uint32) (uint32, error) {
	new, overflowed := bits.Add32(x, y, 0)
	if overflowed > 0 {
		return 0, ErrIntegerOverflow
	}
	return new, nil
}

// Add two unsigned 64bit integers.
// Returns [ErrIntegerOverflow] if an overflow occurred.
func Add64(x, y uint64) (uint64, error) {
	new, overflowed := bits.Add64(x, y, 0)
	if overflowed > 0 {
		return 0, ErrIntegerOverflow
	}
	return new, nil
}

// Subtract two unsigned 32bit integers.
// Returns [ErrIntegerUnderflow] if an underflow occurred.
func Sub32(x, y uint32) (uint32, error) {
	new, underflowed := bits.Sub32(x, y, 0)
	if underflowed > 0 {
		return 0, ErrIntegerUnderflow
	}
	return new, nil
}

// Subtract two unsigned 64bit integers.
// Returns [ErrIntegerUnderflow] if an underflow occurred.
func Sub64(x, y uint64) (uint64, error) {
	new, underflowed := bits.Sub64(x, y, 0)
	if underflowed > 0 {
		return 0, ErrIntegerUnderflow
	}
	return new, nil
}

//-----------------------------------------------------------------------------
// Safe casting

// Cast from a signed 8bit integer to an unsigned 8bit integer.
// Return [ErrIntegerUnderflow] if x contains a negative number.
func Int8ToUint8(x int8) (uint8, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	}
	return uint8(x), nil
}

// Cast from an unsigned 8bit integer to a signed 8bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint8ToInt8(x uint8) (int8, error) {
	if x > math.MaxInt8 {
		return 0, ErrIntegerOverflow
	}
	return int8(x), nil
}

// Cast from a signed 16bit integer to an unsigned 16bit integer.
// Return [ErrIntegerUnderflow] if x contains a negative number.
func Int16ToUint16(x int16) (uint16, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	}
	return uint16(x), nil
}

// Cast from an unsigned 16bit integer to a signed 16bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint16ToInt16(x uint16) (int16, error) {
	if x > math.MaxInt16 {
		return 0, ErrIntegerOverflow
	}
	return int16(x), nil
}

// Cast from a signed 32bit integer to an unsigned 32bit integer.
// Return [ErrIntegerUnderflow] if x contains a negative number.
func Int32ToUint32(x int32) (uint32, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	}
	return uint32(x), nil
}

// Cast from an unsigned 32bit integer to a signed 32bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint32ToInt32(x uint32) (int32, error) {
	if x > math.MaxInt32 {
		return 0, ErrIntegerOverflow
	}
	return int32(x), nil
}

// Cast from a signed 64bit integer to an unsigned 64bit integer.
// Return [ErrIntegerUnderflow] if x contains a negative number.
func Int64ToUint64(x int64) (uint64, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	}
	return uint64(x), nil
}

// Cast from an unsigned 64bit integer to a signed 64bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint64ToInt64(x uint64) (int64, error) {
	if x > math.MaxInt64 {
		return 0, ErrIntegerOverflow
	}
	return int64(x), nil
}

//-----------------------------------------------------------------------------
// Downcasting

// Downcast an unsigned 64bit integer to an unsigned 32bit integer.
// Returns [ErrIntegerOverflow] if an overflow occurred.
func Uint64ToUint32(x uint64) (uint32, error) {
	if x > math.MaxUint32 {
		return 0, ErrIntegerOverflow
	}
	return uint32(x), nil
}

// Downcast a signed 64bit integer to a signed 32bit integer.
// Returns [ErrIntegerOverflow] if an overflow occurred.
func Int64ToInt32(x int64) (int32, error) {
	if x > math.MaxInt32 {
		return 0, ErrIntegerOverflow
	}
	return int32(x), nil
}

// Downcast a signed 64bit integer to an unsigned 32bit integer.
// Returns [ErrIntegerUnderflow] if x is negative.
// Returns [ErrIntegerOverflow] if x is too big.
func Int64ToUint32(x int64) (uint32, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxUint32 {
		return 0, ErrIntegerOverflow
	}
	return uint32(x), nil
}

// Downcast an unsigned 64bit integer to a signed 32bit integer.
// Returns [ErrIntegerOverflow] if x is too big.
func Uint64ToInt32(x uint64) (int32, error) {
	if x > math.MaxInt32 {
		return 0, ErrIntegerOverflow
	}
	return int32(x), nil
}

// Cast from platform dependant signed integer to a signed 8bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
// Return [ErrIntegerOverflow] if x is too big.
func IntToInt8(x int) (int8, error) {
	if x < math.MinInt8 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxInt8 {
		return 0, ErrIntegerOverflow
	}
	return int8(x), nil
}

// Cast from platform dependant signed integer to a signed 16bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
// Return [ErrIntegerOverflow] if x is too big.
func IntToInt16(x int) (int16, error) {
	if x < math.MinInt16 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxInt16 {
		return 0, ErrIntegerOverflow
	}
	return int16(x), nil
}

// Cast from platform dependant signed integer to a signed 32bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
// Return [ErrIntegerOverflow] if x is too big.
func IntToInt32(x int) (int32, error) {
	if x < math.MinInt32 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxInt32 {
		return 0, ErrIntegerOverflow
	}
	return int32(x), nil
}

// Cast from platform dependant signed integer to an unsigned 8bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
// Return [ErrIntegerOverflow] if x is too big.
func IntToUint8(x int) (uint8, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxUint8 {
		return 0, ErrIntegerOverflow
	}
	return uint8(x), nil
}

// Cast from platform dependant signed integer to an unsigned 16bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
// Return [ErrIntegerOverflow] if x is too big.
func IntToUint16(x int) (uint16, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	} else if x > math.MaxUint16 {
		return 0, ErrIntegerOverflow
	}
	return uint16(x), nil
}

//func IntToUint32(x int) (uint32, error) [see safe_32bit.go & safe_64bit.go]

// Cast from platform dependant signed integer to an unsigned 64bit integer.
// Return [ErrIntegerUnderflow] if x is too small.
func IntToUint64(x int) (uint64, error) {
	if x < 0 {
		return 0, ErrIntegerUnderflow
	}
	return uint64(x), nil
}

// Cast from platform dependant unsigned integer to an unsigned 8bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func UintToUint8(x uint) (uint8, error) {
	if x > math.MaxUint8 {
		return 0, ErrIntegerOverflow
	}
	return uint8(x), nil
}

// Cast from platform dependant unsigned integer to an unsigned 16bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func UintToUint16(x uint) (uint16, error) {
	if x > math.MaxUint16 {
		return 0, ErrIntegerOverflow
	}
	return uint16(x), nil
}

// Cast from platform dependant unsigned integer to an unsigned 32bit integer.
// Return [ErrIntegerOverflow] if x is too big.
func UintToUint32(x uint) (uint32, error) {
	if x > math.MaxUint32 {
		return 0, ErrIntegerOverflow
	}
	return uint32(x), nil
}

// Cast from unsigned 32bit integer to platform dependant signed integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint32ToInt(x uint32) (int, error) {
	if (IntSize == 32) && (x > math.MaxInt32) {
		return 0, ErrIntegerOverflow
	}

	return int(x), nil
}

// Cast from unsigned 64bit integer to platform dependant signed integer.
// Return [ErrIntegerOverflow] if x is too big.
func Uint64ToInt(x uint64) (int, error) {
	if (IntSize == 32) && (x > math.MaxInt32) {
		return 0, ErrIntegerOverflow
	} else if x > math.MaxInt64 {
		return 0, ErrIntegerOverflow
	}

	return int(x), nil
}
