## ajfs update

Perform a new scan and update an existing database.

### Synopsis

Perform a new scan on the root path and update an existing database.

The file system hierarchy specified by the root path stored in the database
will be scanned again and effectively a new updated database will be created.

However the strength in this command comes into play if the database also
contains the calculated file signature hashes. In this case the previously
calculated signatures will be copied as is for existing entries and only new
file entries will be calculated.

A backup of the existing database will first be created (with .bak suffix)
and if any error occurred then the database will be restored.


```
ajfs update [flags]
```

### Examples

```
  # update the existing default ./db.ajfs database
  ajfs update

  # update the specific database and show a progress bar
  ajfs update --progress /path/to/database.ajfs
```

### Options

```
  -e, --exclude stringArray   Exclude path regex filter
  -h, --help                  help for update
  -i, --include stringArray   Include path regex filter
  -p, --progress              Display progress information.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

