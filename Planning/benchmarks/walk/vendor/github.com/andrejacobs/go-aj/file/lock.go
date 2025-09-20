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
	"errors"
	"io"
	"os"
	"strconv"
)

// Lockfile is used to acquire a lock on a process for various tasks to be
// performed and thus other processes should not be allowed to run while
// the lock has been acquired.
// NOTE: This is not a lock on a single file, this is a lock on a service
// or process to ensure no other process (that adheres to the lock) will
// run.
// The lock file that is created contains the PID of the process that
// acquired the lock.
type Lockfile struct {
	path string // The path to the lock file
	pid  int    // The PID of the process that has the lock
}

var (
	ErrLockfileAcquired = errors.New("failed to acquire the lock file")
	ErrLockfileNotOwned = errors.New("the current process does not own the lock file")
)

// Attempt to acquire the lock file specified by the path.
// If the lock file does not exist, then it will be created and the current PID
// will be written to the file.
// If the lock file does exist, then the Lockfile info along with the
// error ErrLockfileAcquired will be returned.
func AcquireLockfile(path string) (*Lockfile, error) {
	// Try to create the file
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		// File already exists or something else went wrong
		// Try and get the PID of who owns the file
		pid, pidErr := lockFileGetPid(path)

		lock := &Lockfile{
			path: path,
			pid:  pid,
		}

		return lock, errors.Join(ErrLockfileAcquired, err, pidErr)
	}

	lock := &Lockfile{
		path: path,
		pid:  os.Getpid(),
	}

	_, err = f.WriteString(strconv.Itoa(lock.pid))
	if err != nil {
		// Release the lock
		f.Close()
		os.Remove(path)
		return nil, err
	}

	err = f.Close()
	return lock, err
}

// Attempt to acquire the lock file specified by the path.
// This is the same as AcquireLockfile but allows for the same process to acquire
// the same lock file multiple times (re-entrant)
// If the lock file does exist and is not owned by the same process then the
// Lockfile info along with the error ErrLockfileAcquired will be returned.
// NOTE: You can not call Release multiple times (e.g. 1 Acquire = 1 Release)
// as the first call to Release will release the lock.
func AcquireLockfileReEntrant(path string) (*Lockfile, error) {
	lock, err := AcquireLockfile(path)
	if err != nil {
		if errors.Is(err, ErrLockfileAcquired) && lock.pid == os.Getpid() {
			return lock, nil
		}
	}
	return lock, err
}

// Release the lock so that another process can acquire the lock.
// The lock file can only be released if it was acquired by the same process.
// The error ErrLockfileNotOwned will be returned if the lock file is not owned
// by the current process.
func (l *Lockfile) Release() error {
	if l.pid != os.Getpid() {
		return ErrLockfileNotOwned
	}

	err := os.Remove(l.path)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

// Path of the lock file.
func (l *Lockfile) Path() string {
	return l.path
}

// PID of the process that owns the lock file.
func (l *Lockfile) Pid() int {
	return l.pid
}

//-----------------------------------------------------------------------------

// Open a lock file and read the PID.
func lockFileGetPid(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return readLockfilePid(f)
}

func readLockfilePid(r io.Reader) (int, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(data))
}
