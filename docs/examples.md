# Examples of using ajfs

The core concept to remember is that you use ajfs to take a snapshot of a file system hierarchy (directory layout,
filenames, size, etc.) that you can then use to later perform things such as searching or finding duplicates.
You can perform these operations on a completely different machine than the one on which the file system hierarchy lives.
Please note that a snapshot does not include the actual file content data.

When creating a new snapshot you have the choice of including the file signature hashes or not. The file signature hash
is the value obtained by reading every byte from the file and calculating a hash using an algorithm such as SHA-256.
Certain features of ajfs require the file signature hashes to be present in the snapshot.

The following examples use:

-   `$` to indicate your local machine and `remote$` to indicate another machine.
-   the words `I`, `my` etc. written as if it was you that performed the operations.

## Installing ajfs

TODO:

## Offline search

I have a big collection of E-books that I have bought from a couple of sources over many years and store these on
a NAS running Linux. Before I buy any new books I need a way to quickly check if I have not already bought the book.
This needs to be done from my main computer.

-   First I need to take a snapshot of the existing books.
    -   See [ajfs scan](https://github.com/andrejacobs/ajfs/blob/main/docs/cli/md/ajfs_scan.md) for more details.

```shell
# On my Linux NAS
#  megaladon is a RAID 1 consisting of 12 TB drives

# Create a new database
nas$ ajfs scan ~/mybooks.ajfs /media/megaladon/e-books
```

-   Copy the snapshot (aka database file) to my main computer which happens to be a Mac.

```shell
# On my main computer

$ rsync -av user@nas:/home/user/mybooks.ajfs ~/

# Display info about the snapshot
$ ajfs info ~/mybooks.ajfs

Database path: /Users/andre/mybooks.ajfs
Version:       1
Root path:     /media/megaladon/e-books
OS:            linux
...
```

-   Next I want to know if I have already bought the book `100 Go Mistakes`.
    -   See [ajfs search](https://github.com/andrejacobs/ajfs/blob/main/docs/cli/md/ajfs_search.md) for more details.

```shell
$ ajfs search ~/mybooks.ajfs --iname '*go*'

# Unfortunately I have many directories and files that contain `go` in the base name
# Perhaps I can filter this down by only showing PDF files

$ ajfs search ~/mybooks.ajfs --type f --iname '*go*.pdf'

100 Go Mistakes/100_Go_Mistakes_and_How_to_Avoid_Them.pdf
...
```

-   I bought a couple more books and need to update the snapshot.

```shell
# New books where added to the NAS under the same directory
# /media/megaladon/e-books from which the database ~/mybooks.ajfs
# was created from.

# To update the snapshot:
nas$ ajfs update ~/mybooks.ajfs
```

## Index everything

I want to take a snapshot of all the files stored on my NAS including the file signature hashes. This operation will take hours to perform on over 10TB of data.

---

TODO:
To have a hash table or not?
Index everything with hashes, long running task
What's changed?
Finding duplicates
Figuring out what still needs to be backed up
