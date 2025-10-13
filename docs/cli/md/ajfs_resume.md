## ajfs resume

Resume calculating file signature hashes.

### Synopsis

Resume calculating file signature hashes for a previously interrupted scan.

NOTE: The database must have been created using the "--hash" option.

```
ajfs resume [flags]
```

### Examples

```
  # resume using the default ./db.ajfs database
  ajfs resume

  # resume the specific database and display a progress bar
  ajfs resume --progress /path/to/database.ajfs
```

### Options

```
  -h, --help       help for resume
  -p, --progress   Display progress information.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

