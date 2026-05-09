# SANS Starter Case — MANIFEST

This file inventories the SANS FIND EVIL official starter-case data acquired
for the warvis-findEval hackathon submission. Binaries themselves are stored
on a separate disk (`/mnt/disk1`) via the `evidence/` symlink and are NOT
tracked in git; their identity is anchored by the SHA256 hashes below.

## Provenance

- **Source URL**: <https://sansorg.egnyte.com/fl/HhH7crTYT4JK>
- **Cited at**: <https://findevil.devpost.com/resources>
- **Acquired by**: WoopsFactory (manual browser download, 2026-05-09 KST)
- **Storage path**: `/mnt/disk1/INCUBATOR/warvis-findEvil-fixtures/sans-starter/`
- **Repository symlink**: `repos/find-evil-fixtures/cases/sans-starter/evidence/`

## Files

### base-wkstn-05-memory (Workstation 5 physical memory dump)

| Field | Value |
|-------|-------|
| Original archive | `base-wkstn-05-memory.7z` |
| Archive size (bytes) | `655,932,396` (≈ 626 MB) |
| Archive SHA256 | `9e5184194499c01eddee7538ad23f5a7c74533c427ea387f70a249394cb3a4c2` |
| Archive type | 7-zip archive data, version 0.4 |
| Extracted artifact | `base-wkstn-05-memory.img` |
| Image size (bytes) | `3,221,225,472` (3 GiB exact) |
| Image SHA256 | `74ff679b25727d5fb7a8f70217d6fad965efd806260b7d224f0b38bd1c436115` |
| Image MD5 (SANS-recorded) | `bb6df5c0350d8014b718699f8c0c4fe0` |
| Image type | raw memory dump (`file` reports "data"; produced by dc3dd) |
| Companion file | `base-wkstn-05-memory.md5` (SANS dc3dd hash log) |
| Md5 log SHA256 | `822617f7f02b1683523f3d317268af55b8aecc025d42ee3f40ca13c13e742233` |

### Capture metadata (from companion `.md5` log)

```
dc3dd 7.2.641 started at 2018-09-06 19:51:09 +0000
command line: dc3dd if=/mnt/nfury/base-wkstn-05/pmem/pmem
              of=./base-wkstn-05-memory.img hash=md5 hlog=./base-wkstn-05-memory.md5
input  md5: bb6df5c0350d8014b718699f8c0c4fe0
dc3dd completed at 2018-09-06 19:52:02 +0000  (~53s wall, ~58 MB/s)
```

## Integrity verification

- The 7-zip archive's internal CRC was verified during `py7zr` extraction.
- The extracted image's SHA256 (`74ff67…36115`) is the value to use for
  cross-checks in subsequent UoWs.
- The original MD5 from `dc3dd` (`bb6df5…c4fe0`) is preserved here for forensic
  chain-of-custody traceability back to the SANS capture environment.

## What this case represents

`base-wkstn-05` is a **Windows workstation memory dump from the SANS SRL-2018
multi-host enterprise simulation**. The "base-" naming is a SANS convention
for the canonical case capture (not a "pre-incident baseline"); this folder is
a leaf node — there is no separate `incident-` sibling. Evil, if present, is
hidden in this dump as it would be in a real-world IR engagement.

## Scope (B1 deliverable)

This MANIFEST is the deliverable of UoW `real-sample-acquisition` (B1 of the
3-step `real-sample-integration` split). It only documents what was acquired.

- B2 (`real-tool-installation`): install yara/Volatility binary + smoke test.
- B3 (`real-hunt-trace`): run a Gemma 4 hunt against this image and capture
  the audit trail; update `accuracy-report.md` with measured findings.

## License / Usage

SANS-provided case data; usage is governed by SANS hackathon Terms of Use.
Do not redistribute outside this evaluation context.
