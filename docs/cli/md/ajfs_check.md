## ajfs check

Check the integrity of a database.

### Synopsis

Check the integrity of a database.
This is just a convenient way for running: ajfs fix --dry-run


```
ajfs check [flags]
```

### Examples

```
  # using the default ./db.ajfs database
  ajfs check

  # using a specific database
  ajfs check /path/to/database.ajfs
```

### Options

```
  -h, --help   help for check
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

