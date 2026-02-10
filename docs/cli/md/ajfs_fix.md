## ajfs fix

Attempts to repair a damaged database.

### Synopsis

Attempts to repair a damaged database.

Use '--dry-run' to check the integrity of a database without making changes to
the database. This is equavalent to running 'ajfs check'.

A backup of the database header will be made before applying any changes.
The backup will be created in the current working directory using the same
filename as the database with the extension '.bak' added.

Use '--restore /path/to/___.bak' to restore a backup header to a database. 

>> Is used to display database errors that were found and that can be corrected.
!! Is used when an error happened during the process.



```
ajfs fix [flags]
```

### Examples

```
  # using the default ./db.ajfs database
  ajfs fix

  # using a specific database
  ajfs fix /path/to/database.ajfs

  # restore a backup header file
  ajfs fix --restore /path/to/header.ajfs.bak /path/to/database.ajfs
```

### Options

```
      --dry-run          Only display the repairs that will need to be performed.
  -h, --help             help for fix
      --restore string   Path to a backup header to be restored.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

