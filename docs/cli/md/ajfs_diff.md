## ajfs diff

Display the differences between two databases and or file system hierarchies.

### Synopsis

Display the differences between two databases and or file system hierarchies.

Compares the path entries from the left hand side (LHS) against those of the
right hand side (RHS) and displays what those differences are.

You can compare:
* A database against its root path to see what has changed since the database
  was created.
* A database against another database.
* A database against another file system hierarchy.
* One file system hierarchy against another one.

Differences are displayed in the following format:

* If the file or directory only exists in the left hand side (as in removed 
  from the LHS):

  `d---- Path/of/dir` or `f---- Path/of/file`.

* If the file or directory only exists in the right hand side (as in added 
  to the RHS):

  `d++++ Path/of/dir` or `f++++ Path/of/file`.

 * The item exists in both the LHS and RHS but has a change then the following
   format is used:
 
   fmslx Path/of/file

   * f or d: to denote a file or directory.
   * m: type and or permissions has changed.
   * s: size has changed.
   * l: last modification date has changed.
   * x: file signature hash has changed.
   * ~: this property has not changed.

   For example a file that has changed in size and its last modification date:

   f~sl~ Path/of/file

Differences are displayed in the following order:

* Items that only exist in the left hand side.
* Items that only exist in the right hand side.
* Items that exist on both sides and have changed.

You can also filter on items to be included or excluded from the diff output.
The filter uses the same f, d, m, s, l and x notation.
The filter can also include - for LHS, + for RHS or ~ for something has changed.
Include filters are checked first and at least one need to be matched for the item to appear in the output.
Exclude filters are checked after any include filters and an item need to not match any exclude filter to be kept
in the output.

```
ajfs diff [flags]
```

### Examples

```
  # differences between the default ./db.ajfs database and the root path
  ajfs diff

  # differences between a specific database and its root path
  ajfs diff /path/to/database.ajfs

  # differences between two databases
  ajfs diff /path/to/lhs.ajfs /path/to/rhs.ajfs

  # differences between a database and the file system hierarchy
  ajfs diff /path/to/lhs.ajfs /path/to/rhs

  # differences between two file system hierarchies
  ajfs diff /path/to/lhs /path/to/rhs
 
  # only show differences where the size and hash has been changed
  ajfs diff --include=sx /path/to/lhs /path/to/rhs

  # only show differences where the last modification time has not been changed
  ajfs diff --exclude=l /path/to/lhs /path/to/rhs

  # ignore differences where a directory's size or a file's mode has changed (e.g. copying files from a Mac to a NAS)
  ajfs diff -e=ds -e=fm /path/to/lhs /path/to/rhs

  # only show differences for files on LHS or RHS and exclude if the size or last modification time has been changed
  ajfs diff -i=f- -i=f+ -e=s -e=l /path/to/lhs /path/to/rhs
```

### Options

```
  -e, --exclude stringArray   Exclude filter
  -h, --help                  help for diff
  -i, --include stringArray   Include filter
  -o, --only-stats            Display only statistics
  -s, --stats                 Display diffs and statistics
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

