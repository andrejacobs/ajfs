## ajfs list

Display the database path entries.

### Synopsis

Display all the path entries stored inside a database.

```
ajfs list [flags]
```

### Examples

```
  # using the default ./db.ajfs database
  ajfs list

  # using a specific database
  ajfs list /path/to/database.ajfs

  # display full paths, file signature hashes and more information for each entry
  ajfs list --full --hash --more /path/to/database.ajfs
```

### Options

```
  -f, --full   Display full paths for entries.
  -s, --hash   Display file signature hashes if available.
  -h, --help   help for list
  -m, --more   Display more information about the paths.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

