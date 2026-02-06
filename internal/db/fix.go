// Copyright (c) 2026 Andre Jacobs
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

package db

import (
	"fmt"
	"io"

	"github.com/andrejacobs/go-aj/ajio/trackedoffset"
	"github.com/andrejacobs/go-aj/ajmath/safe"
)

// Attempts to repair a damaged database.
// out is used to display information to the user (normally routed to STDOUT). Things to be fixed will be prefixed with >>.
// path is the file path to an existing database file.
// dryRun when set to true will only output issues to the output writer and not make any changes.
func FixDatabase(out io.Writer, path string, dryRun bool) error {
	// > OpenDatabase -----------------------------------------------

	dbf := &DatabaseFile{
		path: path,
	}

	var err error
	dbf.file, err = trackedoffset.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open the ajfs database file. path: %q. %w", path, err)
	}

	//TODO: think about the close

	// > readHeadersAndVerify ---------------------------------------

	// Check the signature and version
	if err := dbf.prefixHeader.read(dbf.file); err != nil {
		return fmt.Errorf("error reading the ajfs prefix header. path: %q. %w", dbf.path, err)
	}
	if dbf.prefixHeader.Signature != signature {
		return fmt.Errorf("not a valid ajfs file (invalid signature %q, expected %q). path: %q", dbf.prefixHeader.Signature, signature, dbf.path)
	}
	if dbf.prefixHeader.Version > currentVersion {
		return fmt.Errorf("not a supported ajfs file (invalid version %d, expected <= %d). path: %q", dbf.prefixHeader.Version, currentVersion, dbf.path)
	}

	fmt.Fprintf(out, "Signature: %s\n", string(dbf.prefixHeader.Signature[:]))
	fmt.Fprintf(out, "Version: %d\n", dbf.prefixHeader.Version)

	// Read the header
	if err := dbf.header.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs header. path: %q. %w", dbf.path, err)
	}

	// Read the root info
	if err := dbf.root.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs root entry. path: %q. %w", dbf.path, err)
	}

	fmt.Fprintf(out, "Root: %q\n", dbf.root.path)

	// Read the meta info
	if err := dbf.meta.read(dbf.file); err != nil {
		return fmt.Errorf("failed to read the ajfs meta entry. path: %q. %w", dbf.path, err)
	}

	fmt.Fprintf(out, "Meta | OS: %q\n", dbf.meta.OS)
	fmt.Fprintf(out, "Meta | Arch: %q\n", dbf.meta.Arch)
	fmt.Fprintf(out, "Meta | Created at: %q\n", dbf.Meta().CreatedAt)

	// Read entries
	entriesOffset, err := safe.Uint64ToUint32(dbf.file.Offset())
	if err != nil {
		return err
	}

	if dbf.header.EntriesOffset != entriesOffset {
		fmt.Fprintf(out, ">> Entries offset is expected to be %x, actual is %x\n", entriesOffset, dbf.header.EntriesOffset)
	}

	//TODO: spin and read, check for EOF

	return nil
}
