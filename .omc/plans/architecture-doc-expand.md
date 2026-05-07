# Plan: architecture-doc-expand — Expand docs/find-evil/architecture.md for Phase 3+4

## Context
- **project_id**: warvis-findEval
- **dev_session_id**: b7adc834-b099-4404-ae89-0f8430be5cfd
- **uow_id**: architecture-doc-expand
- **target_file**: /home/jang/Workspace/warvis-findEvil/docs/find-evil/architecture.md
- **current_lines**: 342 (Phase 2 subprocess model)
- **target_lines**: ≥600 (delta ~258 lines)
- **tech_stack**: Markdown, Python (schema validation), Go (FSM reference)

## Scope

### Must ADD (do not delete Phase 2 sections):
1. System Overview diagram — Python MCP + Go Bridge integration (high-level flow)
2. Hunt FSM State Machine section — 5 states + allowed tools per state + budget model
3. Agent Loop section — Gemma 4 via Ollama (gemma4:26b-a4b-it-q4_K_M), JSON action parsing
4. MCP Stdio Transport section — JSON-RPC 2.0 subprocess lifecycle, error handling
5. Audit & Persistence section — audit.jsonl format with examples, state.json schema, resume
6. Security Boundaries (Go bridge) — sandbox_root, prompt-injection defense, tool whitelist enforcement
7. Phase 4 Submission Readiness — kill-switch 5/5 results, artifact checklist

### Must NOT:
- Modify warvis-go-architecture.md (1687 lines, separate doc — HANDS OFF)
- Delete existing Phase 2 content (sections 1-10 in architecture.md must remain)
- Modify src/find_evil_mcp/ or warvis/ code

## Verification Strategy

**Final Checks** (automated via tests/test_architecture_doc.py):
- File length ≥ 600 lines
- Keywords present: "INITIALIZE", "TRACE", "SCAN", "EXPOSE", "LOCK", "JSON-RPC", "Gemma", "audit.jsonl", "MCP"
- ≥ 2 diagrams (ASCII art or mermaid blocks in new sections)
- All section headings match required list above
- No Phase 2 content deleted (sections 1-10 still in output)
- Markdown passes basic validation (no broken syntax)

**Test Command:**
```bash
python -m pytest tests/test_architecture_doc.py -v
```

## Milestones

### M1: System Overview & Integration Diagram
**Target**: Add 40-50 lines with mermaid diagram showing Python + Go integration

**TDD Cycle**:
1. TEST: Write `test_system_overview_section_exists()` in tests/test_architecture_doc.py
2. RED: Verify test fails (section not in file)
3. GREEN: Add "## 11. Phase 3 System Overview — Python MCP + Go Bridge Integration" section with:
   - Mermaid flowchart showing: CLI warvis → Hunt FSM (Go) → MCP Client (Go) → MCP Server (Python)
   - Flow arrows: stdin/stdout JSON-RPC, state persistence `/cases/<uuid>/`
   - Tool-call roundtrip flow
   - ≥40 new lines
4. REFACTOR: Polish diagram labels, ensure clarity

**Validation**: `wc -l docs/find-evil/architecture.md` (should be ≥380)

**Risk**: LOW (doc-only, no code)

---

### M2: Hunt FSM State Machine Section
**Target**: Add 80-100 lines detailing all 5 states + tool whitelists + budget model

**TDD Cycle**:
1. TEST: Write `test_fsm_section_all_states()` checking for all 5 state names
2. RED: Verify test fails
3. GREEN: Add "## 12. Hunt FSM State Machine" section with:
   - Table: State name | Allowed Tools | LLM Autonomy % | Purpose
   - Details for each: INITIALIZE (case.open), TRACE (timeline.build, log.query), SCAN (iocs.scan, memory.*, net.*), EXPOSE (verify.cross_check, report.append), LOCK (no tools)
   - Budget model: max 50 LLM turns per state, max 10 invalid JSON attempts, state timeout 600s
   - State transition ASCII diagram (boxes + arrows)
   - ≥60 new lines
4. REFACTOR: Add example state transition sequence

**Validation**: `grep -c "INITIALIZE\|TRACE\|SCAN\|EXPOSE\|LOCK" docs/find-evil/architecture.md` (should be ≥15)

**Risk**: MEDIUM (must match state.go exactly)

---

### M3: Agent Loop & Ollama Integration Section
**Target**: Add 60-80 lines on Gemma 4 loop, JSON parsing, retry/timeout

**TDD Cycle**:
1. TEST: Write `test_agent_loop_keywords()` checking for "Gemma" and "Ollama"
2. RED: Verify test fails
3. GREEN: Add "## 13. Agent Loop — Gemma 4 via Ollama" section with:
   - Model: gemma4:26b-a4b-it-q4_K_M at localhost:11434
   - System prompt template: current_state + available_tools + JSON format spec
   - Action struct: { action: "tool_call"|"transition"|"error", tool?: string, params?: {...}, reason?: string }
   - JSON parsing: Tier 3 (attempt JSON parse → on failure try regex extraction → on failure escalate)
   - Retry logic: max 3 invalid JSON attempts per loop iteration, then state transition failure
   - Budget: max 50 turns per state, timeout 600s per state, tool call timeout 120s
   - Example loop iteration pseudo-code (5-10 lines)
   - ≥50 new lines
4. REFACTOR: Add inline Go code snippet showing ParseAction function signature

**Validation**: `grep -i "gemma\|ollama" docs/find-evil/architecture.md | wc -l` (should be ≥8)

**Risk**: MEDIUM (model name + URL must match CLAUDE.md exactly)

---

### M4: MCP Stdio Transport & Audit Section
**Target**: Add 70-90 lines on JSON-RPC 2.0, subprocess lifecycle, audit.jsonl + state.json

**TDD Cycle**:
1. TEST: Write `test_mcp_stdio_keywords()` checking for "JSON-RPC" and "audit.jsonl"
2. RED: Verify test fails
3. GREEN: Add "## 14. MCP Stdio Transport & Audit Trail" section with:
   - JSON-RPC 2.0 message format (request/response/notification examples)
   - Subprocess lifecycle: warvis main spawns Python MCP server, maintains stdin/stdout pipes
   - Error handling: tool timeout → JSON error, Python exception → JSON error response
   - Audit trail: append-only audit.jsonl in `/cases/<uuid>/audit.jsonl`
   - Audit event schema: timestamp (ISO 8601), event_type, actor, case_id, details (JSON)
   - Example audit.jsonl with 4 events: case_opened, state_initialized, timeline_built, state_transitioned
   - State persistence: `/cases/<uuid>/state.json` with case_id, current_state, budgets, resume_flag
   - Resume mechanism: load state.json on warvis hunt --case-id <uuid>, restore budgets (NOT reset)
   - ≥60 new lines
4. REFACTOR: Add JSON schema blocks (minimal) for state.json and audit event

**Validation**: `grep -E "JSON-RPC|audit.jsonl|state.json" docs/find-evil/architecture.md | wc -l` (should be ≥12)

**Risk**: MEDIUM (must align with actual Go state.go structures)

---

### M5: Security Boundaries (Go Bridge) & Phase 4 Readiness
**Target**: Add 50-70 lines on Go-specific security + Phase 4 artifact checklist

**TDD Cycle**:
1. TEST: Write `test_security_and_phase4_sections()` checking presence of both sections
2. RED: Verify test fails
3. GREEN: Add "## 15. Security Boundaries (Go Bridge)" section with:
   - sandbox_root isolation: `/cases/<uuid>/` enforced at FSM level, no escape paths
   - Prompt injection defense: output sanitized (≤500 chars per tool result, wrapped with "untrusted" prefix)
   - Tool whitelist enforcement: agent only calls tools in state.AllowedTools()
   - Budget enforcement: prevent LLM token exhaustion, max 50 turns per state
   - Resume safety: budgets NOT reset on resume (critical invariant for kill-switch)
   - Then add "## 16. Phase 4 Submission Readiness" section with:
     - Kill-switch status: 5/5 tests PASS (M1-M7 complete per verify-phase3.md)
     - Artifact checklist: README.md, architecture.md (this), dataset.md, accuracy-report.md, devpost-page.md, demo-script.md, LICENSE, harness/find-evil/logs/
     - Go bridge stability: no revert needed if tests PASS by 2026-05-09
     - Phase 3 completion date: 2026-05-03 (M1-M7d complete)
   - ≥50 new lines
4. REFACTOR: Add links to verify-phase3.md and kill-switch status

**Validation**: `grep -E "sandbox_root|prompt.injection|kill.switch|Phase 4" docs/find-evil/architecture.md | wc -l` (should be ≥10)

**Risk**: LOW (status doc, references verify-phase3.md external source)

---

## Stop-and-Fix Rule

Each milestone has validation command. If validation FAILS:
- STOP before next milestone
- Fix section (add missing keywords, expand content, fix syntax)
- Re-run validation
- ONLY proceed to next milestone after PASS

## Done When

- [ ] Line count ≥ 600 (current 342 + ~258 from 5 milestones = ~600 total)
- [ ] All keywords present: INITIALIZE, TRACE, SCAN, EXPOSE, LOCK, JSON-RPC, Gemma, audit.jsonl, MCP, sandbox_root, kill-switch
- [ ] ≥ 2 diagrams (M1 system overview + M2 state machine)
- [ ] Phase 2 sections 1-10 intact (no deletions)
- [ ] `python -m pytest tests/test_architecture_doc.py -v` PASS
- [ ] No broken Markdown syntax
- [ ] devos_verify_dev_session PASS

## Status

| M | Task | State | Evidence |
|---|------|-------|----------|
| M1 | System Overview + diagram | COMPLETE | Section 11 (mermaid flowchart) added |
| M2 | FSM State Machine + tool whitelist | COMPLETE | Section 12 (5-state FSM, 12.1-12.5 subsections) added |
| M3 | Agent Loop (Gemma + Ollama) | COMPLETE | Section 13 (Gemma 4, Ollama, JSON parsing, retry logic) added |
| M4 | MCP Stdio Transport + audit.jsonl + state.json | COMPLETE | Sections 14-15 (JSON-RPC 2.0, subprocess lifecycle, audit.jsonl, state.json, resume) added |
| M5 | Security Boundaries + Phase 4 Checklist | COMPLETE | Sections 16-17 (sandbox isolation, prompt injection defense, kill-switch 5/5 status) added |

**Final Status: DONE**
- Lines: 342 → 650 (delta: +308)
- Sections added: 7 (11-17)
- Tests: 7/7 PASS
- Phase 2 sections (1-10): intact
- Markdown syntax: valid
- Keywords verified: INITIALIZE, TRACE, SCAN, EXPOSE, LOCK, JSON-RPC, Gemma, audit.jsonl, MCP
- Diagrams: 2 (mermaid flowchart + ASCII state machine)
- devos session: END (verified)

## Risk Assessment

**Overall**: **MEDIUM**

| Risk | Mitigation |
|------|-----------|
| Cross-reference accuracy (Go vs doc) | Copy directly from state.go, agent/loop.go; validate via keyword tests |
| Markdown syntax errors | Use `head -20` to spot-check; manual review before commit |
| Unintended Phase 2 deletion | Git status before commit; test suite validates sections 1-10 remain |
| Diagram clarity | ASCII art + mermaid; simple boxes + arrows |
| Wrong Ollama URL/model | Check CLAUDE.md section "kill-switch check": localhost:29134, gemma4:26b-a4b-it-q4_K_M |

---

**Status**: shipped (2026-05-07, 342→650 lines, 7/7 pytest PASS, ruff clean, lesson_id=b27fb4b5-1d66-460c-8bd9-2cdc98a662fc)
**Risk Level**: MEDIUM (mitigated — Phase 2 sections intact, all keywords verified)
