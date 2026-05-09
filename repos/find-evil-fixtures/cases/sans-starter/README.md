# SANS FIND EVIL — Starter Case Fixtures

Drop the official SANS starter case files in this directory.

## Source

- **Official URL**: <https://sansorg.egnyte.com/fl/HhH7crTYT4JK>
- **Cited at**: <https://findevil.devpost.com/resources>
- **Description**: "Sample disk images and memory captures provided for hackathon participants"

## Layout

This directory uses a split-storage layout to keep large binaries off the
project's primary partition:

```
repos/find-evil-fixtures/cases/sans-starter/
  README.md          (tracked — this file)
  MANIFEST.md        (tracked once generated — SHA256 + size + format hint per file)
  evidence/          (symlink → /mnt/disk1/INCUBATOR/warvis-findEvil-fixtures/sans-starter)
    <binary files>   (NOT tracked — live on /mnt/disk1)
```

## How to populate

1. Open the official URL in a browser (Egnyte share link is a folder/landing page, not a direct binary).
2. Download the disk images / memory dumps / log archives.
3. Place each file inside `evidence/` (the symlink resolves to /mnt/disk1):
   ```
   repos/find-evil-fixtures/cases/sans-starter/evidence/
     <disk-image>.E01
     <memory-dump>.raw
     <pcap>.pcap
     ...
   ```
4. Tell the orchestrator "files are placed" — B1 finalization runs:
   - MANIFEST.md generated (SHA256 + size + format hint per file)
   - dataset.md updated with "SANS Starter Case" section
   - tests/test_real_sample_acquisition.py written + verified

## Git policy

- This `README.md` is tracked.
- `MANIFEST.md` (when generated) is tracked.
- The `evidence/` symlink itself is tracked.
- Everything inside `evidence/` (the actual binaries on /mnt/disk1) is NOT tracked.

## License / Usage

SANS-provided case data; usage is governed by SANS hackathon Terms of Use. Do not redistribute outside this evaluation context.
