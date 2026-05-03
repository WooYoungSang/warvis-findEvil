# Harness Config: ITEM-212-day3 — Kill Switch Test 1 (INITIALIZE→TRACE)

**UoW**: ITEM-212-FIND-EVIL Day 3  
**Date**: 2026-05-04  
**Scope**: Full INITIALIZE→TRACE flow with real MCP tool call (M3a–M3e)

---

## P1: Always-on Layers

### CLAUDE.md Evaluation
- **Quality**: `map` (83 lines, concise, clear structure)
- **Format**: Well-organized project instructions + agent delegation table
- **Verdict**: ✓ Not encyclopedia; no refactoring needed

### Local Agents Available
- `warvis-orchestrator.md` — THIS AGENT (lifecycle)
- `warvis-maker.md` — TDD implementation
- `warvis-verifier.md` — quality gates
- `tdd-red.md`, `tdd-green.md`, `tdd-refactor.md` — TDD cycle coordination

### Verdict
✓ Harness is stable. Day 2 FSM work complete (34/34 tests pass).

---

## P2: On-demand Layers

### Skills Reuse Assessment
- **UoW builds on Day 2 FSM** → Reuse existing state types + transitions
- **New work**: MCP integration (case.open call), CLI handler, end-to-end flow
- **TDD pattern**: Red→Green→Refactor for each milestone (M3a–M3e)

### Delegation Decision
- **Implementation**: `direct` via TDD (red→green→refactor per milestone)
- **MCP integration**: Direct (no external SDK, stdlib JSON-RPC only)
- **Testing**: Direct (mock MCP server for unit tests; real server for integration)
- **Security**: Gate 3 pattern scan (no shell injection in MCP calls)

---

## P3: Deterministic Layers

### Test Commands (from Day 2 Harness)
- **Go test**: `cd /home/jang/Workspace/warvis-findEvil/warvis && go test ./...`
- **Go build**: `cd /home/jang/Workspace/warvis-findEvil/warvis && go build -o bin/warvis ./cmd/warvis`
- **Go lint**: `cd /home/jang/Workspace/warvis-findEvil/warvis && golangci-lint run`

### Gate Strategy (Day 3)
1. **Lint** → `golangci-lint run ./...`
2. **Unit test** → `go test ./internal/hunt/... -short` (FSM + MCP integration)
3. **Build check** → `go build -o bin/warvis ./cmd/warvis`
4. **Integration test** → `warvis hunt /evidence/test.img` (real MCP server spawn + case.open)
5. **Kill Switch Test 1** → Verify state.json TRACE + audit.jsonl 2+ lines

### Eval Harness
- **Baseline**: Day 2 tests (34/34 pass)
- **Day 3 target**: 40+ tests (M3a–M3e coverage)
- **Capture**: After each milestone, run `go test ./internal/hunt/... -cover`

---

## P4: Measurable Layers

### Observability
- **Audit log**: Append-only `/cases/<uuid>/audit.jsonl` (JSON-Lines)
- **State persistence**: `/cases/<uuid>/state.json` (JSON)
- **Logging**: Structured errors via fmt.Errorf (no external logging yet)

### Baseline
- **Input**: Day 2 FSM (34/34 tests pass, state types defined, transitions working)
- **Output**: Kill Switch Test 1 PASSES
  - `warvis hunt /evidence/test.img` exits 0
  - `/cases/<uuid>/state.json` contains `"current_state": "TRACE"`
  - `/cases/<uuid>/audit.jsonl` has 2+ lines (case_opened, state_transition)

---

## P5: Harness Config Decision

| Layer | Decision | Rationale |
|-------|----------|-----------|
| **Impl** | TDD direct | M3a–M3e are small, focused milestones; TDD cycle per milestone |
| **Test** | Direct | `go test ./internal/hunt/...` is self-contained |
| **Lint** | Direct | golangci-lint single command |
| **Build** | Direct | `go build` must succeed |
| **Integration** | Direct | MCP client exists (Day 1), spawn real Python server for M3c test |
| **Reused skills** | tdd-red, tdd-green, tdd-refactor | Coordinate TDD per milestone |
| **New skills** | None | Kill-switch work not repeatable; no skill extraction |

---

## Collaboration Map

### Impl Lane (TDD per Milestone)
- **M3a (INITIALIZE handler)**:
  - Red: Test that FSM.Transition("case.open called") works on INITIALIZE state
  - Green: Implement InitializeState handler in FSM (call MCP case.open, store case_id)
  - Refactor: Clean up MCP client integration
  - Acceptance: `go test -run TestInitializeTransition` passes

- **M3b (TRACE entry)**:
  - Red: Test that TRACE state has allowed tools (timeline.build, log.query)
  - Green: Implement TraceState with correct tool list
  - Refactor: N/A (simple state type)
  - Acceptance: `go test -run TestTraceAllowedTools` passes

- **M3c (End-to-end CLI)**:
  - Red: Test that `warvis hunt <evidence>` spawns MCP server
  - Green: Implement main.go Hunt command (parse evidence path, spawn server, call case.open)
  - Refactor: Clean up error handling
  - Acceptance: `warvis hunt /evidence/test.img` exits 0, creates state.json

- **M3d (Audit logging)**:
  - Red: Test that AuditLog.Append() adds events to audit.jsonl
  - Green: Verify AuditLog.Append() called during case_open + state_transition
  - Refactor: N/A (audit.go already complete from Day 2)
  - Acceptance: `go test -run TestAuditLog` passes

- **M3e (Kill Switch Test 1)**:
  - Red: Test harness target `make kill-switch-check-1`
  - Green: Implement Makefile target (run warvis hunt, verify outputs)
  - Refactor: N/A (Makefile is declarative)
  - Acceptance: `make kill-switch-check-1` exits 0

### Test Lane
- Unit: `internal/hunt/fsm_test.go` (INITIALIZE handler)
- Unit: `cmd/warvis/main_test.go` (Hunt command)
- Integration: End-to-end `warvis hunt /evidence/test.img` with real MCP server

### Gate Lane
- **Lint**: No shell execution, no secrets in MCP calls
- **Test**: 90%+ coverage on M3a–M3e code paths
- **Build**: `go build` with no warnings
- **Integration**: `warvis hunt` completes without panic, generates state.json + audit.jsonl

---

## HITL Points

| Risk Level | Task | Approval |
|-----------|------|----------|
| **MEDIUM** | MCP case.open call integration | Auto-proceed; mock in unit tests, real server in integration |
| **MEDIUM** | CLI evidence path validation | Auto-proceed; simple FileExists check |
| **LOW** | Audit log append | Auto-proceed; already tested in Day 2 |
| **NONE** | No destructive operations | N/A |

---

## Dead Weight Check

### CLAUDE.md Sections — Still Needed?
- All sections still relevant (no dead weight detected)

### Kill Switch Constraint
- "Go bridge must be demo-stable by 2026-05-09" → Day 3 is Milestone 1
- "DO NOT BREAK src/find_evil_mcp/" → Enforced (no Python imports)

---

## Milestones & Acceptance Criteria

### M3a — INITIALIZE handler (1.5h)
- **Files**: `internal/hunt/fsm.go`, `internal/hunt/fsm_test.go`
- **Acceptance**:
  - [ ] `HuntFSM.Transition("case.open")` moves INITIALIZE → TRACE
  - [ ] Case ID is UUID v4
  - [ ] `go test -run TestInitializeTransition` passes
  - [ ] No panics on invalid state

### M3b — TRACE entry (0.5h)
- **Files**: `internal/hunt/state.go` (already done, verify)
- **Acceptance**:
  - [ ] TraceState.AllowedTools() == ["timeline.build", "log.query"]
  - [ ] `go test -run TestTraceAllowedTools` passes

### M3c — End-to-end CLI (2h)
- **Files**: `cmd/warvis/main.go`, `cmd/warvis/main_test.go` (new)
- **Acceptance**:
  - [ ] `warvis hunt /evidence/test.img` exits 0
  - [ ] Spawns Python MCP server successfully
  - [ ] Calls case.open via JSON-RPC
  - [ ] Creates `/cases/<uuid>/state.json`
  - [ ] `go test -run TestHuntCommand` passes

### M3d — Audit logging (0.5h)
- **Files**: `internal/hunt/audit.go` (already done, verify)
- **Acceptance**:
  - [ ] AuditLog entries written to `/cases/<uuid>/audit.jsonl`
  - [ ] 2+ lines (case_opened, state_transition)
  - [ ] `jq . /cases/<uuid>/audit.jsonl` parses all lines
  - [ ] `go test -run TestAuditLog` passes

### M3e — Kill Switch Test 1 (1h)
- **Files**: `harness/find-evil/Makefile`, `harness/find-evil/test-fixtures/`
- **Acceptance**:
  - [ ] `make kill-switch-check-1` target exists
  - [ ] Runs `warvis hunt /evidence/test.img`
  - [ ] Verifies state.json contains `"current_state": "TRACE"`
  - [ ] Verifies audit.jsonl is valid JSONL
  - [ ] Exits 0 on success, 1 on failure

---

## Success Criteria (Definition of Done)

- [ ] `go test ./internal/hunt/...` — 40+ tests pass
- [ ] `go test ./cmd/warvis/...` — M3c command test passes
- [ ] `go build -o bin/warvis ./cmd/warvis` — Succeeds
- [ ] `warvis hunt /evidence/test.img` — Creates state.json + audit.jsonl
- [ ] `/cases/<uuid>/state.json` — Contains `"current_state": "TRACE"`
- [ ] `/cases/<uuid>/audit.jsonl` — 2+ valid JSON-Lines entries
- [ ] `make kill-switch-check-1` — Exits 0 (all checks pass)
- [ ] No Python files touched; MCP server stable
- [ ] Audit trail shows case_open + state_transition events

---

## Resource Constraints

- **Max budget**: 5.5 hours (1 day blitz)
- **Context**: Day 2 FSM complete, MCP client functional (from Day 1)
- **Kill switch**: Test 1 must pass by EOD 2026-05-04
- **Dependency**: Python MCP server running (already set up in Day 2)

---

## TDD Coordination

**Per milestone**:
1. Write one failing test (Red)
2. Implement minimum code to pass (Green)
3. Refactor if needed (Refactor)
4. Move to next milestone

**Verification after all milestones**:
- `go test ./...` all pass
- `go build ./cmd/warvis` succeeds
- `make kill-switch-check-1` passes

