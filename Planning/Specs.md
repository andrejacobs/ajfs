# ajfs - Specification

Author: André Jacobs
Date: 3/10/2024

-   2025: Planning and documenting the required features as I develop them.

## Overview

Why mk2 (mark 2)?

-   Some time in 2023 I wrote `ajfs` which I used internally quite a bit.
-   I now consider that project as the mk1 that helped me prototype and play around with ideas.
-   I learned a lot about what I would like to improve in the next version, which now brings me to mk2.

What is this tool for?

-   Used to take snapshots of file hierarchies, meta data and hash signatures.
-   Can be used to track differences. Between different hierarchies, different periods in time.
-   Can be used to browse, search etc. offline from the host on which the files exists.
-   Can be used to see if files have been changed (damaged maybe? need to be backed up maybe? etc.)

Who is it for?

-   The audience is mainly going to be me to start with because it solves a problem I actually have. Maybe others will find this useful too. But there will have to be big disclaimers about using this at own risk. Need to fully establish liability is not with me.

Why build this yourself?

-   I have been needing and thinking about a tool like this for a number of years now but never quite found something that I would use consistently.
    Most of the time it would be cobbled toghether shell scripts and 3rd party tools.
    There was also the mk1 version but I am not quite happy about the UX.

## Abbreviations and terminology

-   LHS: Left Hand Side vs.
-   RHS: Right Hand Side

## Command summary

Overview of the subcommands that will be available.

### Core features

-   `scan`

    -   Used to walk a file hierarchy and store the found paths in a new ajfs database file.
    -   [Optional feature] Calculate the file signature hashes.
        -   `-s, --hash`: Start calculating hashes once the scan is finished. This can be interrupted and continued.
        -   `-a, --algo`: The hashing algorithm to use. Valid values are `sha1`, `sha256` and `sha512`.
    -   [Optional feature] Build the tree structure and cache it.
    -   `-p, --progress`: Show progress while calculating file hash signatures or building the tree.
    -   Store the information in a single file "database".
    -   Will not update an existing database. Use `update` command for this.
    -   Will not override an existing database.
        -   `--force`: Override an existing database.
    -   See the path filtering section on how to control which directories and files will be walked.
    -   `--dry-run`: Instead of creating the database, this will walk the hierarchy and display the files and directories that would have been stored.
    -   Can be safely interrupted (SIGINT Ctrl+C / SIGTERM) while the file signature hashes are being calculated
        given that this process could take hours on a big data set. Use `resume` to resume calculating hashes.
    -   Example: `ajfs scan /dir/to/scan` (creates ./db.ajfs) or `ajfs scan /path/to/db /path/to/scan`

-   `resume`

    -   Resume calculating the file signature hashes if a previous scan was safely interrupted.
    -   The ajfs database must have a hash table for resume to work.
    -   `-p, --progress`: Show progress while calculating file hash signatures.

-   `update`

    -   Perform a scan and update an existing database.
    -   It will only recalculate file signature hashes for things that have changed or been added.
    -   `-p, --progress`: Show progress while calculating file hash signatures or building the tree.
    -   See the path filtering section on how to control which directories and files will be walked.

-   `info`

    -   Display information about a database.
    -   path, version, root path, meta (os, arch), file size, creation date
    -   number of entries (files, directories)
    -   total size of all files, max file size, avg file size
    -   Will also check the file integrity against the stored checksum.
    -   Example: `ajfs info` (uses ./db.ajfs) or `ajfs info /path/to/db`

-   `list`

    -   List out all the path entries in the database.
    -   `-f, --full`: Display the full path including the root path of when the database was created.
    -   `-v, --verbose`: Shows a header as the first line of output.
    -   `-s, --hash`: Also output the file signature hash if available.
    -   `-m, --minimal`: Display only the path. Similar to what `find .` would produce.
    -   Example: `ajfs list` (uses ./db.ajfs) or `ajfs list /path/to/db`

-   `export`

    -   Export a database to different formats.
    -   `--format=csv` Export to a CSV format. The default option.
    -   `--format=json` Export to a JSON format.
    -   `--format=hashdeep` Export to a format compatible with hashdeep.
    -   Will not override any existing files.
    -   Example: `ajfs export /path/export.csv` (uses ./db.ajfs) or
        `ajfs export /path/database.ajfs /path/export.csv`

-   `diff`

    -   Used to describe the differences between two databases.
    -   Can also be used to describe what has changed between the database version and the current file system state.
    -   A diff between a database and the current file system state will ignore comparing hashes. Since hashing could be a long running task.
        The same functionality can be achieved by creating a new scan followed by a diff against two databases.
    -   If two databases are specified but have two different hashing algorithms, then hashes will not be compared.
    -   Example:
        -   `ajfs diff` compares ./db.ajfs and the root path that created the database.
        -   `ajfs diff /path/to/db.ajfs` compares the specific database and the root path that created it.
        -   `ajfs diff /path/to/lhs.ajfs /path/to/rhs.ajfs` compare two databases.
        -   `ajfs diff /path/to/lhs.ajfs /path/to/dirs` compare a database against a directory.
        -   `ajfs diff /path/a /path/b` compare two file systems.

-   `tosync`

    -   Shows what files need to be synced from the LHS to the RHS. NOTE: Does not do any syncing. This is the job for the execellent rsync.
        For example I can use this to see which files on my Mac has not yet been backed up on a Linux server (even though the paths will be different etc.).
        Meaning this is just a quick way of seeing if any files on the LHS has not yet been copied somewhere onto the RHS.
    -   Criteria are:
        -   Files that only appear in the LHS will be shown.
        -   Files that have changed will be shown and thus indicate that the ones on the RHS need to be overwritten.
        -   However permissions and last modification times are ignored since these are bound to be different between two systems.
        -   If both databases have compatible file signature hashes, then items with a different hash will also be shown.
    -   One of my biggest use cases for this tool is to be able to tell if some files on one system have actually been backed up onto another system.
        In these situations the paths are not the same.
        -   `-s, --hash`. When this option is specified then both sides (LHS and RHS) will need to have file signature hashes. Only the hashes will be compared and any hash that appears only in the LHS or that is different to the hash on the RHS will be shown.
    -   Examples:
        -   `ajfs tosync /path/to/rhs.ajfs` compares ./db.ajfs as the LHS against the RHS database.
        -   `ajfs tosync /path/to/lhs.ajfs /path/to/rhs.ajfs` compares the LHS database against the RHS database.
        -   `ajfs tosync --hash lhs.ajfs rhs.ajfs` Will only compare the file signature hashes and can tell which files have changed or are new on the LHS.
            For example this will ignore if the same file exists on both sides but in different locations.

-   `search`

    -   Search and display path entries that match certain criteria.
    -   `-v, --verbose`: Shows a header as the first line of output.
    -   `-f, --full`: Display the full path including the root path of when the database was created.
    -   `-m, --minimal`: Display only the id, path and optionally the file signature hash.
    -   You can specify multiple criteria and they would have to all be matched (e.g. AND). Example: `-e something -t f --size +10k`
    -   `-e, --exp` Match path against a regular expression
    -   `-i, --iexp` Case insensitive match path against a regular expression
    -   `-n, --name` Match base name against the shell pattern (e.g. \* ?). Base name is the last part of a path (e.g. /dir/to/file.txt would have file.txt as the base name)
    -   `--iname` Case insensitive match base name against the shell pattern (e.g. \* ?)
    -   `-p, --path` Match path against the shell pattern (e.g. \* ?)
    -   `--ipath` Case insensitive match path against the Shell pattern (e.g. \* ?)
    -   `-t, --type` Match if the type is one of the following:
        -   d directory
        -   f regular file
        -   l symbolic link
        -   p named pipe (FIFO)
        -   s socket
    -   `-s, --hash` Match if the file signature hash starts with this prefix
    -   `--size` Match the file size according to:
        -   <n> with no suffix means exactly <n> bytes. e.g. --size 100
        -   With one of the following scaling suffixes:
        -   `k/K` Kilobytes (1 KB = 1000 bytes). e.g. --size 1k
        -   `m/M` Megabytes (1 MB = 1000 KB). e.g. --size 1m
        -   `g/G` Gigabytes (1 GB = 1000 MB). e.g. --size 1g
        -   `t/T` Terrabytes (1 TB = 1000 GB). e.g. --size 1t
        -   `p/P` Petabytes (1 PB = 1000 TB). e.g. --size 1p
        -   With one of the following operation prefixes:
        -   `+` Greater than. e.g. --size +1k
        -   `-` Less than. e.g. --size -1k
    -   `-b, --before` Match if the entry's last modification time is before this time.
        -   The following formats are allowed:
        -   `YYYY-MM-DD`
        -   `YYYY-MM-DD HH:mm:ss`
        -   `YYYY-MM-DDTHH:mm:ss`
        -   `<n>D` n Days before now. e.g. -b 10D
        -   `<n>M` n Months before now. .e.g. -b 2M
        -   `<n>Y` n Years before now. e.g. -b 5Y
    -   `-a, --after` Match if the entry's last modification time is after this time.
        -   The following formats are allowed:
        -   `YYYY-MM-DD`
        -   `YYYY-MM-DD HH:mm:ss`
        -   `YYYY-MM-DDTHH:mm:ss`

-   `tree`

    -   Display the whole or partial file hierarchy tree in the same way the CLI `tree` command does.
    -   Examples:
        -   `ajfs tree` will use ./db.ajfs and display entire tree.
        -   `ajfs tree /path/to/db.ajfs` will use the specified database and display the entire tree.
        -   `ajfs tree /path/to/db.ajfs some/sub/path` will use the specified database and only display a sub tree.
    -   TODO: Think about options to support, i.e. limit, dirs only, full path
    -   TODO: Would be nice to also output colours.

-   `dupes`

    -   List out all duplicate files (ignoring 0 sized files) and is based on the file's signature hash.
        Thus the database must contain file signature hashes.
    -   `-d, --dirs` List out all duplicate directory trees.
        Each directory will have a hash based on the filenames and directories inside of it.
        Does not need file signature hashes.
        -   `-t, --tree` Also print out the sub tree so you can see what is being duplicated.

### Global flags

-   `-h, --help`

    -   Displays usage and help information.

-   `--version`

    -   Displays the version of the tool.

-   `-v, --verbose`

    -   Output verbose information

### Path filtering

Some commands can perform path filtering. Filtering either checks whether a file or directory should be included
or if it should be excluded. An include filter will always be performed first and thus skip any exclude filters.

You can include multiple filters on the CLI (e.g. `-i someting -i another -e notThis`)

-   `-i, --include {pattern}`
-   `-e, --exclude {pattern}`

Pattern is a regular expression that can be optionally prefixed with `f:` for file or `d:` for directory.
For example to include all files that match the extension .pdf and exclude any directories that end with temp, you
could use this on the CLI `-i "f:\.pdf$" -e "d:temp$"`.
If the prefix (f: or d:) is not specified then the regular expression will be applied to both files and directories.

See https://pkg.go.dev/regexp/syntax for the syntax.

## Example usage

-   Version: `ajfs --version`
-   Help: `ajfs --help`
-   Help on a specific command: `ajfs scan --help`
-   List out the database entries: `ajfs list --full`

## Miscelaneous

Checksum algorithm chosen: CRC32
Why?

-   I just need a simple check to see if the file has been corrupted in some way.
-   The go stdlib comes with an implementation. So trusted to be maintained and well written implementation and platform support.
-   No extra 3rd party dependency.

Alternatives looked at:

-   https://xxhash.com/
-   https://github.com/cespare/xxhash only implements xxh64. Used by 70k+ and notable projects.
-   CRC32 gives really good performance and is hard to beat by some implementations. Has hardware support etc.
