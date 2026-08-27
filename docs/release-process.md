# Release process

Release tags are immutable publication inputs. A tag must only be created after
the exact commit has passed the `Release candidate` workflow.

1. Update release notes in `docs/releases/<tag>.md` and merge all release work
   into `main`.
2. Copy the full 40-character SHA at the tip of `main`.
3. Dispatch `Release candidate` with that SHA. The workflow rejects stale or
   abbreviated SHAs and verifies source, race tests, vulnerabilities, the
   GoReleaser snapshot, and the candidate container.
4. Create the version tag on the same verified SHA. Do not move or recreate a
   published tag.
5. Confirm that the `Release` workflow published archives, checksums, release
   notes, and both container registries before updating downstream packages.

For recovery when an immutable tag exists but its publishing run is missing,
use the `publish` mode documented in the release notes. This recovery path
checks out and verifies the existing tag; it never moves the tag. Recovery
replaces the existing release notes with the curated document and replaces any
same-named assets, so a partially completed publication can be retried safely.
