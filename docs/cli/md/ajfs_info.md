## ajfs info

Display information about a database.

### Synopsis

Display information about a database such as the path it was created from, meta, features and statistics.
Info will also validate the integrity of the database.

```
ajfs info [flags]
```

### Examples

```
  # using the default ./db.ajfs database
  ajfs info

  # using a specific database
  ajfs info /path/to/database.ajfs
```

### Options

```
  -h, --help   help for info
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

