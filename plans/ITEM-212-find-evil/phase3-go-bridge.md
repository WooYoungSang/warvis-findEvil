# W.A.R.V.I.S Go Bridge — Phase 3 Implementation Plan
## ITEM-212 FIND EVIL — 7-Day Hackathon Blitz (2026-05-02 ~ 2026-05-09)

**Status**: Planning Phase  
**Kill Switch Date**: 2026-05-09 23:59 KST (7 days)  
**Executor Target**: `/oh-my-claudecode:start-work phase3-go-bridge`

---

## 1. Kill Switch Criteria (Non-Negotiable)

All of these MUST pass by 2026-05-09 23:59 KST, or Go bridge abandoned + revert to Python stack:

```bash
# Test 1: INITIALIZE → TRACE state transition
warvis hunt /evidence/test.img
# Expected: FSM moves from INITIALIZE to TRACE, writes state.json

# Test 2: MCP tool invocation
# Expect: case.open + timeline.build called via JSON-RPC stdio
# Verify: /cases/<case_id>/audit.jsonl shows both tool_called events

# Test 3: Ollama Gemma 4 tool-call JSON parsing
# Expect: Gemma returns valid action JSON (call_tool | state_complete | escalate)
# Verify: No manual correction needed, 3+ tool invocations succeed

# Test 4: Structured JSONL logging
# Expect: /cases/<case_id>/audit.jsonl has proper JSON-Lines format
# Verify: jq . on audit.jsonl succeeds for all lines

# Test 5: warvis status <case_id> outputs current state
# Expect: Reads state.json, emits JSON with current FSM state
```

**Failure = Kill Bridge**: If any test fails, accept "Go bridge POC incomplete" and submit Python-only (accept lower placement).

---

## 2. Execution Timeline (7 days)

### Day 1 (2026-05-02, Thu) — MCP Client & Core Scaffolding
**Goal**: Minimal Go module with MCP client that talks to Python server.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M1a** — Go module init | `warvis/go.mod`, `warvis/go.sum` | `go mod tidy` succeeds |
| **M1b** — Package structure | `/cmd/warvis/main.go`, `/internal/mcp/`, `/internal/hunt/`, `/pkg/ollama/` | Directory tree matches architecture spec |
| **M1c** — MCP transport (stdio) | `internal/mcp/transport.go` + tests | Can spawn Python server, read/write JSON-RPC messages without errors |
| **M1d** — JSON-RPC protocol types | `internal/mcp/protocol.go` + `types.go` | `Request`, `Response`, `RPCError` marshal/unmarshal correctly |
| **M1e** — MCP client initialization | `internal/mcp/client.go` (Initialize + ListTools) | `mcp.New(ctx, "python -m find_evil_mcp.server")` returns client, `client.ListTools()` returns 9 tools |

**Definition of Done**: 
- `go build ./cmd/warvis` succeeds
- Unit tests for transport + protocol (mocking Python server) pass
- Python MCP server can be spawned, handshake succeeds

---

### Day 2 (2026-05-03, Fri) — Hunt FSM State Machine
**Goal**: Implement 5-state FSM with transitions, state persistence, budget tracking.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M2a** — Hunt FSM skeleton | `internal/hunt/fsm.go` | Compiles, exports `FSM` interface with 5 states |
| **M2b** — State types | `internal/hunt/state.go` (InitializeState, TraceState, etc.) | All 5 state structs defined with required fields |
| **M2c** — Transitions (typed) | `internal/hunt/transitions.go` | INITIALIZE→TRACE, TRACE→SCAN, SCAN→EXPOSE, EXPOSE→LOCK, LOCK→END all defined; guards validated |
| **M2d** — Case registry (JSON) | `internal/hunt/registry.go` | Persist FSM state to `/cases/<case_id>/state.json` after each transition |
| **M2e** — Budget counters | `internal/hunt/fsm.go` (persisted in state.json) | Max LLM turns, invalid JSON attempts, tool call totals enforced; not reset on resume |
| **M2f** — CLI integration | `cmd/warvis/main.go` (flag parsing) | `--phase`, `--case-id`, `--output` flags work; validate evidence path exists |

**Definition of Done**:
- `go test ./internal/hunt/...` passes (mock state transitions, persistence)
- Case registry creates `/cases/<uuid>/state.json` with correct structure
- Resume with `--case-id <uuid> --phase TRACE` loads prior state + budgets correctly

---

### Day 3 (2026-05-04, Sat) — Kill Switch Test 1: INITIALIZE→TRACE
**Goal**: Full INITIALIZE→TRACE flow with real MCP tool call.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M3a** — INITIALIZE handler | `internal/hunt/fsm.go` (InitState method) | Call `case.open` via MCP client, store case_id + sandbox |
| **M3b** — TRACE entry | `internal/hunt/fsm.go` (TraceState) | Transition to TRACE, stub tool list (timeline.build, log.query available) |
| **M3c** — End-to-end CLI | `cmd/warvis/main.go hunt <evidence>` | Spawns MCP server, calls case.open, transitions INITIALIZE→TRACE, creates state.json |
| **M3d** — Audit logging | `internal/hunt/fsm.go` (AuditLog append) | Append-only `/cases/<case_id>/audit.jsonl`, hash chain per entry |
| **M3e** — Kill switch test 1 | `make kill-switch-check-1` | Run `warvis hunt /evidence/test.img`, verify state.json shows TRACE, audit.jsonl has case_open + state_transition |

**Definition of Done**:
- `warvis hunt /evidence/test.img` succeeds without error
- `/cases/<case_id>/state.json` exists, contains `"current_state": "TRACE"`
- `/cases/<case_id>/audit.jsonl` has 2+ lines (case_opened, state_transition)
- `jq . /cases/<case_id>/audit.jsonl` parses all lines successfully

---

### Day 4 (2026-05-05, Sun) — Ollama + Gemma 4 Integration
**Goal**: LLM agent loop that calls MCP tools based on tool-call JSON.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M4a** — Ollama HTTP client | `pkg/ollama/client.go` (Chat endpoint) | Make requests to `http://127.0.0.1:11434/api/chat`, parse response |
| **M4b** — Tool schema injection | `internal/agent/prompt.go` (BuildSystemPrompt) | Construct Hunt Protocol system prompt with current state + available tools + JSON format spec |
| **M4c** — Tool-call parsing (Tier 3: prompt-only) | `internal/agent/tool_call.go` (ParseAction) | Extract `{action, tool_name, arguments}` from Gemma response; 3-retry loop for invalid JSON |
| **M4d** — Agent loop skeleton | `internal/agent/loop.go` (Run method) | Pseudo-loop: Gemma → parse action → switch (call_tool/state_complete/escalate) |
| **M4e** — MCP tool dispatch | `internal/agent/loop.go` (callTool) | Validate tool allowed in current state, forward to MCP client, sanitize output before Gemma |
| **M4f** — Conversation history | `internal/agent/loop.go` (trimHistory) | Sliding window (20 exchanges max), truncate tool outputs to 500 chars |

**Definition of Done**:
- `go test ./internal/agent/...` passes (mock Ollama, mock MCP)
- Ollama handshake works (probe Gemma 4 model availability)
- Agent loop runs 1 full iteration: LLM call → tool validation → MCP call → response feedback
- No crashes on invalid JSON (retries, escalates after 3 attempts)

---

### Day 5 (2026-05-05, Sun) — Kill Switch Test 2–3: Agent Loop + Tool Calls
**Goal**: Full hunt from TRACE state with Gemma making autonomous tool decisions.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M5a** — TRACE state agent loop | `internal/hunt/fsm.go` (TraceState entry) + `internal/agent/loop.go` | Kick off agent loop in TRACE state, allow timeline.build + log.query tool calls |
| **M5b** — Tool result sanitization | `internal/agent/tool_call.go` (SanitizeToolOutput) | Schema extraction, redaction, 500-char truncation |
| **M5c** — Tool output injection defense | `internal/agent/loop.go` (envelope wrapping) | Wrap tool output with "Forensic data (untrusted source)" prefix, never as bare user message |
| **M5d** — Budget enforcement | `internal/hunt/fsm.go` (checkBudgets) | Max 50 LLM turns per state; force transition to EXPOSE/LOCK if exceeded |
| **M5e** — Kill switch test 2 | `make kill-switch-check-2` | Run `warvis hunt /evidence/test.img`, Gemma calls timeline.build + log.query, verify audit.jsonl has tool_called + tool_result |
| **M5f** — Kill switch test 3 | `make kill-switch-check-3` | Gemma returns valid JSON action format, 3+ consecutive tool invocations without manual fix-up |

**Definition of Done**:
- `warvis hunt /evidence/test.img` completes TRACE state with ≥2 tool calls
- Gemma makes autonomous decisions (no hardcoded tool sequences)
- `/cases/<case_id>/audit.jsonl` shows: case_opened, state_transition (INIT→TRACE), tool_called (timeline.build), tool_result, tool_called (log.query), tool_result, gemma_response (state_complete)
- Zero malformed JSON from Gemma (or retried 3x then escalated)

---

### Day 6 (2026-05-06, Mon) — Minimal SCAN + Kill Switch Tests 4–5
**Goal**: Parse SCAN state, stub iocs.scan tool, finalize logging + CLI status command.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M6a** — SCAN state skeleton | `internal/hunt/fsm.go` (ScanState) | FSM transitions TRACE→SCAN, allows iocs.scan + memory.* + net.* tools |
| **M6b** — EXPOSE + LOCK stubs | `internal/hunt/fsm.go` (ExposeState, LockState) | Transition SCAN→EXPOSE→LOCK (minimal, no agent logic required for kill switch) |
| **M6c** — Structured JSONL validation | `internal/hunt/fsm.go` (AuditLog) | All audit.jsonl entries valid JSON-Lines, no truncation/corruption |
| **M6d** — warvis status CLI | `cmd/warvis/main.go status <case_id>` | Read state.json, emit JSON with current_state, case_id, timestamp, progress |
| **M6e** — Kill switch test 4 | `make kill-switch-check-4` | Verify audit.jsonl is valid JSONL (jq . succeeds on all lines) |
| **M6f** — Kill switch test 5 | `make kill-switch-check-5` | Run `warvis status <case_id>`, output contains `"current_state": "TRACE"` (or SCAN if progressed) |

**Definition of Done**:
- `warvis status a1b2c3d4-...` outputs valid JSON with state information
- `/cases/<case_id>/audit.jsonl` passes JSONL validation (no malformed lines)
- All 5 kill-switch tests pass
- No runtime panics or hangs (timeouts enforced)

---

### Day 7 (2026-05-07, Tue) — Integration Testing + Hardening
**Goal**: End-to-end integration, kill-switch validation, edge case handling.

| Milestone | Deliverable | Acceptance |
|-----------|-------------|-----------|
| **M7a** — Integration test suite | `tests/integration_test.go` | Full hunt: `warvis hunt /evidence/test.img` → INITIALIZE → TRACE (≥2 tools) → state.json + audit.jsonl valid |
| **M7b** — Error handling | `cmd/warvis/main.go` (error messages) | Missing evidence → exit 1 with clear message; MCP server unavailable → error with retry suggestion; Ollama down → error, do NOT crash |
| **M7c** — Timeout enforcement | `internal/hunt/fsm.go` (context deadlines) | Tool call timeout 120s; state duration 600s; total hunt 1800s (all configurable) |
| **M7d** — Resume capability | `cmd/warvis/main.go hunt ... --case-id <uuid> --phase TRACE` | Load prior state + budgets + conversation history, resume without reset |
| **M7e** — Makefile targets | `harness/find-evil/Makefile` (kill-switch-check) | Add `make kill-switch-check` target that runs all 5 tests in sequence, fails fast |
| **M7f** — Verification + sign-off | `plans/ITEM-212-find-evil/verify-phase3.md` | Document kill-switch pass/fail, blockers, handover to Phase 4 (demo + docs) |

**Definition of Done**:
- All 5 kill-switch tests PASS
- `make kill-switch-check` succeeds (all tests green)
- Integration test covers: spawn MCP server, case.open, timeline.build, state transitions, audit trail
- No panics, all errors caught + logged
- Go module ready for commit

---

### Post-Day 7 (2026-05-08 ~ 2026-05-09) — Buffer + Phase 4 Handover
If kill-switch passes:
- Hand off to Phase 4 (demo video, README updates, accuracy report finalization, Devpost submission prep)
- Executor role shifts to `pitch-writer` (demo script) + `verifier` (accuracy, compliance)
- Go bridge locked; no further changes unless Phase 4 finds bugs

---

## 3. File Implementation Priority Order

**Tier 0 — Must-have first (blocks everything else)**:
1. `warvis/go.mod` — Module init
2. `internal/mcp/transport.go` — stdio client (spawn server, read/write JSON)
3. `internal/mcp/protocol.go` — JSON-RPC types (Request, Response, RPCError)
4. `internal/mcp/client.go` — Initialize + ListTools (handshake)

**Tier 1 — Core FSM (Day 2)**:
5. `internal/hunt/state.go` — State interface + 5 concrete types
6. `internal/hunt/fsm.go` — FSM core, transitions, registry, audit log
7. `internal/hunt/transitions.go` — Typed transition rules + guards

**Tier 2 — Agent loop (Day 4)**:
8. `pkg/ollama/client.go` — Ollama HTTP client
9. `pkg/ollama/types.go` — Chat request/response types
10. `internal/agent/prompt.go` — System prompt builder
11. `internal/agent/tool_call.go` — Action parsing (call_tool/state_complete/escalate)
12. `internal/agent/loop.go` — Main agent loop, tool dispatch, sanitization

**Tier 3 — CLI + integration (Day 3, 6)**:
13. `cmd/warvis/main.go` — CLI entry (hunt, status, list commands)
14. `cmd/warvis/flags.go` — Flag definitions + validation
15. `internal/display/formatter.go` — Output formatting (JSON)
16. `harness/find-evil/Makefile` — kill-switch-check target

**Tier 4 — Config + helpers (ongoing)**:
17. `internal/config/config.go` — Configuration struct + defaults
18. `internal/display/logger.go` — Structured logging helper

---

## 4. Interface Contracts (Pre-Implementation Consensus)

Lock down these interfaces BEFORE implementation to avoid rework.

### 4.1 MCP Client Interface
```go
// internal/mcp/client.go
type Client interface {
    Initialize(ctx context.Context) (ServerInfo, error)
    ListTools(ctx context.Context) ([]Tool, error)
    CallTool(ctx context.Context, name string, args map[string]interface{}) (ToolResult, error)
    Close() error
}

type Tool struct {
    Name        string
    Description string
    InputSchema interface{}
}

type ToolResult struct {
    Content string // Sanitized tool output
    Error   error
}

type ServerInfo struct {
    ProtocolVersion string
    Capabilities    map[string]interface{}
}
```

### 4.2 Hunt FSM Interface
```go
// internal/hunt/fsm.go
type FSM interface {
    CurrentState() State
    IsToolAllowed(toolName string) bool
    CallTool(ctx context.Context, toolName string, args map[string]interface{}) (string, error)
    Transition(reason string) error
    Pause(reason string) error
    Resume() error
    GetBudgetStatus() BudgetStatus
}

type State interface {
    Name() string
    AllowedTools() []string
}

type BudgetStatus struct {
    MaxLLMTurns              int
    CurrentLLMTurns          int
    MaxInvalidJSONAttempts   int
    CurrentInvalidJSONAttempts int
    // ... etc
}
```

### 4.3 Agent Loop Interface
```go
// internal/agent/loop.go
type Loop interface {
    Run(ctx context.Context) error
    SetFSM(fsm hunt.FSM)
    SetMCP(mcp mcp.Client)
    SetOllama(ollama *ollama.Client)
}

type Action struct {
    Type      string                 // call_tool | state_complete | escalate
    ToolName  string                 // if call_tool
    Arguments map[string]interface{} // if call_tool
    Reason    string                 // if state_complete | escalate
}
```

### 4.4 CLI Handler Interface
```go
// cmd/warvis/main.go
type Command interface {
    Run(ctx context.Context, args []string) error
}

// Hunt command binds all components
type HuntCommand struct {
    fsm    hunt.FSM
    agent  agent.Loop
    mcp    mcp.Client
    ollama *ollama.Client
}
```

---

## 5. Test Strategy per Milestone

### Day 1 Tests (MCP Client)
- `mcp_transport_test.go`: Spawn mock server (cat + jq), exchange JSON-RPC messages
- `mcp_protocol_test.go`: Marshal/unmarshal Request/Response, validate error handling
- `mcp_client_test.go`: Initialize handshake, ListTools parsing (9 tools)

### Day 2 Tests (FSM)
- `hunt_fsm_test.go`: State transitions (INIT→TRACE, TRACE→SCAN, etc.), guard validation
- `hunt_registry_test.go`: Persist state.json, resume loading, budget counter persistence
- `hunt_transitions_test.go`: All typed rules fire correctly, max-count enforcement

### Day 3 Tests (Kill Switch 1)
- `integration_init_trace_test.go`: Full flow warvis hunt → state.json valid → audit.jsonl has 2+ events

### Day 4 Tests (Agent Loop)
- `ollama_client_test.go`: Mock HTTP server, Chat request/response parsing
- `agent_prompt_test.go`: System prompt injection with tool schemas + success criteria
- `agent_tool_call_test.go`: Parse action JSON, 3-retry loop on invalid JSON
- `agent_loop_test.go`: Mock Ollama + MCP, run 1 iteration (LLM → tool → feedback)

### Day 5 Tests (Kill Switch 2–3)
- `integration_trace_agent_test.go`: Full TRACE state with ≥2 real tool calls, budget enforcement

### Day 6 Tests (Kill Switch 4–5)
- `audit_jsonl_validation_test.go`: All lines valid JSON, no corruption
- `cli_status_test.go`: `warvis status <case_id>` outputs JSON, contains current_state

### Day 7 Tests (Integration)
- `integration_full_hunt_test.go`: INITIALIZE → TRACE → SCAN → EXPOSE → LOCK
- `integration_error_handling_test.go`: MCP unavailable, Ollama down, missing evidence
- `integration_resume_test.go`: Resume with `--case-id` + `--phase`

---

## 6. Acceptance Criteria per Milestone

### M1 (Day 1) ✓ DONE = Kill Switch Ready
- [ ] `go build ./cmd/warvis` succeeds (no build errors)
- [ ] `mcp.New(ctx, cmd)` spawns Python server process
- [ ] `client.Initialize()` returns ServerInfo with protocolVersion
- [ ] `client.ListTools()` returns 9 tools with correct names + inputSchema
- [ ] `go test ./internal/mcp/...` passes (all unit tests)

### M2 (Day 2) ✓ DONE = FSM Ready
- [ ] FSM implements `Current()`, `IsToolAllowed()`, `Transition()`, `GetBudgetStatus()`
- [ ] `registry.Save(state)` writes `/cases/<uuid>/state.json`
- [ ] `registry.Load(caseID)` restores prior state + budgets (not reset)
- [ ] All 5 state types defined with required fields
- [ ] Audit log is append-only, never overwritten
- [ ] `go test ./internal/hunt/...` passes

### M3 (Day 3) ✓ KILL SWITCH TEST 1 PASS
- [ ] `warvis hunt /evidence/test.img` exits 0 (no panic)
- [ ] `/cases/<uuid>/state.json` exists, contains `"current_state": "TRACE"`
- [ ] `/cases/<uuid>/audit.jsonl` has ≥2 lines (case_opened, state_transition)
- [ ] `jq . /cases/<uuid>/audit.jsonl` parses all lines without error
- [ ] Case ID is valid UUID v4

### M4 (Day 4) ✓ DONE = Agent Ready
- [ ] `ollama.Chat(ctx, request)` makes HTTP POST to localhost:11434
- [ ] System prompt includes current state + available tools + JSON format spec
- [ ] `ParseAction(response)` extracts valid Action struct from Gemma JSON
- [ ] Invalid JSON triggers retry (max 3), then escalate
- [ ] Agent loop runs without panic (mock MCP + Ollama)
- [ ] `go test ./internal/agent/...` passes

### M5 (Day 5) ✓ KILL SWITCH TEST 2–3 PASS
- [ ] `warvis hunt /evidence/test.img` progresses through TRACE state
- [ ] Gemma calls ≥2 tools (autonomously, not hardcoded)
- [ ] Tool result sanitized (≤500 chars, wrapped with "untrusted" prefix)
- [ ] `/cases/<uuid>/audit.jsonl` shows tool_called + tool_result events
- [ ] No invalid JSON output from Gemma (or retried 3x + escalated)
- [ ] Budget enforcement working (LLM turns, invalid attempts tracked)

### M6 (Day 6) ✓ KILL SWITCH TEST 4–5 PASS
- [ ] `warvis status <case_id>` outputs JSON with current_state + case_id + timestamp
- [ ] `/cases/<uuid>/audit.jsonl` passes JSONL validation
- [ ] SCAN → EXPOSE → LOCK transitions work
- [ ] `make kill-switch-check` target exists + runs all 5 tests
- [ ] All 5 kill-switch tests PASS (green)

### M7 (Day 7) ✓ READY FOR PHASE 4
- [ ] Integration tests pass (full hunt flow)
- [ ] Error handling tested (MCP down, Ollama down, missing evidence)
- [ ] Resume capability works (load prior state + budgets)
- [ ] Timeout enforcement working (no infinite loops)
- [ ] Go module ready for commit (go.mod + go.sum)
- [ ] verify-phase3.md written with pass/fail verdict + blocker list

---

## 7. Kill Switch Check Makefile Target

```makefile
.PHONY: kill-switch-check kill-switch-check-1 kill-switch-check-2 \
        kill-switch-check-3 kill-switch-check-4 kill-switch-check-5

kill-switch-check: kill-switch-check-1 kill-switch-check-2 \
                   kill-switch-check-3 kill-switch-check-4 \
                   kill-switch-check-5
	@echo "✓ All 5 kill-switch tests PASSED"
	@exit 0

kill-switch-check-1:
	@echo "Test 1: INITIALIZE → TRACE state transition..."
	@$(GO) run ./cmd/warvis hunt /evidence/test.img 2>&1 | tee /tmp/test1.log
	@jq . /tmp/case_id.json 2>/dev/null | grep -q '"current_state":"TRACE"' || \
	  { echo "✗ Test 1 FAILED: state.json missing or not TRACE"; exit 1; }
	@echo "✓ Test 1 PASSED"

kill-switch-check-2:
	@echo "Test 2: MCP tool invocation (case.open + timeline.build)..."
	@grep -q 'tool_called' /tmp/audit.jsonl && grep -q 'case.open' /tmp/audit.jsonl || \
	  { echo "✗ Test 2 FAILED: case.open not invoked"; exit 1; }
	@grep -q 'timeline.build' /tmp/audit.jsonl || \
	  { echo "✗ Test 2 FAILED: timeline.build not invoked"; exit 1; }
	@echo "✓ Test 2 PASSED"

# ... etc
```

---

## 8. Complexity Estimate & Risk

### Complexity: MEDIUM
- **MCP client**: ~500 lines (straightforward JSON-RPC stdio implementation)
- **FSM + registry**: ~800 lines (state machines, persistence)
- **Agent loop**: ~600 lines (Ollama integration, tool dispatching, sanitization)
- **CLI**: ~200 lines (flag parsing, command routing)
- **Total**: ~2100 lines of Go code

### Risk Assessment

| Risk | Probability | Mitigation |
|------|-------------|-----------|
| Ollama Gemma 4 tool-call format unclear | Medium | Day 4 includes probe test; Tier 3 fallback (prompt-only JSON) always works |
| Python MCP server subprocess dies silently | Low | Capture stderr, log it, error to user |
| Context timeout not enforced | Low | Explicit context.WithTimeout at each state/tool call |
| Go module dependency conflicts | Low | Lock to minimal deps (no external MCP SDK, use stdlib JSON-RPC) |
| Resume loads corrupted state.json | Low | Validate JSON schema on load, error if invalid |

---

## 9. Handover & Phase 4 Prerequisites

**If Kill Switch PASSES by 2026-05-09**:
- Go bridge locked (no further Go changes unless Phase 4 bug fixes)
- Phase 4 executor role: `pitch-writer` (demo script) + artifact finalization
- Artifacts to finalize:
  - Demo video script (terminal: `warvis hunt`, show state transitions, tool calls, audit log)
  - README: Go bridge setup + quick-start + limitations
  - Accuracy report: recall/precision on curated fixtures
  - Devpost submission package

**If Kill Switch FAILS by 2026-05-09**:
- Go bridge abandoned
- Revert to Phase 2 (Python-only orchestration)
- Accept lower hackathon placement (likely Honorable Mention)
- Redirect resources to Phase 4 finalization (demo, docs, submission)

---

## 10. Consensus Checkpoints

**Before Day 1 starts (2026-05-02 morning)**:
- [ ] Approve interface contracts (MCP, FSM, Agent, CLI)
- [ ] Confirm Ollama endpoint (localhost:11434 default, or remote URL?)
- [ ] Confirm Gemma 4 model name (gemma4:26b-a4b-instruct-q4_K_M or alternative?)
- [ ] Approve test fixtures (use /evidence/test.img from Phase 2, or new minimal test?)

**After Day 3 (2026-05-04 evening)**:
- [ ] Kill Switch Test 1 PASS: state transition works
- [ ] MCP client stable, no crashes
- [ ] Decision: Proceed to Agent Loop (Day 4) or pivot?

**After Day 5 (2026-05-06 evening)**:
- [ ] Kill Switch Tests 2–3 PASS: agent loop autonomous
- [ ] Gemma 4 tool-calling works (Tier 1/2/3, whichever)
- [ ] Decision: Proceed to Phase 4 or accept Go bridge incomplete?

---

## 11. Assumptions & Constraints

### Assumptions
- Ollama Gemma 4 26B running locally at localhost:11434
- Python MCP server can be spawned as subprocess via `python -m find_evil_mcp.server`
- `/evidence/` directory exists with test image (from Phase 2)
- Docker containers (SIFT) running for full end-to-end (but not required for kill-switch basic tests)

### Constraints
- **Time**: 7 days, hard cutoff 2026-05-09 23:59 KST
- **Scope**: Only INITIALIZE, TRACE, SCAN, EXPOSE, LOCK states (no advanced orchestration)
- **LLM**: Gemma 4 26B only (no swapping to Claude or GPT)
- **No external Go MCP SDK**: Use stdlib JSON-RPC only (simpler, fewer dependencies)
- **No agent framework**: No GenAI SDK, no LangChain-style abstractions (just JSON-RPC + HTTP)

---

## 12. Success Criteria (Final)

**Go bridge is successful if**:
1. ✓ All 5 kill-switch tests pass by 2026-05-09 23:59 KST
2. ✓ `warvis hunt <evidence>` spawns MCP server, calls tools, progresses FSM, writes audit trail
3. ✓ Gemma 4 makes autonomous tool decisions (not hardcoded sequences)
4. ✓ Zero panics, all errors caught + logged
5. ✓ Code ready for submission (go.mod locked, go.sum checked in)

**If all 5 criteria met**: Phase 4 handover (demo, docs, submission finalization). Expected placement: **1st–3rd** (based on accuracy + autonomous execution quality).

**If any kill-switch fails**: Accept "Go bridge incomplete, revert to Python", expect **Honorable Mention** placement.

---

## 13. Executor Handoff Template

**Executor**: Apply this checklist as you work.

- [ ] Day 1 EOD: Tier 0 files written, `go build ./cmd/warvis` succeeds
- [ ] Day 2 EOD: FSM compiles, state types + transitions defined, registry tested
- [ ] Day 3 EOD: Kill Switch Test 1 PASSES (INITIALIZE→TRACE, state.json, audit.jsonl)
- [ ] Day 4 EOD: Agent loop compiles, mock tests pass
- [ ] Day 5 EOD: Kill Switch Tests 2–3 PASS (tool calls, autonomous decisions)
- [ ] Day 6 EOD: Kill Switch Tests 4–5 PASS (JSONL validation, warvis status)
- [ ] Day 7 EOD: Integration tests pass, verify-phase3.md written, ready for Phase 4

**Commit frequency**: Commit at end of each day (Day 1 → Day 7 = 7 commits minimum).

---

**Plan authored**: 2026-05-02  
**Status**: Ready for execution  
**Next step**: Executor confirms interface contracts, starts Day 1 (MCP client + Go module)

