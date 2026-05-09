# Plan: real-sample-acquisition — SANS Starter Case Download & Integration

**UoW ID**: `real-sample-acquisition` (B1 of 3-step split for P0 real-sample-integration)  
**Dev Session**: `5dccb3d0-bc23-4cb8-8815-3787edc82759`  
**Project**: warvis-findEvil  
**Date**: 2026-05-09

---

## Context

### Spec Source
- **Title**: Acquire and verify SANS FIND EVIL official starter case data
- **Motivation**: Largest remaining gap per `docs/find-evil/valhuntir-comparison.md §6 A1` — real samples needed to validate Phase 3 accuracy claims
- **URL**: https://sansorg.egnyte.com/fl/HhH7crTYT4JK (cited on https://findevil.devpost.com/resources)
- **Deadline**: 2026-06-01 (Phase D-15 recall gate)

### Tech Stack
- **Python**: `python -m pytest`, `ruff check`
- **Test entry**: `tests/test_real_sample_acquisition.py` (to be created)
- **Integration point**: Append to `docs/find-evil/dataset.md` (current Phase 2-era content)

### Critical Discovery (URL Probe)
```
curl -sI https://sansorg.egnyte.com/fl/HhH7crTYT4JK
→ HTTP/2 200
→ content-type: text/html;charset=UTF-8  [NOT binary!]
```
**Conclusion**: Egnyte URL is a **folder/landing page**, not direct file download. Requires:
- Browser-based navigation + manual download, OR
- Egnyte API key (if available), OR
- User-provided pre-downloaded files

**Risk assessment**: MEDIUM (user intervention required; scope depends on file size/complexity)

### Gitignore Policy
Current `.gitignore` does NOT ignore `repos/` — large fixture binaries must be explicitly gitignored.
Policy decision: **MANIFEST.md tracked; fixture binaries gitignored** (parallel to case-001 pattern in docs).

---

## Scope & Constraints

### In Scope
1. Create directory structure: `repos/find-evil-fixtures/cases/sans-starter/`
2. **Acquire** files from Egnyte (manual or API-driven)
3. **Verify** integrity: SHA256 checksums in MANIFEST.md
4. Document each file: size, mime/format, retrieval timestamp, source URL
5. Append "## SANS Starter Case" section to `docs/find-evil/dataset.md`
6. Write test suite: `tests/test_real_sample_acquisition.py`

### Out of Scope (B2/B3 handle)
- Installing yara/plaso/volatility3/zeek binaries
- Running actual hunts against real samples
- Updating accuracy-report.md with new measurements
- Processing / analyzing the samples

### No-Go Constraints
- **Never break** `src/find_evil_mcp/` (Phase 1+2 complete, stable)
- Do not force large downloads if file >10 GB without user confirmation
- Do not commit fixture binaries to git

---

## Milestones

### M1: Establish Directory Structure & Gitignore Policy
**Objective**: Create fixture hierarchy and suppress large binaries from git tracking

**Tasks**:
- [ ] **TDD Red**: Write test expecting `repos/find-evil-fixtures/cases/sans-starter/` to exist with MANIFEST.md
- [ ] **TDD Green**: Create directory tree and `.gitignore` rule
- [ ] **TDD Refactor**: Verify git status shows MANIFEST.md as tracked, evidence/ as ignored

**Validation**: 
```bash
pytest tests/test_real_sample_acquisition.py::TestDirectoryStructure -v
```

**Risk**: LOW — filesystem ops only, no network

---

### M2: Probe & Document Egnyte Acquisition Path
**Objective**: Determine how to retrieve files; document the method in MANIFEST.md preamble

**Tasks**:
- [ ] **TDD Red**: Write test expecting MANIFEST.md to contain acquisition method + timestamp
- [ ] **TDD Green**: Probe Egnyte URL (curl -sIL), document findings in MANIFEST.md
- [ ] **TDD Refactor**: Add clear acquisition instructions (manual, API, or user-provided)

**Validation**:
```bash
pytest tests/test_real_sample_acquisition.py::TestManifestPreamble -v
```

**Risk**: MEDIUM — If auth required, work pauses for user credentials/files

---

### M3: Acquire & Validate File Integrity
**Objective**: Download/place files in `evidence/` and record SHA256 + metadata

**Tasks**:
- [ ] **TDD Red**: Write test expecting MANIFEST.md to contain SHA256 hashes (64-char hex)
- [ ] **TDD Green**: Compute SHA256 for all files, populate MANIFEST.md table
- [ ] **TDD Refactor**: Verify each hash matches file on disk; check disk space

**Validation**:
```bash
pytest tests/test_real_sample_acquisition.py::TestFileIntegrity -v
```

**Risk**: MEDIUM-HIGH — File sizes unknown; stop-and-ask if >10 GB

---

### M4: Append Documentation Section
**Objective**: Update `docs/find-evil/dataset.md` with SANS case section

**Tasks**:
- [ ] **TDD Red**: Write test expecting "sans-starter" + source URL in dataset.md
- [ ] **TDD Green**: Append "## SANS Starter Case" section with metadata
- [ ] **TDD Refactor**: Verify markdown syntax; check links

**Validation**:
```bash
pytest tests/test_real_sample_acquisition.py::TestDatasetDocumentation -v
```

**Risk**: LOW — documentation only

---

### M5: Write Comprehensive Test Suite
**Objective**: Implement `tests/test_real_sample_acquisition.py` with all 4 acceptance criteria

**Tasks**:
- [ ] **TDD Red**: Stub all test classes; run to confirm failures
- [ ] **TDD Green**: Implement 4 test methods
  - test_manifest_exists
  - test_manifest_contains_sha256
  - test_evidence_directory_nonempty
  - test_dataset_md_updated
- [ ] **TDD Refactor**: Add docstrings, improve assertions

**Validation**:
```bash
python -m pytest tests/test_real_sample_acquisition.py -v
ruff check tests/test_real_sample_acquisition.py
```

**Risk**: LOW — test code only

---

## Stop-and-Fix Rule

Before advancing to next milestone, current milestone must achieve green status:

| M | Validation | Failure Action |
|---|-----------|-----------------|
| M1 | `pytest tests/test_real_sample_acquisition.py::TestDirectoryStructure -v` | Fix structure; retest |
| M2 | `pytest tests/test_real_sample_acquisition.py::TestManifestPreamble -v` | Document method; retest |
| M3 | `pytest tests/test_real_sample_acquisition.py::TestFileIntegrity -v` | Verify hashes; retest |
| M4 | `pytest tests/test_real_sample_acquisition.py::TestDatasetDocumentation -v` | Edit dataset.md; retest |
| M5 | `pytest tests/test_real_sample_acquisition.py -v && ruff check tests/test_real_sample_acquisition.py` | Fix test/lint; retest |

---

## Done When

- [x] M1: Directory structure created + `.gitignore` applied
- [x] M2: MANIFEST.md preamble documents acquisition method + timestamp
- [x] M3: All files have SHA256 checksums in MANIFEST.md
- [x] M4: `docs/find-evil/dataset.md` appended with "## SANS Starter Case"
- [x] M5: `tests/test_real_sample_acquisition.py` passes all 4 criteria
- [x] **Lint**: `ruff check` clean on all modified Python files
- [x] **Integration**: No regressions in `python -m pytest`
- [x] **devos_verify_dev_session**: All gates report PASS

---

## Risk Assessment

| Risk Factor | Level | Mitigation |
|-------------|-------|-----------|
| **Egnyte URL is landing page** | MEDIUM | M2 probes; documents manual vs API path |
| **File size unknown; disk insufficient** | MEDIUM | M3 stops if file >10 GB; ask user |
| **Auth required by Egnyte** | MEDIUM | M2 documents credential requirement |
| **Real samples may not match layout** | LOW | MANIFEST.md flexible; document any structure |
| **Break Phase 1+2 stability** | LOW | Gitignore + test isolation ensure safety |

**Overall Risk**: **MEDIUM** → **LOW after M2** (once acquisition path clear)

---

## Implementation Notes

### Gitignore Update
```
# Large fixture binaries (tracked via MANIFEST.md checksums)
repos/find-evil-fixtures/cases/*/evidence/**
!repos/find-evil-fixtures/cases/*/MANIFEST.md
```

### MANIFEST.md Format
- **Preamble**: Acquisition method, timestamp, source URL
- **File Table**: name, size, SHA256, mime, retrieval date
- **Ground-Truth Labels**: expected_findings count per tool (if SANS provides)
- **Provenance**: License, restrictions, curation notes

### Test Isolation
- Do NOT mock `src/find_evil_mcp/` tools
- Do NOT modify existing Phase 2 fixtures
- Do NOT create network calls (except M2 HEAD probe)

### Verification Strategy
- **Unit**: `pytest tests/test_real_sample_acquisition.py -v`
- **Integration**: `python -m pytest` (full suite)
- **Lint**: `ruff check src/ tests/ docs/`
- **Security**: No secrets in MANIFEST.md or dataset.md

---

## Status Tracking

| M | Title | Status |
|---|-------|--------|
| M1 | Directory structure & gitignore | TODO |
| M2 | Egnyte probe & method documentation | TODO |
| M3 | File acquisition & SHA256 validation | TODO |
| M4 | Documentation append (dataset.md) | TODO |
| M5 | Test suite (test_real_sample_acquisition.py) | TODO |

---

## References

- **Spec**: UoW real-sample-acquisition (B1 of 3-step split)
- **Dataset Doc**: `docs/find-evil/dataset.md`
- **Accuracy Report**: `docs/find-evil/accuracy-report.md`
- **Egnyte URL**: https://sansorg.egnyte.com/fl/HhH7crTYT4JK
- **Devpost Resource**: https://findevil.devpost.com/resources

---

**Status**: shipped (2026-05-09, base-wkstn-05-memory.img acquired + integrity verified, MANIFEST.md committed, 55/55 tests PASS, lesson_id=fc970b48-c44f-4d59-a623-b318b931ca6b)
**Risk Level**: LOW (Egnyte browser-gating discovered up front, human-in-the-loop path executed cleanly)
