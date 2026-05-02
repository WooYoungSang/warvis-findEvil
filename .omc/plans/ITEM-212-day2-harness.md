# Harness Config: ITEM-212-day2 — Hunt FSM State Machine

**UoW**: ITEM-212-FIND-EVIL Day 2  
**Date**: 2026-05-02  
**Scope**: Internal Hunt FSM (5 state types, transitions, persistence, budgets)

---

## P1: Always-on Layers

### CLAUDE.md Evaluation
- **Quality**: `map` (82 lines, concise, clear structure)
- **Format**: Well-organized project instructions + agent delegation table
- **Verdict**: ✓ Not encyclopedia; no refactoring needed

### Local Agents
- **Count**: 6 agents in `.claude/agents/`
  - `warvis-orchestrator.md` — lifecycle orchestration (THIS AGENT)
  - `warvis-planner.md` — planning spec→plan
  - `warvis-maker.md` — TDD implementation
  - `warvis-verifier.md` — quality gates
  - `warvis-finisher.md` — session completion
  - `warvis-initiator.md` — startup checks

### Verdict
✓ Harness is well-structured. Agents live at `.claude/agents/` and are managed by devos_harness_propagate.

---

## P2: On-demand Layers

### Skills Available
- `.claude/commands/` — local command overrides
- `.claude/skills/` — domain-specific skills (warvis-specific)
- No root-level `skills/` (centralized via oh-my-claudecode global)

### Skills Reuse Assessment
- **UoW is purely Go domain** (no Python touches — locked)
- **Reusable patterns**:
  - `tdd-red` / `tdd-green` / `tdd-refactor` — Use for each FSM milestone
  - No warvis-specific skills needed for Day 2 (FSM is generic state machine logic)

### Delegation Decision
- **Implementation**: `direct` (warvis-maker can delegate, but orchestrator can do TDD coordination)
- **Test verification**: `direct` (small scope, no complex test harness)
- **Security**: Built-in gate 3 (no external depencies, stdio-only MCP)

---

## P3: Deterministic Layers

### Test Commands
- **Go**: `cd /home/jang/Workspace/warvis-findEvil/warvis && go test ./...`
- **Lint**: `cd /home/jang/Workspace/warvis-findEvil/warvis && golangci-lint run`
- **Build**: `cd /home/jang/Workspace/warvis-findEvil/warvis && go build -o bin/warvis ./cmd/warvis`

### Gate Strategy (Day 2)
1. **Lint** → `golangci-lint run`
2. **Test** → `go test ./internal/hunt/...`
3. **Security** — Pattern scan: no shell execution, no hardcoded secrets
4. **Integration** — Build succeeds: `go build ./...`

### Eval Harness
- **Exists**: No eval datasets yet (Phase 3 not reached)
- **Baseline**: None (Day 2 is structural, not behavior-driven)
- **Observe**: Capture `go test` pass rate after each milestone

---

## P4: Measurable Layers

### Observability
- **Logging**: Structured JSON in `internal/display/logger.go` (future)
- **Tracing**: Audit log format specified in architecture (`/cases/<uuid>/audit.jsonl`)
- **Metrics**: Test coverage via `go test -cover ./internal/hunt/...`

### Baseline
- **Day 1 status**: MCP client tests pass (transport + protocol)
- **Day 2 target**: 90%+ coverage on FSM transitions and state registry

---

## P5: Harness Config Decision

| Layer | Decision | Rationale |
|-------|----------|-----------|
| **Impl** | direct | Small scope, TDD-friendly, no complex domain knowledge needed |
| **Test** | direct | `go test ./internal/hunt/...` is self-contained, no integration dependencies |
| **Lint** | direct | golangci-lint is single command |
| **Security** | gate3-pattern-scan | No shell, no secrets; apply to all Go files |
| **Integration** | build-check | `go build ./...` must succeed; no external services |
| **Reused skills** | tdd-red, tdd-green, tdd-refactor | Use for milestone coordination |
| **New skills** | None | Day 2 work is not repeatable enough to warrant skill extraction |

---

## Collaboration Map

### Impl Lane (TDD)
- **Red**: Write failing test for missing FSM interface/method
- **Green**: Implement minimum code to pass test
- **Refactor**: Improve code quality while keeping tests green
- **Executor**: Direct (orchestrator does coordination)

### Test Lane
- **Unit tests**: `internal/hunt/fsm_test.go`, `state_test.go`, `transitions_test.go`, `registry_test.go`
- **Mock state transitions**: No real MCP needed for FSM logic tests
- **Persistence tests**: Create temp `/cases/<uuid>/` directories, verify JSON structure

### Gate Lane
- **Lint**: Enforce Go standard style + nil checks
- **Test**: Min 80% coverage on critical paths (FSM, transitions, registry)
- **Security**: No `exec.Command()`, no `os.Setenv()` in tests
- **Build**: `go build ./...` with no warnings

---

## HITL Points

| Risk Level | Task | Approval |
|-----------|------|----------|
| **MEDIUM** | FSM state transitions (5 types, 8 rules) | Auto-proceed; tests validate guards |
| **MEDIUM** | Persistence to `/cases/<uuid>/state.json` | Auto-proceed; use temp directories in tests |
| **LOW** | CLI flag parsing (`--phase`, `--case-id`) | Auto-proceed; simple string validation |
| **NONE** | No destructive operations (no Python changes, no prod deployments) | N/A |

---

## Dead Weight Check

### CLAUDE.md Sections — Still Needed?
- "Agent Delegation" — ✓ Used (warvis-maker for implementation)
- "Hunt Protocol (FSM)" — ✓ Used (defines state structure)
- "Harness Policy" — ✓ Used (no Python changes constraint)
- **NO dead weight detected**

### Kill Switch Constraint
- "Go bridge must be demo-stable by 2026-05-09" — ✓ Day 2 is part of timeline
- "DO NOT BREAK src/find_evil_mcp/" — ✓ Enforced (no Python imports, direct Go only)

---

## Timeline (Day 2: 2026-05-02)

### Milestone M2a: FSM skeleton + state types (2h)
- File: `warvis/internal/hunt/fsm.go`
- File: `warvis/internal/hunt/state.go`
- Test: `warvis/internal/hunt/fsm_test.go` (basic interface test)
- Acceptance: `go test -run TestFSMNew` passes

### Milestone M2b: Transition rules (2h)
- File: `warvis/internal/hunt/transitions.go`
- Test: `warvis/internal/hunt/transitions_test.go` (all 8 rules)
- Acceptance: All transition guards validated

### Milestone M2c: Case registry + persistence (2h)
- File: `warvis/internal/hunt/registry.go`
- Test: `warvis/internal/hunt/registry_test.go` (save/load roundtrip)
- Acceptance: JSON structure verified, resume preserves budgets

### Milestone M2d: CLI integration (1h)
- Update: `warvis/cmd/warvis/main.go` (add `--phase` flag)
- Test: Flag validation only (not end-to-end hunt)
- Acceptance: Flag parsing works

### Verification (30m)
- Build: `go build ./cmd/warvis`
- Test: `go test ./internal/hunt/...`
- Lint: `golangci-lint run ./internal/hunt/...`

---

## Success Criteria (Definition of Done)

- [ ] `go test ./internal/hunt/...` passes (all 4 test files)
- [ ] Case registry creates `/cases/<uuid>/state.json` with correct structure
- [ ] Resume with `--case-id <uuid> --phase TRACE` loads prior state + budgets (not reset)
- [ ] All 5 state types defined with required fields from architecture spec
- [ ] Audit log is append-only (never overwritten), JSON-Lines format
- [ ] `go build ./cmd/warvis` succeeds without warnings
- [ ] No Python files touched; no MCP server calls needed for Day 2 tests

---

## Resource Constraints

- **Max budget**: 7 hours (1 day blitz)
- **Context**: Architecture spec is complete (no design decisions needed)
- **Kill switch**: Demo-stable by 2026-05-09 (not Day 2 gate, but context)
- **Dependency**: Day 1 MCP client is complete (`go test ./internal/mcp/...` passes)

