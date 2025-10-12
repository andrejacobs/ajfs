# TODO

-   [] Export, Diff and ToSync need to support -f --full paths.
-   [] Check that the vardata stuff respects endianess
-   [] Document each command and what the output means
-   [] Write a "system" test for each of the commands.
-   [] Could do with more testing on failure paths

-   [] Ensure go-collection is using the new stdlib maps packages

-   [] Build, test and fix so it can run on Windows as well

-   [] Use go mod vendor
-   [] Setup a github pages site and write end user documentation
-   [] Setup github action to check code quality etc
-   [] Add the usual polish for a public repo
-   [] Look into using goreleaser to build all the binaries etc for different platforms

-   [x] Remove the replace for local go mod repo.
-   [x] Info needs to display info about the hash table.
-   [x] Scan needs to output some verbose info.
-   [x] Scan needs to output progress info.
-   [x] Remove the idea of the tree being cached.
-   [x] Ensure list uses the same output and header as search when using hashes
-   [x] Should commands like list, search etc. by default be minimal output and then -m, --more is used for more details?
-   [x] Scan with only --progress needs a tiny bit more verbose info while scanning
-   [x] verbose and progress is a bit crazy. Perhaps when using progress, then don't output each file being hashed.
-   [x] Export needs verbose info
-   [x] Resume needs verbose info
-   [x] Add elapsed time when using verbose
-   [x] Add option to search to find Id
-   [x] Check for TODOs in the code and move them here.
-   [x] Tree: Think about options to support, i.e. limit, dirs only, full path
-   [x] Ensure all packages have a package level comment

## Future (nice to have)

-   [] `ajfs dedupe` To process and "deal" with duplicates. Will need different strategies.
    -   Delete, just simply delete any duplicates _Dangerous_
    -   Symlink, delete duplicates but leave a symlink in place
    -   Move duplicates to another directory while preserving hierarchy tree. Can thus be reviewed by a human.
-   [] Add support to search, to be able to parse a file (or from stdin) with more expressions like OR, NOT etc.
-   [] Add a flag for printing out memory usage.
-   [] Tree: Would be nice to also output colours.
    Turns out not to be a quick win to support LS_COLORS
    This is the best implementation for go I can find for reference:
    https://github.com/elves/elvish/blob/v0.21.0/pkg/cli/lscolors/lscolors.go
