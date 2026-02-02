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

package file

import (
	"bufio"
	"context"
	"crypto/md5"  // #nosec G501 -- MD5 is not used for cryptography
	"crypto/sha1" // #nosec G505 -- SHA1 is not used for cryptography
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"reflect"

	"github.com/andrejacobs/go-aj/file/contextio"
)

// Do buffered reads from rd and write to the hasher and optional io.Writer.
// Return the calculated hash and the total number of bytes copied.
func HashFromReader(ctx context.Context, rd io.Reader, hasher hash.Hash, w io.Writer) ([]byte, uint64, error) {
	r := contextio.NewReader(ctx, bufio.NewReader(rd))

	var dest io.Writer
	if (w != nil) && !reflect.ValueOf(w).IsNil() {
		dest = io.MultiWriter(hasher, w)
	} else {
		dest = hasher
	}

	count, err := io.Copy(dest, r)
	if err != nil {
		return nil, uint64(count), err
	}

	return hasher.Sum(nil), uint64(count), nil
}

// Hash the specified file and optionally copy the read bytes to the io.Writer.
// Return the calculated hash and the total number of bytes copied.
func Hash(ctx context.Context, path string, hasher hash.Hash, w io.Writer) ([]byte, uint64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to hash the file '%s'. %w", path, err)
	}
	defer f.Close()

	return HashFromReader(ctx, f, hasher, w)
}

func HashMD5(ctx context.Context, path string, w io.Writer) ([]byte, uint64, error) {
	return Hash(ctx, path, md5.New(), w) // #nosec G401 -- MD5 is not used for cryptography
}

func HashSHA1(ctx context.Context, path string, w io.Writer) ([]byte, uint64, error) {
	return Hash(ctx, path, sha1.New(), w) // #nosec G401 -- SHA1 is not used for cryptography
}

func HashSHA256(ctx context.Context, path string, w io.Writer) ([]byte, uint64, error) {
	return Hash(ctx, path, sha256.New(), w)
}

func HashSHA512(ctx context.Context, path string, w io.Writer) ([]byte, uint64, error) {
	return Hash(ctx, path, sha512.New(), w)
}
