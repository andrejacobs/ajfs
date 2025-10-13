## ajfs export

Export a database.

### Synopsis

Export a database into one of the following formats: CSV, JSON or Hashdeep

```
ajfs export [flags]
```

### Examples

```
  # export the default ./db.ajfs to a CSV file
  ajfs export /path/to/export.csv

  # export a database to a CSV file
  ajfs export /path/to/database.ajfs /path/to/export.csv

  # export with full path information to a JSON file
  ajfs export --full --format=json /path/to/database.ajfs /path/to/export.json

  # export to a hashdeep file. NOTE: the database must contain file signature hashes
  ajfs export --format=hashdeep /path/to/export.sha256
```

### Options

```
      --format string   Export format: csv, json or hashdeep. (default "csv")
  -f, --full            Export full paths for entries.
  -h, --help            help for export
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

