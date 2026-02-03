# ajfs

[![](https://github.com/andrejacobs/ajfs/actions/workflows/makefile.yml/badge.svg)](https://github.com/andrejacobs/ajfs/actions/workflows/makefile.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/andrejacobs/ajfs)](https://goreportcard.com/report/github.com/andrejacobs/ajfs)

Index a file hierarchy with meta data and file signatures so that it can be searched, compared, duplicates found all while being on a different machine.

## Overview

Use `ajfs` to take snapshots of a file system hierarchy so that you can later perform operations such as searching
or finding duplicates even while being on a completely different machine. Please note that a snapshot does not
include the actual file content data.

Please see the [documentation](docs/cli/md/ajfs.md) for more details.

## Install

- Using [Homebrew](https://brew.sh/):

    ```shell
    brew tap andrejacobs/ajfs
    brew install --cask ajfs
    ```

## Manual Installation

You can download the pre-compiled binary for your system directly from the [Releases Page](https://github.com/andrejacobs/ajfs/releases).

### Linux / macOS

1. Determine your architecture:

    ```shell
    uname -m
    ```

    - `x86_64` -> Download `ajfs_Linux_x86_64.tar.gz` (Intel/AMD PCs).
    - `aarch64` -> Download `ajfs_Linux_arm64.tar.gz` (ARM64, Raspberry Pi).
    - `i386` or `i686` -> Download `ajfs_Linux_i386.tar.gz` (Older 32-bit systems).

2. Download and Install:

    Replace `VER` with the latest version (e.g., `1.0.0`) and `ARCH` with your architecture (e.g., `x86_64`).

    ```shell
    # 1. Download the archive
    wget https://github.com/andrejacobs/ajfs/releases/download/vVER/ajfs_Linux_ARCH.tar.gz

    # 2. Extract the executable
    tar -xzf ajfs_Linux_ARCH.tar.gz

    # 3. Move the executable to a system path
    sudo mv ajfs /usr/local/bin/

    # 4. Verify installation
    ajfs --version
    ```

### Windows

1. Download the latest `ajfs_Windows_x86_64.zip`:
2. Extract:
    - Right-click the downloaded `.zip` file.
    - Select **Extract All...** and choose a location (e.g. `C:\ajfs`).

3. Run:

    Open PowerShell or Command Prompt and navigate to that folder and run ajfs:

    ```powershell
    cd C:\ajfs
    .\ajfs.exe --version
    ```

    > **Tip:** To run `ajfs` from anywhere, add the folder location (e.g., `C:\ajfs`) to your system's **PATH** environment variable.

## Development

- Clone this repo.

    ```shell
    git clone git@github.com:andrejacobs/ajfs.git
    ```

- Build.

    ```shell
    make build

    # build and run
    make build && ./build/bin/ajfs --help
    ```

- Run unit-tests.

    ```shell
    make test
    ```

- Confirm code quality.

    ```shell
    make check
    ```

## Usage

Please see the [examples](docs/examples.md) for more detailed example use cases.

The "root path" is the path to the file system hierarchy from which a snapshot is created from.

The following is just a couple of examples of what is possible with `ajfs`.

- Create a new snapshot.

    ```shell
    # create the default ./db.ajfs database and scan the specified path
    ajfs scan /media/backups/e-books

    # specify where the snapshot database should be stored
    ajfs scan ~/mybooks.ajfs /media/backups/e-books

    # calculate file signature hashes and show progress updates
    ajfs scan --hash --algo=sha1 --progress ~/database.ajfs /media/backups
    ```

- Resume calculating file signature hashes.

    ```shell
    ajfs resume --progress ~/database.ajfs
    ```

- Update the snapshot to reflect the current file system hierarchy.

    ```shell
    ajfs update --progress ~/database.ajfs
    ```

- List a snapshot.

    ```shell
    ajfs list mydata.ajfs

    ajfs tree mydata.ajfs
    ```

- Search for matching entries.

    ```shell
    # list all PDF files that have `go` in the filename
    ajfs search ~/mybooks.ajfs --type f --iname '*go*.pdf'

    # list all entries that match the regular expression
    ajfs search -i "\.txt$"

    # list all files smaller than 1GB
    ajfs search --type f --size -1G
    ```

- See what has changed.

    ```shell
    # diff the default ./db.ajfs and the root path it was created from
    ajfs diff snap1.ajfs

    # diff two snapshots
    ajfs diff snap1.ajfs snap2.ajfs
    ```

- Find duplicates.

    ```shell
    # find duplicate files
    ajfs dupes database.ajfs

    # find duplicate directory subtrees
    ajfs dupes --dirs database.ajfs
    ```

- See what still needs to be backed up.

    ```shell
    # what needs to be copied from my laptop to my nas
    ajfs tosync ~/laptop.ajfs ~/nas.ajfs

    # which files exist on the nas that has been deleted on my laptop
    ajfs tosync ~/nas.ajfs ~/laptop.ajfs

    # which files from my laptop has not yet been backed up on the nas regardless of filename or location
    ajfs tosync --hash ~/laptop.ajfs ~/nas.ajfs
    ```

- Export the snapshot to other formats.

    ```
    ajfs export database.ajfs export.csv

    ajfs export --format=json database.ajfs export.json
    ```

## Disclaimer

This tool is provided "as is" and is intended for use at your own risk. The author makes no warranties as to its
performance, merchantability, or fitness for a particular purpose. Under no circumstances shall the author be liable
for any direct, indirect, special, incidental, or consequential damages (including, but not limited to, loss of data
or profit) arising out of the use or inability to use this software, even if advised of the possibility of such damage.

## License

`ajfs` is released under the MIT license. See [LICENSE](LICENSE).
