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
	"bytes"
	"crypto/sha1" // #nosec G505 -- SHA1 is not used for cryptography
	"sort"
)

const (
	PathHashSize = sha1.Size
)

type PathHash [PathHashSize]byte

// Calculate the unique hash for a path.
func CalculatePathHash(path string) PathHash {
	return sha1.Sum([]byte(path)) // #nosec G401 -- Not used for crypto
}

// Calculate the unique hash for a given slice of file paths.
func CalculatePathsHash(paths []string) (PathHash, error) {
	// Using sha1 since I need a hash that is consistent (maphash is great but requires to store the seed value)
	// sha1 turns out to be faster on the Intel CPU I intend to mainly run this code on
	// sha256 is slightly faster on my M2 Macbook
	// To test: openssl speed md5 sha1 sha256
	sorted := append([]string{}, paths...)
	sort.Strings(sorted)

	var buf bytes.Buffer
	for _, p := range sorted {
		if _, err := buf.WriteString(p); err != nil {
			return PathHash{}, err
		}
	}

	return sha1.Sum(buf.Bytes()), nil // #nosec G401 -- Not used for crypto
}
