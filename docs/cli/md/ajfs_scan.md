## ajfs scan

Create a new database.

### Synopsis

Create a new database by walking an existing file system hierarchy.

The directory to be walked is also known as the root path and this path is
stored in the database. Use "ajfs info" to see the root path.
The root path can be used to display the full path for database entries while
using the "-f, --full" flag on some of the other commands.

Additionally the file signature hashes of files can be calculated and stored in
the database. This can be very valuable for later finding duplicates or
differences. Calculating the file signature hashes can be a long running
process depending on the number of files and sizes.

The file signature hash calculation process can be safely interrupted using
Ctrl+C (SIGTERM) and be resumed at another time using "ajfs resume".
However it is not safe to interrupt the initial database creation, 
use "--verbose" or "--progress" to know when the calculation process has
started.

Path filtering:

Used to check whether a file or directory should be included or if it should
be excluded. An include filter will always be performed first and thus skip
any exclude filters.

You can include multiple filters on the CLI (e.g. "-i someting -i another")

  "-i, --include {pattern}"
  "-e, --exclude {pattern}"

Pattern is a regular expression that can be optionally prefixed with "f:" for
file or "d:" for directory.
For example to include all files that match the extension .pdf and exclude
any directories that end with temp, you could use this on the CLI
-i "f:\.pdf$" -e "d:temp$".
If the prefix (f: or d:) is not specified then the regular expression will be
applied to both files and directories.

See https://pkg.go.dev/regexp/syntax for the syntax.

```
ajfs scan [flags]
```

### Examples

```
  # create the default ./db.ajfs database from the specified path
  ajfs scan /path/to/be/scanned

  # create a new database from the specified path
  ajfs scan /path/to/database.ajfs /path/to/be/scanned

  # see which paths will be included without creating the database
  ajfs scan --dry-run -i "f:\.pdf$" /path/to/be/scanned

  # override the existing database if it exists
  ajfs scan --force /path/to/database.ajfs /path/to/be/scanned

  # create a new database and calculate the file signature hashes using SHA-256
  ajfs scan --hash /path/to/database.ajfs /path/to/be/scanned

  # create a new database and calculate the file signature hashes using SHA-1 while showing a progress bar
  ajfs scan --hash --algo=sha1 --progress /path/to/database.ajfs /path/to/be/scanned

  # create a new database and only include PDF and EPUB files
  ajfs scan -i "f:\.pdf$" -i "f:\.epub$" /path/to/be/scanned

  # create a new database and exclude all directories that contain the word "temp"
  ajfs scan -e "d:temp" /path/to/be/scanned
```

### Options

```
  -a, --algo string           Hashing algorithm to use. Valid values are 'sha1', 'sha256' and 'sha512'. (default "sha256")
      --dry-run               Only display files and directories that would be stored in the database.
  -e, --exclude stringArray   Exclude path regex filter
      --force                 Override any existing database.
  -s, --hash                  Calculate file signature hashes.
  -h, --help                  help for scan
  -i, --include stringArray   Include path regex filter
  -p, --progress              Display progress information.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

