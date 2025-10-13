## ajfs tosync

Show which files need to be synced from the LHS to the RHS.

### Synopsis

Show which files need to be synced from the left hand side (LHS) to the
right hand side (RHS).

Think of this as a quick way to see which files on the LHS has been changed or
added and have not yet been copied onto the RHS.
 
NOTE: Does not do any syncing. This is the job for the excellent rsync.

 Criteria are:
* Only files that appear on the LHS will be shown.
* Files that have changed will be shown and thus indicate that the ones on
  the RHS need to be overwritten.
* Permissions and last modification times are ignored since these are bound
  to be different between two systems.
* If both databases have compatible file signature hashes, then items with
  a different hash will also be shown.

One of the biggest use cases for this command is to be able to see which files
on one system (e.g. laptop) has not yet been backed up somewhere on another
system (e.g. Linux server). In which case the file locations are different
between the systems. In order to do this you need to perform a scan with
file signature hash calculations on both systems and the use:
  ajfs tosync lhs.ajfs rhs.ajfs


```
ajfs tosync [flags]
```

### Examples

```
  # compares the default database ./db.ajfs as the LHS against the RHS database
  ajfs tosync /path/to/rhs.ajf

  # compares the LHS database against the RHS database
  ajfs tosync /path/to/lhs.ajfs /path/to/rhs.ajfs

  # only compare the file signature hashes. Useful when the files are in different locations
  ajfs tosync --hash lhs.ajfs rhs.ajfs

```

### Options

```
  -f, --full   Display full paths for entries.
  -s, --hash   Compare only the file signature hashes.
  -h, --help   help for tosync
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

