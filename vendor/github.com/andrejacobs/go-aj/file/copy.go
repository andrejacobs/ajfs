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
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/andrejacobs/go-aj/file/contextio"
)

// Copy the source file to the destination and return the number of bytes that were copied.
func CopyFile(ctx context.Context, source string, destination string) (int64, error) {
	src, dest, srcInfo, err := openFilesForCopying(source, destination)
	if err != nil {
		return 0, fmt.Errorf("failed to copy the file %q to %q. %w", source, destination, err)
	}
	defer src.Close()
	defer dest.Close()

	wc, err := copyN(ctx, src, dest, srcInfo.Size())
	if err != nil {
		return wc, fmt.Errorf("failed to copy the file %q to %q. %w", source, destination, err)
	}

	return wc, nil
}

// Copy N bytes from the source file to the destination and return the number of bytes that were copied.
func CopyFileN(ctx context.Context, source string, destination string, count int64) (int64, error) {
	src, dest, _, err := openFilesForCopying(source, destination)
	if err != nil {
		return 0, fmt.Errorf("failed to copy the file %q to %q. %w", source, destination, err)
	}
	defer src.Close()
	defer dest.Close()

	wc, err := copyN(ctx, src, dest, count)
	if err != nil {
		return wc, fmt.Errorf("failed to copy the file %q to %q. %w", source, destination, err)
	}

	return wc, nil
}

func openFilesForCopying(source string, destination string) (*os.File, *os.File, fs.FileInfo, error) {
	src, err := os.Open(source)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open the source file %q. %w", source, err)
	}

	srcStat, err := src.Stat()
	if err != nil {
		src.Close()
		return nil, nil, nil, fmt.Errorf("failed to do Stat() on the source file %q. %w", source, err)
	}

	dest, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcStat.Mode().Perm())
	if err != nil {
		src.Close()
		return nil, nil, nil, fmt.Errorf("failed to create the destination file %q. %w", destination, err)
	}

	return src, dest, srcStat, nil
}

func copyN(ctx context.Context, src io.Reader, dest io.Writer, count int64) (int64, error) {
	in := contextio.NewReader(ctx, src)
	out := contextio.NewWriter(ctx, dest)

	wc, err := io.CopyN(out, in, count)
	return wc, err
}
