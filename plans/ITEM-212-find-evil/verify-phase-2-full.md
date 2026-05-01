# ITEM-212: FIND-EVIL — Phase 2 Full Verification Report

**Verdict**: **PASS**

**Verification timestamp**: 2026-04-30 11:30 UTC  
**Verifier context**: find-evil-verifier (read-only audit lane)  
**Verified-against commit**: 752c979 (phase-0 scaffold)

---

## Executive Summary

All 4 required Phase 2 gates **PASS**. Detector wiring, 9-tool MCP server, test suite (90 cases), and accuracy harness are operationally complete. Synthetic recall reaches 100% on test fixture; production recall unmeasured. Solidity gates (slither/aderyn) are N/A for Python package and require `applies_when: language=solidity` in gates.yaml for Phase 3.

---

## Phase 2 Gates — Detailed Verification

| Gate ID | Check Type | Target/Metric | Expected | Observed | Status | Evidence |
|---------|-----------|-------------|----------|----------|--------|----------|
| **tools_full_count** | schema_count | `harness/find-evil/mcp-schema/*.json` | ≥ 9 | **9** | ✅ PASS | ls count = 9 |
| **evil_detection_recall_min** | float_ge | recall metric | ≥ 0.60 | **1.00** (100%) | ✅ PASS | `make accuracy` → 13/13 detections |
| **slither_high** | int_eq | slither_high metric | 0 | N/A | ⚠️ PASS_WITH_WARN | Python package, no .sol files in src/find_evil_mcp/ |
| **aderyn_high** | int_eq | aderyn_high metric | 0 | N/A | ⚠️ PASS_WITH_WARN | Python package, no .sol files in src/find_evil_mcp/ |

---

## Re-Executed Validation Commands

### 1. Build Gate
```bash
$ make -C harness/find-evil build
make: Entering directory '/home/jang/Workspace/warvis-forRich/harness/find-evil'
build: PASS
make: Leaving directory '/home/jang/Workspace/warvis-forRich/harness/find-evil'
```
✅ **PASS**

### 2. Test Suite (90 cases)
```bash
$ make -C harness/find-evil test
...
tests/test_schema.py::TestCaseOpenSchema::test_case_open_valid_image_path_input PASSED
tests/test_schema.py::TestCaseOpenSchema::test_case_open_valid_mem_path_input PASSED
...
tests/test_detector.py::TestTimelineQuery::test_parse_timeline_respects_limit PASSED

============================== 90 passed in 0.24s ==============================
```
✅ **PASS** — All 90 tests passing (no regressions from prior partial verify)

### 3. MCP Conformance Handshake
```bash
$ make -C harness/find-evil mcp-conformance
make: Entering directory '/home/jang/Workspace/warvis-forRich/harness/find-evil'
mcp-conformance: PASS — find-evil-mcp v0.1.0 
  tools: ['case.open', 'iocs.scan', 'log.query', 'memory.malfind', 
          'memory.process_list', 'net.flow_summary', 'report.append', 
          'timeline.build', 'verify.cross_check']
make: Leaving directory '/home/jang/Workspace/warvis-forRich/harness/find-evil'
```
✅ **PASS** — 9 tools registered, capability handshake valid

### 4. Tool Schema Count
```bash
$ ls -la harness/find-evil/mcp-schema/*.json | wc -l
9
```
✅ **PASS** — Exactly 9 schema files

### 5. Fixture Ground Truth
```bash
$ python3 -c "import json; m=json.load(open('repos/find-evil-fixtures/cases/case-001/manifest.json')); 
  print('Expected findings:', sum(len(v) for v in m['expected_findings'].values()))"
Expected findings: 13
```
✅ **PASS** — 13 expected findings in case-001 manifest

### 6. Accuracy Harness (Recall Measurement)
```bash
$ make -C harness/find-evil accuracy
case: /mnt/disk1/forrich/data/find-evil-fixtures/cases/case-001
threshold: 0.60
  [OK] iocs.scan: 2/2
  [OK] log.query: 4/4
  [OK] memory.process_list: 3/3
  [OK] memory.malfind: 1/1
  [OK] net.flow_summary: 3/3
recall: 100.00%
```
✅ **PASS** — Recall = 13/13 = 100% >> 0.60 threshold

### 7. Tool Forbidden-Tool Guard
Custom verification script output:
```
[INFO] Total tools: 9
[INFO] Tool names: ['case.open', 'iocs.scan', 'log.query', 'memory.malfind', 
                     'memory.process_list', 'net.flow_summary', 'report.append', 
                     'timeline.build', 'verify.cross_check']
[OK] No forbidden tools detected
```
✅ **PASS** — No exec/shell/subprocess_unsafe tools present

### 8. All Tools Have outputSchema
Custom verification script output:
```
[OK] case.open: has outputSchema
[OK] iocs.scan: has outputSchema
[OK] log.query: has outputSchema
[OK] memory.malfind: has outputSchema
[OK] memory.process_list: has outputSchema
[OK] net.flow_summary: has outputSchema
[OK] report.append: has outputSchema
[OK] timeline.build: has outputSchema
[OK] verify.cross_check: has outputSchema

[OK] All 9 tools have non-None outputSchema
```
✅ **PASS** — 9/9 tools have non-None outputSchema dict

### 9. Solidity Gate Applicability
```bash
$ find src/find_evil_mcp -name "*.sol" -type f | wc -l
0
```
**N/A** — find-evil-mcp is a pure Python MCP server; no Solidity contracts in scope.  
Slither and Aderyn checks apply only to DeFi smart contract code, not DFIR tooling.

---

## Honest Caveats & Recommendations

### ⚠️ Caveat 1: Synthetic Recall
The 100% recall measurement is achieved against a **hand-crafted test fixture** (`case-001`) with known signal distribution:
- 2 IOC matches (YARA-alike)
- 4 log query hits
- 3 process list detections
- 1 malfind signature
- 3 network flows

This does **NOT** constitute production DFIR recall validation. Real forensic image samples (SANS DFIR datasets, actual incident evidence) are unmeasured. The lite scanners in `accuracy.py` reproduce synthetic signal in the absence of plaso/yara/volatility3/zeek; real tools are not available in dev environment and are monkeypatched in tests.

**Mitigation for Phase 3**: Phase 3 must source a SANS dataset sample or equivalent forensic image with ground-truth annotations to measure real-world recall.

---

### ⚠️ Caveat 2: Lite Scanner Mode
The accuracy harness uses pure-Python lite detectors (regex+JSON parsing) that emulate real DFIR tool output:

| Tool | Real Path | Lite Emulation | Status |
|------|-----------|----------------|--------|
| `plaso` | /tools/plaso | regex JSON parser | ✅ Monkeypatched in tests |
| `yara` | /tools/yara | YARA rule simulator | ✅ Emulated |
| `volatility3` | /tools/volatility3 | pslist mock | ✅ Emulated |
| `zeek` | /tools/zeek | conn.log parser | ✅ Emulated |

This is **acceptable for Phase 2** (architecture validation) but **insufficient for Phase 3** (production readiness). Real tool integration requires:
1. Actual tool binaries or Docker containers
2. Integration test against real pcap/memory/evidence samples
3. Cross-validation vs. published forensic reports

---

### ⚠️ Caveat 3: Solidity Gates Are N/A
`slither_high == 0` and `aderyn_high == 0` gates in `gates.yaml` apply only to Solidity smart contracts. The find-evil-mcp codebase is **pure Python** (MCP server + DFIR detectors); no Solidity code is present.

**Recommendation**: Update `gates.yaml` Phase 2 to conditionally gate Slither/Aderyn:
```yaml
phase_2:
  required:
    ...
    - id: slither_high
      check: int_eq
      metric: slither_high
      value: 0
      applies_when: language=solidity  # NEW: skip for Python
    - id: aderyn_high
      check: int_eq
      metric: aderyn_high
      value: 0
      applies_when: language=solidity  # NEW: skip for Python
```

Mission-control should action this patch before Phase 3 verifier consumes gates.yaml.

---

## Phase 3 Entry Blockers & Recommendations

### Open Risks
1. **Real-data recall unmeasured** — Synthetic fixture ≠ production DFIR samples. Phase 3 must source SANS/ICS-CERT forensic images.
2. **Lite scanners are development mode** — Real plaso/yara/volatility3/zeek integration required for ship gate (recall ≥ 0.80 on real samples).
3. **8-artifact submission spec unclear** — Devpost page, demo script, and dataset documentation must be authored. Current spec.md §6 describes *what* to deliver but not *where* in repo each artifact lives.

### Phase 3 Pre-Flight Checklist
- [ ] Source forensic sample dataset (SANS DFIR or equivalent) with >10 cases + ground-truth findings
- [ ] Integrate real DFIR tools (plaso, yara, volatility3, zeek) or containerized equivalents
- [ ] Run `make accuracy` against real samples; confirm recall ≥ 0.80 on ≥3 diverse cases
- [ ] Author 6 Devpost submission artifacts (README updates, architecture.md, dataset.md, accuracy-report.md, devpost-page.md, demo-script.md)
- [ ] Validate all 8 required files exist in repo (incl. LICENSE)
- [ ] Confirm license is MIT or Apache-2.0 (gates.yaml Phase 3 requirement)
- [ ] Patch gates.yaml with `applies_when: language=solidity` on Slither/Aderyn gates

---

## Verdict Justification

**PASS** is issued because:
1. ✅ tools_full_count = 9 (required ≥ 9)
2. ✅ evil_detection_recall_min = 100% (required ≥ 0.60)
3. ⚠️ slither_high = N/A (inapplicable to Python; gate passed vacuously)
4. ⚠️ aderyn_high = N/A (inapplicable to Python; gate passed vacuously)

All 4 required gates achieve their thresholds or are legitimately inapplicable. No deterministic gates (spec.md §6 recall, capability, license) are violated.

---

## Files Verified

- `/home/jang/Workspace/warvis-forRich/harness/find-evil/gates.yaml` (Phase 2 gate spec)
- `/home/jang/Workspace/warvis-forRich/harness/find-evil/Makefile` (test/accuracy targets)
- `/home/jang/Workspace/warvis-forRich/src/find_evil_mcp/server.py` (9-tool MCP registration)
- `/home/jang/Workspace/warvis-forRich/tests/test_detector.py` (90 test cases)
- `/home/jang/Workspace/warvis-forRich/repos/find-evil-fixtures/cases/case-001/manifest.json` (13 ground-truth findings)
- `/home/jang/Workspace/warvis-forRich/harness/find-evil/scripts/accuracy.py` (recall harness)

---

**Next gate**: Phase 3 verifier (scheduled post-handoff) will validate 8 submission artifacts, real-data recall ≥ 0.80, and license compliance.

