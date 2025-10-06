# ajfs - Specification

Author: André Jacobs
Date: 3/10/2024

-   28/09/2025: Added more commands and details.

## Overview

Why mk2 (mark 2)?

-   About a year ago I wrote `ajfs` which I used internally quite a bit.
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
