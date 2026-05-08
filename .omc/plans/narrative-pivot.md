# Plan: narrative-pivot — Rewrite README, devpost-page, CLAUDE.md to honest Valhuntir positioning

## Context
- **project_id**: warvis-findEval
- **dev_session_id**: 46e65694-3220-4fa2-8228-0cfaa2de50e9
- **tech_stack**: Python 3.10 (tests: pytest + ruff), Go 1.21+ (tests: go test ./...), Markdown
- **ssot_source**: valhuntir-comparison.md §5 (primary narrative source) + accuracy-report.md (honesty pattern reference)
- **deadline**: 2026-06-15 (SANS FIND EVIL Hackathon submission)

## Tagline Choice

**Selected**: Option #1 (RECOMMEND)

> "The smallest IR agent whose architecture — not its prompt — guarantees it cannot escape its forensic role"

**Why #1**:
- Precisely reflects §5 Honest Positioning Statement
- Emphasizes architectural differentiation vs. prompt discipline
- Distinct from Valhuntir's breadth narrative
- Supports narrow-but-defensible positioning on 3 differentiators: Go binary, 5-state FSM, resume+kill-switch

**Consistency target**: Must appear in README.md + devpost-page.md + CLAUDE.md (line 4 tagline area)

---

## Scope

### Files to Modify (3 total)

#### 1. README.md (first 100 lines)
- **Current**: Title + "Evil Has Nowhere to Hide" + generic MCP forensic IR narrative
- **Target**: Replace with honest positioning + new tagline + "single Go binary", "FSM", "resume" keywords
- **Action**: Rewrite Overview section (lines 11–22) + add "Why warvis" sub-section if absent
- **Key phrases to include**:
  - "smallest IR agent whose architecture guarantees"
  - "single Go binary"
  - "5-state FSM"
  - "resume" or "budget-preserving resume"
  - Mention Valhuntir (per §5 honesty pattern)

#### 2. docs/find-evil/devpost-page.md (narrative sections)
- **Current**: Generic "autonomous forensics" + generic "how we built it"
- **Target**: Honest positioning + Valhuntir comparison sub-section
- **Actions**:
  - Tagline: Replace "Evil Has Nowhere to Hide" with Option #1 tagline (line 5)
  - "What It Does" first paragraph: Emphasize FSM states + architecture-as-guarantee
  - "How We Built It" first paragraph: Mention "architecture guarantees" vs. "prompt discipline"
  - Add new sub-section "How We Compare to Valhuntir" (after "Accomplishments") — 2–3 paragraphs summarizing §2–§4 of valhuntir-comparison.md
  - Replace "What We Learned" honesty tone to match accuracy-report.md pattern

#### 3. CLAUDE.md (line 4 tagline)
- **Current**: "Evil Has Nowhere to Hide" (line 3)
- **Target**: Change to Option #1 tagline
- **Action**: Single-line replacement; preserve "Author", "Hackathon", "Deadline" metadata

### Files to NOT modify
- `docs/find-evil/warvis-go-architecture.md` — append-only history pattern
- `docs/find-evil/accuracy-report.md` — already locked (honesty pattern set)
- `docs/find-evil/demo-script.md` — regenerated only on narrative ship
- `src/find_evil_mcp/` — locked (Phase 1+2 complete)
- `warvis/` — locked (Phase 3 in progress)

---

## Milestones

### M1: CLAUDE.md tagline update (5 min, TDD: test first)
**Objective**: Replace line 3 with Option #1 tagline; verify via grep.

**Tasks**:
- [ ] **RED**: Write test that greps for old tagline ("Evil Has Nowhere to Hide") in CLAUDE.md — should FAIL
- [ ] **GREEN**: Edit CLAUDE.md line 3 to new tagline; test should PASS
- [ ] **VERIFY**: `grep "smallest IR agent whose architecture" CLAUDE.md` returns exact match

**Validation**:
```bash
grep "smallest IR agent whose architecture" /home/jang/Workspace/warvis-findEvil/CLAUDE.md
```
Expected: 1 match (line 3)

**Risk**: LOW

---

### M2: README.md title + tagline + overview rewrite (20 min, TDD cycle)
**Objective**: Rewrite README.md lines 1–22 to match honest positioning; include 3 keywords.

**Tasks**:
- [ ] **RED**: Write test `test_narrative_pivot.py::test_readme_contains_keywords()` that checks README.md for:
  - "single Go binary" (case-insensitive)
  - "FSM" or "finite state machine" (case-insensitive)
  - "resume" (case-insensitive)
  - Should FAIL on current README
- [ ] **RED**: Add test `test_readme_mentions_valhuntir()` — checks for "Valhuntir" in README.md, should FAIL
- [ ] **RED**: Add test `test_readme_honesty_phrase()` — checks for honesty pattern ("we don't claim", "we do not claim", or "we focus"), should FAIL
- [ ] **GREEN**: Edit README.md:
  - Line 1: Keep "# W.A.R.V.I.S — Find Evil"
  - Line 3: Replace tagline with Option #1
  - Lines 11–22 (Overview): Rewrite to emphasize 5-state FSM, Go binary, resume + narrow differentiation vs. Valhuntir
  - Add "Why warvis" sub-section (2–3 paragraphs) if absent, referencing Valhuntir as reference and claiming only 3 differentiators
- [ ] **VERIFY**: All 4 tests PASS

**Validation**:
```bash
python -m pytest test_narrative_pivot.py::test_readme_contains_keywords -v
python -m pytest test_narrative_pivot.py::test_readme_mentions_valhuntir -v
python -m pytest test_narrative_pivot.py::test_readme_honesty_phrase -v
```

**Risk**: MEDIUM (must preserve technical accuracy + 3 keyword coverage)

---

### M3: devpost-page.md tagline + narrative sections (25 min, TDD cycle)
**Objective**: Update devpost-page.md tagline, "What It Does", "How We Built It", add Valhuntir comparison sub-section.

**Tasks**:
- [ ] **RED**: Write test `test_devpost_tagline()` — checks for Option #1 tagline in devpost-page.md, should FAIL
- [ ] **RED**: Write test `test_devpost_valhuntir_section()` — checks for "How We Compare to Valhuntir" sub-section (heading), should FAIL
- [ ] **RED**: Write test `test_devpost_mentions_valhuntir()` — checks for "Valhuntir" keyword, should FAIL
- [ ] **GREEN**: Edit devpost-page.md:
  - Line 5: Replace "Evil Has Nowhere to Hide" with Option #1 tagline
  - "What It Does" (lines 17–33): Rewrite first paragraph to emphasize architecture-as-guarantee + FSM determinism (not prompt discipline)
  - "How We Built It" (lines 37–69): Rewrite first paragraph to reference "compile-time tool whitelist" + "architecture guarantees" language
  - After "Accomplishments" (line 107): Insert new sub-section with Valhuntir comparison (2–3 paragraphs)
- [ ] **VERIFY**: All 3 tests PASS

**Validation**:
```bash
python -m pytest test_narrative_pivot.py::test_devpost_tagline -v
python -m pytest test_narrative_pivot.py::test_devpost_valhuntir_section -v
python -m pytest test_narrative_pivot.py::test_devpost_mentions_valhuntir -v
```

**Risk**: MEDIUM-HIGH (narrative coherence + Valhuntir comparison accuracy + tone consistency)

---

### M4: Regression test suite (10 min, verification gate)
**Objective**: Confirm all 4 existing test suites still PASS after edits.

**Tasks**:
- [ ] **VERIFY**: `python -m pytest tests/test_submission_artifacts.py -v` — PASS
- [ ] **VERIFY**: `python -m pytest tests/test_accuracy_report.py -v` — PASS
- [ ] **VERIFY**: `python -m pytest tests/test_architecture_doc.py -v` — PASS
- [ ] **VERIFY**: `python -m pytest tests/test_valhuntir_comparison.py -v` — PASS
- [ ] **VERIFY**: `python -m pytest test_narrative_pivot.py -v` (all tests from M1–M3) — PASS

**Validation**:
```bash
python -m pytest tests/ test_narrative_pivot.py -v --tb=short
```

Expected: All 4 existing + narrative-pivot test suites PASS

**Risk**: LOW (existing tests should be unaffected by prose edits)

---

### M5: Lint + final verification (5 min, quality gate)
**Objective**: Confirm no syntax or formatting regressions in edited files.

**Tasks**:
- [ ] **VERIFY**: Manual spot-check: README.md + devpost-page.md + CLAUDE.md read cleanly (no broken links, no doubled taglines, no corrupted formatting)
- [ ] **VERIFY**: Tagline appears in exactly 3 places: README.md (line 3), devpost-page.md (line 5), CLAUDE.md (line 3)

**Validation**:
```bash
grep -n "smallest IR agent whose architecture" /home/jang/Workspace/warvis-findEvil/README.md /home/jang/Workspace/warvis-findEvil/docs/find-evil/devpost-page.md /home/jang/Workspace/warvis-findEvil/CLAUDE.md
```

Expected: 3 matches (one per file)

**Risk**: LOW

---

## Stop-and-Fix Rule

If any milestone validation fails:
1. **Do not proceed** to next milestone
2. **Debug & fix** the failing test/grep/assertion
3. **Re-run validation** until PASS
4. Only then advance to next milestone

Example: If M2 keyword test fails, edit README.md until all 3 keywords are present; re-validate before M3.

---

## Done When

- [x] All 5 milestones validated (M1–M5)
- [x] `test_narrative_pivot.py` — all tests PASS
- [x] Existing test suites (submission_artifacts, accuracy_report, architecture_doc, valhuntir_comparison) still PASS
- [x] Tagline appears exactly 3 times (README.md, devpost-page.md, CLAUDE.md)
- [x] README.md + devpost-page.md contain 3 keywords: "single Go binary", "FSM", "resume" (case-insensitive)
- [x] README.md + devpost-page.md mention "Valhuntir"
- [x] README.md + devpost-page.md contain honesty phrase
- [x] devpos-page.md includes "How We Compare to Valhuntir" sub-section (2–3 paragraphs)
- [x] No broken formatting or doubled content

---

## Status

| M | Name | Status | Evidence |
|---|------|--------|----------|
| M1 | CLAUDE.md tagline | pending | — |
| M2 | README.md rewrite | pending | — |
| M3 | devpost-page.md rewrite | pending | — |
| M4 | Regression tests | pending | — |
| M5 | Lint + final verification | pending | — |

---

## Risk Assessment

**Overall**: MEDIUM

**Breakdown**:
- **M1 (tagline replacement)**: LOW — single-line edit
- **M2 (README rewrite)**: MEDIUM — must balance technical accuracy, keyword coverage, Valhuntir mention, honesty tone
- **M3 (devpost-page rewrite)**: MEDIUM-HIGH — narrative coherence risk; Valhuntir comparison must be fair and accurate per §2–§4
- **M4 (regression)**: LOW — existing tests should not be affected by prose edits
- **M5 (lint)**: LOW — formatting check only

**Key risks**:
1. **Narrative incoherence**: Tagline must be consistent across 3 files and match §5 honest positioning
2. **Keyword coverage**: README.md must contain all 3 keywords ("single Go binary", "FSM", "resume")
3. **Valhuntir accuracy**: Comparison section must reflect valhuntir-comparison.md §2–§4 fairly; do not overstate our advantages
4. **Honesty tone**: Must match accuracy-report.md pattern ("we do not claim production-grade accuracy without real-sample measurement")

**Mitigation**:
- M2–M3 use TDD (RED→GREEN→VERIFY) to ensure test-driven accuracy
- M4 runs full regression suite before declaring completion
- M5 spot-checks all 3 files for tagline consistency and formatting integrity

---

## Notes

- **Test file location**: `/home/jang/Workspace/warvis-findEvil/test_narrative_pivot.py` (to be created during M1)
- **Test suite trigger**: `python -m pytest test_narrative_pivot.py -v`
- **Existing suites**: pytest auto-discovers tests in `tests/` directory
- **No code changes**: This UoW is prose-only (README, docs, CLAUDE.md tagline)
- **No Python/Go source modified**: src/find_evil_mcp/ and warvis/ remain locked
- **Harness policy**: Treat valhuntir-comparison.md §5 as SSOT for narrative; accuracy-report.md as honesty pattern reference

---

**Status**: shipped (2026-05-08, 47/47 tests PASS, ruff clean, zero regression, lesson_id=b099499b-d5d7-4d91-8b71-2d2742b45a40)
**Risk Level**: MEDIUM (mitigated — regression gate held)
