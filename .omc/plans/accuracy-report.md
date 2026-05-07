# Plan: accuracy-report — Phase 4 Accuracy Report Generation

## Context
- **project_id**: warvis-findEval
- **dev_session_id**: 64079e0e-e2ab-4010-9921-dab8c0508bbc
- **uow_id**: accuracy-report
- **tech_stack**: Python (src/find_evil_mcp) + Go (warvis/, locked post 2026-05-09) + accuracy.py harness
- **ssot_source**: inline spec + handover-r4.md + verify-phase3.md

## Scope

### Create
- `docs/find-evil/accuracy-report.md` (single deliverable, ≤3000 words)

### Modify
- `tests/test_accuracy_report.py` (NEW test — validates deliverable structure)

### No-Go (LOCKED)
- `src/find_evil_mcp/` — Phase 1+2 complete, zero changes
- `warvis/` code — Phase 3 locked post 2026-05-09; documentation only

### Constraints
- SANS real DFIR samples NOT available this cycle → cannot measure recall ≥ 0.80 on real samples
- Synthetic case-001: 13 ground-truth findings hardcoded in fixture
- `harness/find-evil/scripts/accuracy.py` already measures lite-mode (pure Python emulation)
- Must document lite-vs-real divergence **honestly** (this is the critical section)

## Milestones

### M1: Research & Fixture Analysis
**Goal**: Understand case-001 test fixture, ground truth, and accuracy.py harness behavior.

- **Tasks**:
  - [ ] Examine `repos/find-evil-fixtures/cases/case-001/manifest.json` → extract 13 ground-truth findings
  - [ ] Run `python harness/find-evil/scripts/accuracy.py --case repos/find-evil-fixtures/cases/case-001 --json` → capture baseline output
  - [ ] Document the 5 lite-mode scanner categories: iocs.scan, log.query, memory.process_list, memory.malfind, net.flow_summary
  - [ ] Read `harness/find-evil/scripts/accuracy.py` lines 40–203 → understand lite-scanner implementation
  - [ ] Identify which 13 findings map to which scanner (per manifest)

- **Validation**: `pytest -xvs tests/test_accuracy_report.py::test_m1_fixture_integrity`
  - ✅ manifest.json loads without error
  - ✅ expected_findings dict has ≥13 entries
  - ✅ accuracy.py runs successfully with default case-001

- **Risk**: LOW (read-only, no code changes)

### M2: Synthetic Path Documentation
**Goal**: Document lite-mode recall measurement (100% on case-001) with exact reproduction steps.

- **Tasks**:
  - [ ] Write "Executive Summary" section (2-3 paras)
    - Recall result: 100% on synthetic case-001 (lite-mode)
    - Acknowledge: real DFIR samples unavailable → recall on production tools unmeasured
    - Honest caveat: "This cycle measures only synthetic recall via pure-Python emulation"
  - [ ] Write "Methodology: Test Fixture (case-001)" section
    - Describe 13 ground-truth findings by category (yara rules, log queries, memory PIDs, network flows)
    - Reference manifest.json structure
    - Explain why synthetic-only (cost + time constraints)
  - [ ] Write "Results: Lite-Mode Recall Table" section
    - 5-row table: iocs.scan | log.query | memory.process_list | memory.malfind | net.flow_summary
    - Columns: Tool | Expected | Matched | Recall (%)
    - All show 100% recall on case-001
  - [ ] Write "Reproduction: Command & Expected Output" section
    - Exact command: `make -C harness/find-evil accuracy`
    - Or: `python harness/find-evil/scripts/accuracy.py --case repos/find-evil-fixtures/cases/case-001`
    - Expected output snippet showing 100% recall + all tools green

- **Validation**: `pytest -xvs tests/test_accuracy_report.py::test_m2_synthetic_section`
  - ✅ File exists, contains "recall"
  - ✅ Contains "synthetic" and "case-001"
  - ✅ Reproduction command is valid & runnable
  - ✅ Table present with ≥5 rows

- **Risk**: LOW (documentation only)

### M3: Lite-vs-Real Path Divergence (CRITICAL)
**Goal**: Honestly document where synthetic path diverges from real tool path, and why.

- **Tasks**:
  - [ ] Write "Lite-vs-Real Path Divergence" section (2-3 paras, THIS IS THE KEY SECTION)
    - Explain: lite-mode uses pure-Python string matching + binary parsing (iocs.scan = substring search, memory = JSON header parsing, net = PCAP struct parsing)
    - Real tools: YARA engine, plaso/log2timeline event reconstruction, Volatility3 memory forensics, Zeek network analysis
    - **Why divergence matters**: "lite-mode is sufficient for architecture validation; real tools required for production IR"
    - **Evaluator concern**: "Real tool output validation NOT measured this cycle — evaluators should test with actual SIFT tools before accepting for production"
    - Point to Phase 3 handover-r4.md risk #9: "Lite scanners may diverge from production code path"
  - [ ] Write "Real-Sample Integration Guide" subsection (evaluator methodology)
    - How to source SANS forensic dataset (3+ cases with ground-truth labels)
    - How to wire real tools into accuracy.py (replace lite_* functions with MCP tool calls)
    - Expected real-tool recall measurement process
    - Definition of acceptable recall threshold (recommend ≥ 0.80 per spec.md §8 D-7)

- **Validation**: `pytest -xvs tests/test_accuracy_report.py::test_m3_divergence_section`
  - ✅ File contains "lite" AND "real" (both present)
  - ✅ Section title includes "divergence" or "path"
  - ✅ Explains why real samples unavailable (SANS constraint, cost, time)
  - ✅ Provides integrator methodology (at least 5 steps to add real tools)

- **Risk**: MEDIUM (honesty + clarity required; misstatement could mislead evaluators)

### M4: Evaluator Methodology & Integration Guide
**Goal**: Enable future phases or external reviewers to measure real-tool recall.

- **Tasks**:
  - [ ] Write "Evaluator Methodology Guide" section
    - Prerequisites: SIFT environment or Docker proxy (see architecture.md)
    - Step 1: Procure ≥3 forensic datasets with ground-truth labels (SANS, ICS-CERT, or internal)
    - Step 2: Structure case dirs matching case-001 layout (evidence/, manifest.json)
    - Step 3: Replace lite_* scanner functions in accuracy.py with actual MCP tool calls
    - Step 4: Run `make accuracy` with real cases
    - Step 5: Measure recall per tool and aggregate
  - [ ] Write "Metrics Definition" subsection
    - Recall = TP / (TP + FN)
    - Precision = TP / (TP + FP)
    - F1 = 2 * (Precision * Recall) / (Precision + Recall)
    - Why recall ≥ 0.80 matters: no false negatives in incident response
  - [ ] Write "Fixture Structure Reference" subsection
    - Link to case-001 as template: `repos/find-evil-fixtures/cases/case-001/`
    - JSON schema for manifest.json (expected_findings dict per tool)
    - Evidence files layout (evidence/*, manifest.json, etc.)

- **Validation**: `pytest -xvs tests/test_accuracy_report.py::test_m4_evaluator_guide`
  - ✅ File mentions "accuracy.py" explicitly
  - ✅ Integration steps documented (4+)
  - ✅ Metrics (recall, precision, F1) defined
  - ✅ Reference to case-001 structure provided

- **Risk**: LOW (procedural documentation)

### M5: Assembly, Keywords, & Verification
**Goal**: Integrate all sections, verify keywords, validate reproducibility, pass lint.

- **Tasks**:
  - [ ] Merge all M1–M4 sections into single markdown document
  - [ ] Add "Risk & Limitations" section
    - Synthetic-only recall does NOT validate production tool accuracy
    - Lite-mode suitable for architecture review; not for production IR
    - Phase 3 kill-switch depends on Go bridge; Phase 4 report uses Python harness
    - Recommend real-tool validation gate before any production deployment
  - [ ] Verify all 7 keywords present in final doc:
    1. "recall" (✅)
    2. "synthetic" (✅)
    3. "case-001" (✅)
    4. "lite-vs-real" or "divergence" (✅)
    5. "13 ground-truth findings" or similar (✅)
    6. "accuracy.py" (✅)
    7. Evaluator methodology section (✅)
  - [ ] Run reproducibility check:
    ```bash
    cd /home/jang/Workspace/warvis-findEvil
    python harness/find-evil/scripts/accuracy.py --case repos/find-evil-fixtures/cases/case-001
    ```
    Expect: "recall: 100.00%"
  - [ ] Run linting (no code to lint, but check markdown formatting):
    ```bash
    ruff check docs/find-evil/accuracy-report.md 2>&1 | grep -c "error" || echo "0"
    ```
  - [ ] Run pytest validation suite:
    ```bash
    pytest -xvs tests/test_accuracy_report.py
    ```

- **Validation**: `pytest -xvs tests/test_accuracy_report.py::test_m5_final_assembly`
  - ✅ File size 1500–3000 words
  - ✅ All 7 keywords present (grep count ≥ 7 occurrences total)
  - ✅ Reproducibility command succeeds
  - ✅ No ruff errors
  - ✅ All tests pass

- **Risk**: LOW (integration + validation)

## Stop-and-Fix Rule

**Milestone validation failure → next milestone blocked until fix applied.**

Example: If M2 synthetic section test fails (reproduction command doesn't produce 100% recall), do NOT proceed to M3. Instead:
1. Debug accuracy.py output
2. Fix case-001 fixture or accuracy.py logic (if applicable)
3. Re-run M2 validation test
4. Only then proceed to M3

## Done When

- [x] Milestone M1: Research complete (fixture analyzed, baseline measured)
- [x] Milestone M2: Synthetic path documented with reproduction
- [x] Milestone M3: Lite-vs-real divergence documented honestly
- [x] Milestone M4: Evaluator methodology guide provided
- [x] Milestone M5: Keywords verified, tests pass, reproducible

### Final Verification Checklist

- [x] `docs/find-evil/accuracy-report.md` exists (1537 words, 6.8 KB)
- [x] All 7 keywords present: recall (21×), synthetic (13×), case-001 (15×), 13 (14×), accuracy.py (6×), lite (17×), real (30×)
- [x] `python harness/find-evil/scripts/accuracy.py --case harness/find-evil/fixtures/cases/case-001` returns 0 (100% recall)
- [x] `pytest -xvs tests/test_accuracy_report.py` passes (all 8 tests)
- [x] No ruff lint errors (tests/test_accuracy_report.py clean)
- [x] Document reviewed for honesty (lite-vs-real divergence is clear, comprehensive, candid)
- [x] Ready for evaluator review

## Status Log

| Date | Event | Owner |
|------|-------|-------|
| 2026-05-07 | UoW accuracy-report assigned to warvis-planner | warvis-initiator |
| 2026-05-07 | Milestone plan created | warvis-planner |
| 2026-05-07 | TDD cycle (M1–M5): RED→GREEN→REFACTOR complete | warvis-maker |
| 2026-05-07 | Tests passing (8/8), reproduction verified, keywords confirmed | warvis-maker |
| TBD | Final verification | warvis-verifier |
| TBD | Session close + lesson capture | warvis-finisher |

## Risk Assessment

| # | Risk | Severity | Mitigation |
|---|------|----------|-----------|
| 1 | Case-001 fixture structure unclear | LOW | M1 validates manifest.json parsing |
| 2 | accuracy.py behavior unexpected | LOW | Run baseline before writing docs |
| 3 | "Lite-vs-real" section unclear or misleading | **MEDIUM** | Multiple review passes; cross-check against handover-r4.md risk #9 |
| 4 | Evaluator methodology incomplete | LOW | Use case-001 as reference template; document step-by-step |
| 5 | Keywords missing from final doc | LOW | Grep validation before final commit |
| 6 | Reproducibility command fails | LOW | Test before doc publication |

**Overall Risk Level**: LOW (mostly documentation; single critical section on divergence honesty)

## Appetite & Time Budget

- **Appetite**: 2 days (document-focused, no code implementation)
- **Breakdown**:
  - M1: 2 hours (research + baseline)
  - M2: 3 hours (synthetic section writing)
  - M3: 2 hours (divergence — most careful section)
  - M4: 2 hours (evaluator guide)
  - M5: 1 hour (assembly + verification)
  - **Total**: ~10 hours (fits in 2-day appetite)

## Notes

1. **Honesty is critical**: This is a Phase 4 documentation artifact. Evaluators will read the lite-vs-real divergence section. If it's not clear that we're measuring synthetic-only recall, we risk credibility damage.

2. **Phase 3 handover context**: Go bridge kill-switch deadline is 2026-05-09. This accuracy-report UoW is Phase 4 (post-Phase-3). If Go bridge reverts (kill-switch fails), report should note Python-only orchestration path.

3. **Future integration**: Real-sample recall measurement (≥0.80 gate) is planned for Phase 4 or later, pending SANS dataset procurement. This report documents the **why** and **how** for evaluators.

4. **No code changes to Phase 1+2**: Phase 1+2 MCP server is locked. All work is documentation + test harness validation.

---

**Status**: shipped (2026-05-07, all milestones complete, verify=PASS, lesson_id=08edc27e-7d16-451b-8e5b-4179bf7fd1a8)
**Risk Level**: LOW (M3 honesty=MED, mitigated)
