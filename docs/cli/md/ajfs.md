## ajfs

Andre Jacobs' file hierarchy snapshot tool.

### Synopsis

Andre Jacobs' file hierarchy snapshot tool is used to save a file system
hierarchy to a single flat file database.

Which can then be used in an offline and independant way to do the following:
* Find duplicate files or entire duplicate subtrees.
* Compare differences between databases (snapshots) and or file systems.
* Find out which files would still need to be synced to another system.
* Search for entries that match certain criteria.
* List or export the entries to CSV, JSON or Hashdeep.
* Display the entries as a tree.


### Options

```
  -h, --help      help for ajfs
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs diff](ajfs_diff.md)	 - Display the differences between two databases and or file system hierarchies.
* [ajfs dupes](ajfs_dupes.md)	 - Display all duplicate files or directory trees.
* [ajfs export](ajfs_export.md)	 - Export a database.
* [ajfs info](ajfs_info.md)	 - Display information about a database.
* [ajfs list](ajfs_list.md)	 - Display the database path entries.
* [ajfs resume](ajfs_resume.md)	 - Resume calculating file signature hashes.
* [ajfs scan](ajfs_scan.md)	 - Create a new database.
* [ajfs search](ajfs_search.md)	 - Search for matching path entries.
* [ajfs tosync](ajfs_tosync.md)	 - Show which files need to be synced from the LHS to the RHS.
* [ajfs tree](ajfs_tree.md)	 - Display the file hiearchy tree.
* [ajfs update](ajfs_update.md)	 - Perform a new scan and update an existing database.

