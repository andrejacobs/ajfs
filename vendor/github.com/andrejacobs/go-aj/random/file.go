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

package random

import (
	"crypto/rand"
	"io"
	"os"
)

// Create a file and fill it with random bytes.
// NOTE: This will override any existing file.
// path The path of the file to be created.
// size The number of random bytes to write to the file.
func CreateFile(path string, size int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.CopyN(f, rand.Reader, size)
	if err != nil {
		return err
	}

	return nil
}

// Create a temporary file and fill it with random bytes.
// NOTE: This will override any existing file.
// See os.CreateTemp for details on dir and pattern.
// size The number of random bytes to write to the file.
// Returns the path to the file that was created.
func CreateTempFile(dir, pattern string, size int64) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.CopyN(f, rand.Reader, size)
	if err != nil {
		return "", err
	}

	return f.Name(), nil
}
