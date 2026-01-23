# Examples of using ajfs

The core concept to remember is that you use ajfs to take a snapshot of a file system hierarchy (directory layout,
filenames, size, etc.) that you can then use to later perform things such as searching or finding duplicates.
You can perform these operations on a **completely different machine** than the one on which the file system hierarchy lives.
Please note that a snapshot does not include the actual file content data.

When creating a new snapshot you have the choice of including the file signature hashes or not. The file signature hash
is the value obtained by reading every byte from the file and calculating a hash using an algorithm such as SHA-256.
Certain features of ajfs require the file signature hashes to be present in the snapshot.

The following examples use:

-   `$` to indicate your local machine and `remote$` to indicate another machine.
-   the words `I`, `my` etc. written as if it was you that performed the operations.
-   `...` indicate that there is more output and not included in the example.

## Installing ajfs

TODO:

## Offline search

I have a big collection of E-books that I have bought from a couple of sources over many years and store these on
a NAS running Linux. Before I buy any new books I need a way to quickly check if I have not already bought the book.
This needs to be done from my main computer.

-   First I need to take a snapshot of the existing books.

    -   See [ajfs scan](cli/md/ajfs_scan.md) for more details.

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

    -   See [ajfs search](cli/md/ajfs_search.md) for more details.

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

I want to take a snapshot of all the files stored on my NAS including the file signature hashes.
This operation will take hours to perform on over 15TB of data.
I also want to be able to check in every few hours and see the progress being made.
It also happens to be that the SHA-1 algorithm performs much better on this "old" x86-64 machine.

-   Start the scan.

    ```shell
    nas$ ajfs scan --hash --algo=sha1 --progress ~/database.ajfs /media

    Scanning ...
    Calculating progress information ...
    [1525/4464882]   0% |                | (12 GB/15 TB, 149 MB/s) [1m12s:27h16m8s]
    ```

-   After a few minutes I realise that I need to run this in `tmux` or `screen` so that the process doesn't get terminated when the connection gets dropped.
-   `Ctrl c` to stop the process.
-   Start a `tmux` or `screen` session.
-   Continue the process to calculate the file signature hashes. This will still take +/- 27 hours!

    ```shell
    nas$ ajfs resume --progress ~/database.ajfs

    Resuming database file at "./db.ajfs"
    Calculating progress information ...
    [1589/4464882]   0% |                | (14 GB/15 TB, 152 MB/s) [9s:26h38m17s]
    ```

-   Now the process can be interuppted at any point and I can simply resume when needed.

-   At some point I want to update the snapshot but really don't want to wait another 27 hours.

    -   [ajfs update](cli/md/ajfs_update.md) will create a new
        snapshot but only calculate file signature hashes for new files.

    ```shell
    nas$ ajfs update --progress ~/database.ajfs
    ```

## What has changed?

I need to see which files are being modified when I run a certain program.

-   See [ajfs diff](cli/md/ajfs_diff.md) for more details.
-   Create the initial snapshot.

    ```shell
    $ ajfs scan ~/snap1.ajfs ~/.config
    ```

-   Assume I run a program called `bravo` for the first time.
-   I want to see which files this program created.

    ```shell
    $ ajfs diff ~/snap1.ajfs

    d++++ bravo
    f++++ bravo/bravo.settings
    f++++ bravo/cached_ids.json
    d~sl~ .
    ```

-   I can see that it has created a new directory called bravo and two new files.
-   Take another snapshot.

    ```shell
    $ ajfs scan ~/snap2.ajfs ~/.config
    ```

-   Assume I make changes to `bravo` and run it again.
-   I want to see which files this program modified.

    ```shell
    $ ajfs diff ~/snap2.ajfs

    f---- bravo/cached_ids.json
    d++++ bravo/data
    f++++ bravo/data.xyz
    f++++ bravo/data/1.json
    f++++ bravo/data/2.json
    f~sl~ bravo/bravo.settings
    d~sl~ bravo
    ```

-   From this I can see that the file 'cached_ids.json' was deleted. A new directory and files were created under 'data'. The 'bravo.settings' was changed (size and last modification time).

-   I can also see the differences between two snapshots.

    ```shell
    $ ajfs diff ~/snap1.ajfs ~/snap2.ajfs
    ...
    ```

## Browse a snapshot

-   I want to see what file and directory information was stored in a snapshot.

    -   See [ajfs list](cli/md/ajfs_list.md) for more details.
    -   Remember you can also perform this on another computer with only the .ajfs database file.

    ```shell
    ajfs list ~/snap3.ajfs

    .
    bravo
    bravo/bravo.settings
    bravo/data
    bravo/data/1.json
    bravo/data/2.json
    bravo/data.xyz
    mc
    mc/ini
    mc/mcedit
    mc/panels.ini
    ...
    ```

-   I want to see this information in a `tree` like manner.

    -   See [ajfs tree](cli/md/ajfs_tree.md) for more details.

    ```shell
    $ ajfs tree ~/snap3.ajfs

    /Users/andre/.config
    ...
    ├── bravo
    │   ├── bravo.settings
    │   ├── data
    │   │   ├── 1.json
    │   │   └── 2.json
    │   └── data.xyz
    ├── mc
    │   ├── ini
    │   ├── mcedit
    │   └── panels.ini
    ...
    ```

-   I want to only see the tree for a certain directory.

    ```shell
    $ ajfs tree ~/snap3.ajfs bravo/data

    data
    ├── 1.json
    └── 2.json

    1 directory, 2 files
    ```

## Finding duplicates

I have the problem where a lot of data was backed up over the years to a number of different locations on my NAS and
I want to find these duplicates.

-   See [ajfs dupes](cli/md/ajfs_dupes.md) for more details.
-   First create a snapshot. I will be including the file signature hashes so that I can find duplicates even if the filenames are different.

    ```shell
    nas$ ajfs scan --hash --algo=sha1 --progress ~/database.ajfs /media
    ```

-   I want to see all the subtrees that are the same. This is a good indicator of where certain directories were copied into different locations over time.

    ```shell
    nas$ ajfs dupes --dirs ~/database.ajfs
    ...
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
    ...
    # I can see that I have a duplicate set of photos backed up under a slightly different directory structure
    ```

-   I want to find all the duplicate files that might have different filenames but the exact same file data.

    ```shell
    nas$ ajfs dupes ~/database.ajfs

    >>>
    Hash: c88e6e3b20f8478468288d2bef9cf624f5707ebcdad6113d4a545469333271a1
    Size: 11167407 [11 MB]

    [0]: Bought/Books/Some-awesome-book.pdf
    [1]: Another/dir/somewhere/with/backups/Same-book.pdf

    Count: 2
    Total Size: 22334814 [22 MB]
    <<<

    # I can see that I have 2 books with different locations and filenames, however the content is the exact same.
    # A single instance of the data would only take up 11MB. In this case the total size including all duplicates are 22MB.
    ```

## Export to other formats

A snapshot can also be exported to other formats like: CSV, JSON and hashdeep.

-   See [ajfs export](cli/md/ajfs_export.md) for more details.

    ```shell
    $ ajfs export ~/database.ajfs export.csv
    ```

## What needs to be backed up?

Scenario 1: I need to figure out which files on my laptop has not yet been backed up to my NAS. Unfortunately my files on the NAS
could also be located in different directories and even have different filenames.

-   In this case I need to take snapshots of both machines as well as include file signature hashes.

    ```shell
    $ ajfs scan --hash ~/laptop.ajfs ~/
    nas$ ajfs scan --hash ~/nas.ajfs /media/backup
    ```

-   I can now get the list of files that still need to be backed up.

    ```shell
    # I copied the nas.ajfs file from the NAS onto my laptop
    $ ajfs tosync --hash ~/laptop.ajfs ~/nas.ajfs
    ...
    Cached/E-books/Go/100_go_mistakes.pdf
    Cached/E-books/Go/100_go_mistakes.epub
    ...
    ```

Scenario 2: I have a local copy of a file hierarchy that is also mirrored on my NAS. Unfortunately I am not on the
same network as my NAS and thus can't run `rsync --dry-run` to find out which files still need to be backed up.
I do however have a snapshot from my NAS. In scenario 1 the files could be stored anywhere on the NAS and I just wanted
to ensure I at least backed up my local files somewhere on the NAS. This scenario does not require the file signature
hashes to be calculated.

-   Get the list of files that still need to be backed up.

    ```shell
    $ ajfs tosync ~/laptop.ajfs ~/nas.ajfs
    ```

-   Which files exist on the NAS that I have deleted locally?

    ```shell
    $ ajfs tosync ~/nas.ajfs ~/laptop.ajfs
    ```
