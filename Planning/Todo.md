# TODO

- [] Need to spin up a VM for each of the Linuxes and test the install and app works.

- [] Fix the flaky unit tests!
- [] Write a "system" test for each of the commands.
- [] Could do with more testing on failure paths

- [] Build, test and fix so it can run on Windows as well

- [] Setup github action to check code quality etc

### Done

- [x] Remove the replace for local go mod repo.
- [x] Info needs to display info about the hash table.
- [x] Scan needs to output some verbose info.
- [x] Scan needs to output progress info.
- [x] Remove the idea of the tree being cached.
- [x] Ensure list uses the same output and header as search when using hashes
- [x] Should commands like list, search etc. by default be minimal output and then -m, --more is used for more details?
- [x] Scan with only --progress needs a tiny bit more verbose info while scanning
- [x] verbose and progress is a bit crazy. Perhaps when using progress, then don't output each file being hashed.
- [x] Export needs verbose info
- [x] Resume needs verbose info
- [x] Add elapsed time when using verbose
- [x] Add option to search to find Id
- [x] Check for TODOs in the code and move them here.
- [x] Tree: Think about options to support, i.e. limit, dirs only, full path
- [x] Ensure all packages have a package level comment
- [x] Add license headers
- [x] Export and ToSync need to support -f --full paths.
- [x] Check that the vardata stuff respects endianess
- [x] Fix linting errors (make check)
- [x] Document each command and what the output means
- [x] Specify explicit command ordering when using --help.
- [x] I need to test the following: What happens when you resume but some files have been deleted?
      Answer: The operation is stopped with ERROR: failed to calculate the hash for ...
- [x] What happens when you resume, but all the signatures have already been calced?
      Answer: It just works.
- [x] UPDATE: Nope. Also make mk1 public for historical reasons and slap a big "archived" sticker on it
- [x] Ignore vendor directory from code quality checks.
- [x] Fixed the bug where the root command's help function overwrites the output for subcommands.
- [x] Use go mod vendor
- [x] Add the usual polish for a public repo
- [x] What about installing on Linux (Ubuntu, Arch) UPDATE: Just use manual install.
- [x] Provide a different release note for the 1st release. UPDATE: goreleaser release --clean --release-notes ../release-ajfs.md
- [x] Don't generate i386 for Windows.
- [x] Make this repo public
- [x] Can install and works using brew on macOS.
- [x] Can install and works on Ubuntu x86-64 using manual install.

## Future (nice to have)

- [] `ajfs dedupe` To process and "deal" with duplicates. Will need different strategies.
    - Delete, just simply delete any duplicates _Dangerous_
    - Symlink, delete duplicates but leave a symlink in place
    - Move duplicates to another directory while preserving hierarchy tree. Can thus be reviewed by a human.
- [] Add support to search, to be able to parse a file (or from stdin) with more expressions like OR, NOT etc.
- [] Add a flag for printing out memory usage.
- [] Tree: Would be nice to also output colours.
  Turns out not to be a quick win to support LS_COLORS
  This is the best implementation for go I can find for reference:
  https://github.com/elves/elvish/blob/v0.21.0/pkg/cli/lscolors/lscolors.go
- [] Setup a github pages site and write end user documentation. UPDATE: Most end user docs are in the ./docs dir.
- [] Strip out the ``` from the Long description for dupes. This is used to ensure the markdown docs look correct.

## Removed from scope

- [] Test out the shell completion support. Also look into getting homebrew to install these.
- [] Ensure homebrew releases are signed. https://goreleaser.com/customization/homebrew_casks/
- [] Provide an alternative README.md for the packaged up releases.
