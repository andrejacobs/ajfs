## ajfs tree

Display the file hiearchy tree.

### Synopsis

Display the file hiearchy as a tree in a similar way the popular tree command does.

```
ajfs tree [flags]
```

### Examples

```
  # display the entire hierarchy from the default ./db.ajfs
  ajfs tree

  # display the entire hierarchy of the specified database
  ajfs tree /path/to/database.ajfs

  # display a subtree from the default ./db.ajfs
  ajfs tree /sub/tree/path/inside

  # display only directories
  ajfs tree --dirs /path/to/database.ajfs

  # display only directories and limit the depth to 3 layers starting at the subtree
  ajfs tree --dirs --limit 3 /path/to/database.ajfs /sub/tree/path/inside
```

### Options

```
  -d, --dirs        Display only directories.
  -h, --help        help for tree
  -l, --limit int   Limit the tree depth.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

