## ajfs dupes

Display all duplicate files or directory trees.

### Synopsis

Display all duplicate files or directory subtrees that are the same.

The database must contain the calculated file signature hashes if you are using
this command to find duplicate files. The default mode.

Duplicate files will be displayed in the following example format:

```
>>>
Hash: c88e6e3b20f8478468288d2bef9cf624f5707ebcdad6113d4a545469333271a1
Size: 11167407 [11 MB]

[0]: Bought/Books/Some-awesome-book.pdf
[1]: Another/dir/somewhere/with/backups/Same-book.pdf

Count: 2
Total Size: 22334814 [22 MB]
<<<
```
To find all duplicate subtrees use the "-d, --dirs" option.
Each parent of a subtree in the hierarchy is given a unique signature that is 
calculated based on each of its children's signatures. Thus it can be used
to find subtrees in the hierarchy that share the same children regardless
of where in the hierarchy they are.

For example: We have 2 copies of the Day1 directory.

```
root
├── Photos
│   └── Holiday2025
│       └── Day1
│           ├── Photo1.jpg
│           └── Photo2.jpg		
└── Backup
    └── MyPhotos
        └── 2025
            └── Day1
                ├── Photo1.jpg
                └── Photo2.jpg

```
Using "--dirs" would produce the example format:

```
Signature: eb14cf7cad5f771d25bc5d2fa8ac012473a58044
  Photos/Holiday2025/Day1
  Backup/MyPhotos/2025/Day1

```
Using "--dirs --tree" would produce the example format:

```
Signature: eb14cf7cad5f771d25bc5d2fa8ac012473a58044
  Photos/Holiday2025/Day1
  Backup/MyPhotos/2025/Day1
  ├── Photo1.jpg     [15730819566f2bc79c3c6f151c5572b58b14a1c6]
  └── Photo2.jpg     [9aff76baba26e2e51f7e94b16efbf0505ddb71a9]
```


```
ajfs dupes [flags]
```

### Examples

```
  # display duplicate files from the default ./db.ajfs database
  ajfs dupes

  # display duplicate files from the specified database
  ajfs dupes /path/to/database.ajfs

  # display duplicate subtrees in the tree format
  ajfs dupes --dirs --tree /path/to/database.ajfs
```

### Options

```
  -d, --dirs   Display duplicate subtree directories.
  -h, --help   help for dupes
  -t, --tree   Display the tree hierarchy of duplicate subtrees.
```

### Options inherited from parent commands

```
  -v, --verbose   Display verbose information.
```

### SEE ALSO

* [ajfs](ajfs.md)	 - Andre Jacobs' file hierarchy snapshot tool.

