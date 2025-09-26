This directory contains the minimal test data that is being used by the unit-tests.

You will need to have hashdeep installed to generate the file hash signatures. `brew install hashdeep`

Run `./setup.sh` to generate the diff, need-sync and expected directories as well as the expected file hash signatures.
If you modify any of the files then you need to run the `./generate-expected-hashes.sh`.

PS: This was taken as is from mk1