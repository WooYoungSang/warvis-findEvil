# Plan: valhuntir-analysis — Precision Differentiation Matrix vs Valhuntir

## Context
- **project_id**: warvis-findEval
- **dev_session_id**: 4fc5f606-d7ca-47a3-8d3d-3c1f5447d7f1
- **uow_id**: valhuntir-analysis
- **tech_stack**: Python 3.10+ (Markdown analysis, testing via pytest + ruff)
- **ssot_source**: Prior WebFetch (Valhuntir facts) + project CLAUDE.md + internal architecture docs
- **output**: docs/find-evil/valhuntir-comparison.md

## Scope

### To Generate
- Structured differentiation matrix (6 SANS criteria, equal 16.7% weight each)
- Executive summary (1-line gap + 3 narrow axes)
- Architecture side-by-side comparison (Valhuntir vs warvis-findEval)
- "What Valhuntir Doesn't Have" section (≥3 of 7 candidates, with honesty)
- Honest positioning statement for devpost-page (single paragraph, no overclaiming)
- 38-day action items (5 concrete tasks)

### No-Go (Scoped Out)
- Re-fetching Valhuntir GitHub (info already extracted in prior session)
- Modifying warvis/ or src/find_evil_mcp/ code
- Creating new Python tools or integration code
- Changing devpost-page.md (this UoW just generates analysis; a later UoW will rewrite it)

## Milestones

### M1: Data Extraction & Honesty Calibration (TDD: Test-First)

**Purpose**: Extract facts about both systems and establish a "no overclaim" checklist.

**Tasks**:
- [ ] TDD-Red: Write test_valhuntir_comparison.py with 5 core assertions:
  - File exists: `docs/find-evil/valhuntir-comparison.md`
  - "Valhuntir" appears ≥5 times
  - "criterion" or "criteria" appears ≥6 times
  - "differentiation" or "advantage" present
  - ≥3 markdown tables
  - "honest" or "we don't claim" or "we focus" phrase present (honesty gate)
- [ ] Read accuracy-report.md §4 (lite-vs-real divergence) to establish honesty pattern
- [ ] Create a checklist: What claims we CAN make (kill-switch 5/5 PASS, FSM architecture, audit immutability) vs. what we CANNOT (>80% real-sample accuracy, production-ready DFIR)
- [ ] Document 7 candidates in working notes:
  1. offline-only (no API key required)
  2. single Go binary deployment
  3. FSM-deterministic state machine
  4. budget-preserving resume
  5. kill-switch test harness
  6. sub-500ms cold-start
  7. hardcoded tool whitelist (vs gateway-level discipline)

**Validation**: `pytest test_valhuntir_comparison.py::test_structure -v`

**Risk**: LOW — data extraction is deterministic, no new facts needed from web.

### M2: Architecture Comparison Table (TDD: Green)

**Purpose**: Build side-by-side architecture table comparing both systems.

**Tasks**:
- [ ] TDD-Green: Write minimal valhuntir-comparison.md with:
  - Executive summary stub (1 line + 3 axes)
  - Architecture comparison table (≥5 rows × 3 columns)
  - Pass test assertions for table count, vocabulary

**Validation**: `pytest test_valhuntir_comparison.py::test_architecture_table -v`

**Risk**: LOW — table is declarative, based on known facts.

### M3: 6-Criteria Comparison (TDD: Green + Honesty Check)

**Purpose**: Build equal-weight comparison across all 6 SANS judging criteria.

**Tasks**:
- [ ] TDD-Green: Add "6-Criteria Comparison" table with:
  - All 6 criteria (16.7% each): Autonomous Execution Quality, IR Accuracy, Breadth & Depth, Constraint Implementation, Audit Trail Quality, Usability & Documentation
  - Columns: Criterion | Valhuntir Strength | warvis-findEval Strength | Winner | Why
  - Row notes ensure "honest" phrase appears (e.g., "We don't claim production accuracy without real samples")
- [ ] Ensure ≥6 "criterion" mentions across file

**Validation**: `pytest test_valhuntir_comparison.py::test_six_criteria -v`

**Risk**: MEDIUM — honesty check is critical. Must not overclaim on IR Accuracy.

### M4: Differentiation Axes & "What We Don't Have" (TDD: Honesty First)

**Purpose**: Identify 3–5 true differentiation points where we can credibly claim advantage + be honest about gaps.

**Tasks**:
- [ ] TDD-Red: Define "honest differentiation" checklist:
  - Can only claim "advantage" if:
    1. Kill-switch test (5/5) proves it OR
    2. Architecture doc explicitly states it OR
    3. Accuracy report explicitly measures it (or caveats why not)
  - Cannot claim: "higher accuracy," "more tools," "better IR outcomes" without validation
- [ ] Evaluate 7 candidates and SELECT 3 strongest:
  - Single Go binary deployment (✓ keep)
  - FSM-deterministic state machine (✓ keep)
  - Budget-preserving resume (✓ keep)
  - Bonus: Kill-switch test harness (5/5 PASS)
  - Bonus: Hardcoded per-state tool whitelist
- [ ] Add section: "What Valhuntir Doesn't Have" with ≥3 items
- [ ] Ensure "differentiation" or "advantage" keyword appears

**Validation**: `pytest test_valhuntir_comparison.py::test_honest_positioning -v`

**Risk**: MEDIUM-HIGH — honesty calibration is subjective.

### M5: Honest Positioning Statement + Action Items (TDD: Final Integration)

**Purpose**: Draft single paragraph for devpost-page + 38-day roadmap.

**Tasks**:
- [ ] TDD-Green: Add "Honest Positioning Statement" section (~200 words):
  - Acknowledge Valhuntir's breadth and proven accuracy
  - Clearly state our trade-offs (architectural simplicity, deployability, resume)
  - Use "we don't claim," "we focus on," or "honest" phrasing
- [ ] Add "38-Day Action Items" section (5 concrete tasks with dates/specificity)
- [ ] Pass all test assertions

**Validation**: `pytest test_valhuntir_comparison.py -v && ruff check docs/find-evil/valhuntir-comparison.md`

**Risk**: LOW — final synthesis of prior milestones.

## Stop-and-Fix Rule

**Milestone validation failure → do not advance to next milestone.** Example:
- M2 table count <3 → fix M2, re-validate, then proceed to M3
- M3 fails "honest" phrase check → re-read accuracy-report.md §4, revise M3 positioning, re-test
- M4 claim cannot be substantiated → remove claim, select different candidate, re-test

## Done When

- [x] test_valhuntir_comparison.py exists with assertions
- [x] docs/find-evil/valhuntir-comparison.md exists with all 6 sections
- [x] `pytest test_valhuntir_comparison.py -v` → ALL PASS
- [x] `ruff check docs/find-evil/valhuntir-comparison.md` → 0 errors
- [x] "Valhuntir" appears ≥5 times
- [x] "criterion" appears ≥6 times
- [x] "differentiation" or "advantage" present
- [x] ≥3 markdown tables
- [x] Honesty phrase present
- [x] devos_verify_dev_session → PASS

## Status

| M | Name | Status | Completed | Evidence |
|---|------|--------|-----------|----------|
| M1 | Data Extraction & Honesty Calibration | Ready | — | test_valhuntir_comparison.py draft |
| M2 | Architecture Comparison Table | Blocked on M1 | — | — |
| M3 | 6-Criteria Comparison | Blocked on M2 | — | — |
| M4 | Differentiation Axes | Blocked on M3 | — | — |
| M5 | Honest Positioning + Action Items | Blocked on M4 | — | — |

---

## Verification Strategy

### Per-Milestone Validation
- M1: `pytest test_valhuntir_comparison.py::test_structure -v`
- M2: `pytest test_valhuntir_comparison.py::test_architecture_table -v`
- M3: `pytest test_valhuntir_comparison.py::test_six_criteria -v`
- M4: `pytest test_valhuntir_comparison.py::test_honest_positioning -v`
- M5: `pytest test_valhuntir_comparison.py -v && ruff check docs/find-evil/valhuntir-comparison.md`

### Final Gate (Before devos_verify_dev_session)
```bash
pytest test_valhuntir_comparison.py -v
```

**Expected**: All assertions PASS, exit code 0.

---

## Risk Assessment

**Overall Risk Level**: **MEDIUM**

- **M1-M2 (LOW)**: Data extraction + table building are deterministic
- **M3 (MEDIUM)**: Honesty calibration requires careful reading of accuracy-report.md §4
- **M4 (MEDIUM-HIGH)**: Differentiation claim validation is subjective; use kill-switch docs + architecture.md only
- **M5 (LOW)**: Final synthesis, no new facts needed

**Mitigation**:
1. Use accuracy-report.md "Lite-vs-Real Path Divergence" pattern for honesty checks
2. Only claim advantages verified by kill-switch tests (5/5 PASS) or explicit architecture statements
3. Match honesty tone from accuracy-report.md
4. Leave "action items" as forward-looking (don't claim they're done)

---

## Notes for Agent

- **No web re-fetch**: Valhuntir facts are fixed (prior WebFetch)
- **No code changes**: Warvis and Python MCP are locked; analysis-only
- **Honesty is load-bearing**: Overclaiming will be discounted by evaluators. Match accuracy-report tone.
- **Time budget**: 3–4 hours for full 5-milestone cycle
- **Handoff**: Output feeds into later UoW that rewrites devpost-page.md + README.md

---

**Created**: 2026-05-08  
**Target Completion**: 2026-05-14

---

**Status**: shipped (2026-05-08, 16/16 pytest PASS, ruff clean, binary size 7.8 MB verified, lesson_id=fc0748f3-5f92-4d8e-927e-6a35aad18b7c)
**Risk Level**: MEDIUM (honesty calibration — mitigated, factual verification gate passed)
