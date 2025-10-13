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
```

### Options

```
  -h, --help   help for diff
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

