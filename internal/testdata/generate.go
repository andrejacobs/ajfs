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

package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// This program is used to generate the test files
// The API is simply, if it errors then it log.Fatal
// This was written to only work on *nix/darwin
func main() {
	log.Println("Generating test data")

	flag.Parse()

	rootDir := ""
	if len(flag.Args()) > 0 {
		rootDir = flag.Arg(0)
	}

	var err error
	rootDir, err = filepath.Abs(rootDir)
	if err != nil {
		log.Fatal(err)
	}

	generateDiffFiles(rootDir)
	generateNeedSyncFiles(rootDir)
}

func generateDiffFiles(rootDir string) {
	baseDir := filepath.Join(rootDir, "diff")
	log.Println("generating 'diff' files: " + baseDir)
	os.RemoveAll(baseDir)
	makeDir(baseDir)

	// a -> b
	// Expected output:
	// d----- quick
	// f----- quick/1.txt
	// f----- quick/2.txt
	// d----- dir1
	// d----- dir1/lhs-only

	// d+++++ fox
	// f+++++ fox/3.txt
	// d+++++ hole
	// f+++++ hole/4.txt
	// d+++++ dir2
	// d+++++ dir2/rhs-only

	// d~~sm~ .				<-- valid
	// d~~~m~ both
	// f~~s~~ both/6.txt
	// f~p~~~ both/7.txt
	// f~~~m~ both/8.txt
	// d~~~m~ dir3
	// d~~~m~ dir3/both

	// LHS only
	makeFile(filepath.Join(baseDir, "a/quick/1.txt"), "The quick brown fox", 0644)
	makeFile(filepath.Join(baseDir, "a/quick/2.txt"), "Jumped over the lazy dog", 0644)
	makeFile(filepath.Join(baseDir, "a/dir1/lhs-only"), "lhs-only", 0644)

	// RHS only
	makeFile(filepath.Join(baseDir, "b/fox/3.txt"), "Alpha Bravo 17", 0644)
	makeFile(filepath.Join(baseDir, "b/hole/4.txt"), "Only exists on the RHS", 0644)
	makeFile(filepath.Join(baseDir, "b/dir2/rhs-only"), "rhs-only", 0644)

	// Same on both sides
	makeFile(filepath.Join(baseDir, "a/both/5.txt"), "LHS and RHS equal", 0644)
	copy(filepath.Join(baseDir, "a/both/5.txt"), filepath.Join(baseDir, "b/both/5.txt"))
	setLastMod(filepath.Join(baseDir, "a/both/5.txt"), "202310310530.42")
	setLastMod(filepath.Join(baseDir, "b/both/5.txt"), "202310310530.42")
	makeDir(filepath.Join(baseDir, "a/dir3/both"))
	copy(filepath.Join(baseDir, "a/dir3/both"), filepath.Join(baseDir, "b/dir3/both"))

	// Changed
	// size
	makeFile(filepath.Join(baseDir, "a/both/6.txt"), "LHS version", 0644)
	makeFile(filepath.Join(baseDir, "b/both/6.txt"), "RHS version is bigger", 0644)
	setLastMod(filepath.Join(baseDir, "a/both/6.txt"), "202310310530.42")
	setLastMod(filepath.Join(baseDir, "b/both/6.txt"), "202310310530.42")

	// perms
	makeFile(filepath.Join(baseDir, "a/both/7.txt"), "Different permissions", 0644)
	copy(filepath.Join(baseDir, "a/both/7.txt"), filepath.Join(baseDir, "b/both/7.txt"))
	chmodX(filepath.Join(baseDir, "b/both/7.txt"))
	setLastMod(filepath.Join(baseDir, "a/both/7.txt"), "202310310530.42")
	setLastMod(filepath.Join(baseDir, "b/both/7.txt"), "202310310530.42")

	// last mod
	makeFile(filepath.Join(baseDir, "a/both/8.txt"), "Different last modification times", 0644)
	copy(filepath.Join(baseDir, "a/both/8.txt"), filepath.Join(baseDir, "b/both/8.txt"))
	setLastMod(filepath.Join(baseDir, "a/both/8.txt"), "202310280730.02")
	setLastMod(filepath.Join(baseDir, "b/both/8.txt"), "202310310530.42")

	// c -> d [only the hashed data should be different]
	// d~~~m~ .
	// f~~~~x changed.txt
	makeFile(filepath.Join(baseDir, "c/changed.txt"), "Jumped over the lazy dog", 0644)
	makeFile(filepath.Join(baseDir, "d/changed.txt"), "jumped over the lazy dog", 0644) // only first character is different
	setLastMod(filepath.Join(baseDir, "c/changed.txt"), "202310280730.02")
	setLastMod(filepath.Join(baseDir, "d/changed.txt"), "202310280730.02")
}

func generateNeedSyncFiles(rootDir string) {
	// a -> b: Used to check what needs copying from LHS to RHS. Same paths
	// a -> c: Not using same paths, thus need to use hashes for comparison

	// Expected output for "need to sync" a -> b
	// blank.txt
	// cached/2.txt

	// Expected output for "need to sync" a -> c
	// blank.txt

	baseDir := filepath.Join(rootDir, "need-sync")
	log.Println("generating 'need to sync' files: " + baseDir)
	os.RemoveAll(baseDir)
	makeDir(baseDir)

	// a -> b
	makeFile(filepath.Join(baseDir, "a/cached/1.txt"), "The quick brown fox", 0644)
	makeFile(filepath.Join(baseDir, "a/cached/2.txt"), "Jumped over the lazy dog", 0644)
	makeFile(filepath.Join(baseDir, "a/cached/3.txt"), "Alpha Bravo 17", 0644)
	makeFile(filepath.Join(baseDir, "a/cached/dupe.txt"), "backed up multiple times", 0644)
	makeFile(filepath.Join(baseDir, "a/cached/4.txt"), "The quick brown fox", 0644) // a dupe of 1.txt
	setLastMod(filepath.Join(baseDir, "a/cached/1.txt"), "202310280730.02")

	copy(filepath.Join(baseDir, "a"), filepath.Join(baseDir, "b"))
	makeFile(filepath.Join(baseDir, "a/blank.txt"), "", 0644) // only exists on the LHS
	makeDir(filepath.Join(baseDir, "a/dir1/dir1-1"))

	makeFile(filepath.Join(baseDir, "b/cached/5.txt"), "Only exists on the RHS", 0644)
	makeFile(filepath.Join(baseDir, "b/cached/2.txt"), "jumped over the lazy cow. 42", 0644) // Updated on the RHS
	chmodX(filepath.Join(baseDir, "b/cached/3.txt"))                                         // Permission changed on RHS
	setLastMod(filepath.Join(baseDir, "b/cached/1.txt"), "202310310530.42")                  // Last mod changed on RHS

	// c
	makeFile(filepath.Join(baseDir, "c/dupe.txt"), "backed up multiple times", 0644)
	copy(filepath.Join(baseDir, "c/dupe.txt"), filepath.Join(baseDir, "c/backup/dupe.txt"))
	copy(filepath.Join(baseDir, "a/cached/1.txt"), filepath.Join(baseDir, "c/backup/1.txt"))
	copy(filepath.Join(baseDir, "a/cached/2.txt"), filepath.Join(baseDir, "c/backup/2-another-name.txt"))
	copy(filepath.Join(baseDir, "a/cached/3.txt"), filepath.Join(baseDir, "c/cached/3.txt"))
	chmodX(filepath.Join(baseDir, "c/cached/3.txt"))                        // Permission changed on RHS
	setLastMod(filepath.Join(baseDir, "c/backup/1.txt"), "202310310530.42") // Last mod changed on RHS
	makeFile(filepath.Join(baseDir, "c/abc.txt"), "only on RHS", 0644)
}

func makeDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Fatal(err)
	}
}

func makeFile(path string, content string, perm os.FileMode) {
	makeDir(filepath.Dir(path))

	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		log.Fatalf("failed to create the file %q. %v", path, err)
	}
}

func copy(source string, dest string) {
	//AJ### TODO: This can now be replaced with fileutils.CopyFile (which is platform agnostic)
	makeDir(filepath.Dir(dest))

	cmd := exec.Command("cp", "-r", source, dest)
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to copy %q to %q. %v", source, dest, err)
	}
}

func chmodX(path string) {
	// Only setting the executable permission since this is the only one Git tracks
	cmd := exec.Command("chmod", "+x", path)
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to: chmod +x %s. %v", path, err)
	}
}

func setLastMod(path string, date string) {
	cmd := exec.Command("touch", "-mt", date, path)
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to: touch -mt %s %s. %v", date, path, err)
	}
}
