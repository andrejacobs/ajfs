mk2 will use a single flat file as the snapshot.

default scan mode is to not add space for the hash. Simply because a lot of the use case is to just take a snapshot and not worry about file hashes (which is computationally more expensive).
So unless you ask for checksums, you get a smaller file.
You can always update an existing snapshot to add the checksums.
impl: on update, backup existing file, stream new file with space for checksums (that will be calculated afterwards)

snapshot file verification
mk1 checksummed the entire file
I think mk2 should checksum the file hierarchy (without the file checksums/hashes) as well as the entire snapshot file.
Reason is, this way the file hierarchy can be quickly checked for changes and verified in the update/copy-over process.

when using file checksums
file hierarchy is saved first with space for the calculated hashes.
then the long running process of calculating file hashes is done and each file's record is updated.
updating should be able to continue where left off (meaning can cancel and restart file hash calcs).

when updating the snapshot to reflect current file hierarchy
new hierarchy is written and any existing hashes are copied as is (for file entries we believe have not been changed)
perhaps add a --force-update kind of thing to ensure hashes are recalculated regardless if meta changed.
