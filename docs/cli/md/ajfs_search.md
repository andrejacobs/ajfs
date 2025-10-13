## ajfs search

Search for matching path entries.

### Synopsis

Search for entries in the database that match certain criteria.

Criteria include:
* Matching a path against a regular expression.
* Matching the path or the base name (last component e.g. filename) against
  a shell pattern (e.g. * ?).
* Matching the type of entry (e.g. a directory, file etc.).
* Matching the path identifier against a prefix.
* Matching the file signature hash against a prefix.
* Matching if the size is exactly, greater or less than a value.
* Matching if the last modification date is before or after a value.


```
ajfs search [flags]
```

### Examples

```
  # search for all .txt files in the default ./db.ajfs database
  ajfs search -i "\.txt$"

  # search for all .txt using shell pattern
  ajfs search --iname "*.txt"

  # display all files bigger than 1MB
  ajfs search --type f --size +1M

  # display all files smaller than 1GB
  ajfs search --type f --size -1G

  # display all entries with a last modification date before the date
  ajfs search --before 2019-03-01

  # display all entries with a last modification date before 30 days ago
  ajfs search --before 30D

  # display all entries with a last modification date after the date
  ajfs search --after 1999-03-24

```

### Options

```
  -a, --after string        Match if the entry's last modification time is after this time.
                              The following formats are allowed:
                              YYYY-MM-DD
                              YYYY-MM-DD HH:mm:ss   Also supports YYYY-MM-DDTHH:mm:ss
                            
  -b, --before string       Match if the entry's last modification time is before this time.
                              The following formats are allowed:
                              YYYY-MM-DD
                              YYYY-MM-DD HH:mm:ss   Also supports YYYY-MM-DDTHH:mm:ss
                              <n>D  n Days before now
                              <n>M  n Months before now
                              <n>Y  n Years before now
                            
  -e, --exp stringArray     Match path against the regular expression.
  -f, --full                Display full paths for entries.
  -s, --hash string         Match if the file signature hash starts with this prefix.
  -h, --help                help for search
      --id string           Match if the entry's identifier starts with this prefix.
  -i, --iexp stringArray    Case insensitive match path against the regular expression.
      --iname stringArray   Case insensitive match base name against the shell pattern (e.g. * ?).
      --ipath stringArray   Case insensitive match path against the shell pattern (e.g. * ?).
  -m, --more                Display more information about the matching paths.
  -n, --name stringArray    Match base name against the shell pattern (e.g. * ?).
  -p, --path stringArray    Match path against the shell pattern (e.g. * ?).
      --size stringArray    Match the file size according to:
                              <n> with no suffix means exactly <n> bytes. e.g. --size 100
                            
                              With one of the following scaling suffixes:
                              k/K   Kilobytes (1 KB = 1000 bytes). e.g. --size 1k
                              m/M   Megabytes (1 MB = 1000 KB). e.g. --size 1m
                              g/G   Gigabytes (1 GB = 1000 MB). e.g. --size 1g
                              t/T   Terrabytes (1 TB = 1000 GB). e.g. --size 1t
                              p/P   Petabytes (1 PB = 1000 TB). e.g. --size 1p
                            
                              With one of the following operation prefixes:
                              +   Greater than. e.g. --size +1k
                              -   Less than. e.g. --size -1k
  -t, --type string         Match if the type is one of the following:
                              d  directory
                              f  regular file
                              l  symbolic link
                              p  named pipe (FIFO)
                              s  socket
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

