# TODO

-   [] Remove the replace for local go mod repo.
-   [x] Info needs to display info about the hash table.
-   [x] Scan needs to output some verbose info.
-   [x] Scan needs to output progress info.
-   [] Export needs verbose info
-   [] Resume needs verbose info
-   [] Could do with more testing on failure paths
-   [] verbose and progress is a bit crazy. Perhaps when using progress, then don't output each file being hashed.
-   [] Add elapsed time when using verbose
-   [] Add a flag for printing out memory usage.
-   [] Export, Diff and ToSync need to support -f --full paths.
-   [] Ensure list uses the same output and header as search when using hashes
-   [] Add option to search to find Id
-   [] Should commands like list, search etc. by default be minimal output and then -m, --more is used for more details?
-   [] Check that the vardata stuff respects endianess
-   [] Ensure all packages have a package level comment
-   [] Document each command and what the output means

-   [] Ensure go-collection is using the new stdlib maps packages

## Future (nice to have)

-   [] Add support to search, to be able to parse a file (or from stdin) with more expressions like OR, NOT etc.
