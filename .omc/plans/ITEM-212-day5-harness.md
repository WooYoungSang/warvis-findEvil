# Harness Config: ITEM-212-day5 — Kill Switch Tests 2–3 (TRACE agent loop)

**UoW**: ITEM-212-FIND-EVIL Day 5  
**Date**: 2026-05-06 (Mon)  
**Scope**: Full hunt from TRACE state with Gemma making autonomous tool decisions (M5a–M5f)  
**Kill Switch Deadline**: 2026-05-09 23:59 KST

---

## P1: Always-on Layers

### CLAUDE.md Evaluation
- **Quality**: `map` (82 lines, well-structured agent delegation table)
- **Verdict**: ✓ Harness is stable

### Local Agents Available
- `warvis-orchestrator.md` — THIS AGENT (lifecycle management)
- `warvis-maker.md` — TDD implementation
- `warvis-verifier.md` — Quality gates (lint, test, security)
- `tdd-red.md`, `tdd-green.md`, `tdd-refactor.md` — TDD cycle coordination
- `contract-auditor.md` — Interface contract verification
- `spec-auditor.md` — Spec→implementation gap detection

### Reuse Assessment
Day 4 completed:
- ✓ Ollama client (pkg/ollama/) — 3 test functions
- ✓ Agent loop skeleton (internal/agent/loop.go) — skeleton with callOllama
- ✓ Tool-call parser (internal/agent/tool_call.go) — ParseAction + retry logic
- ✓ System prompt builder (internal/agent/prompt.go) — BuildSystemPrompt function

Day 5 builds on:
- Reuse existing FSM state types (InitializeState, TraceState, etc.)
- Reuse existing agent loop structure (Run method, callOllama, conversation history)
- NEW: TRACE state entry point + tool call integration + budget enforcement + audit events

---

## P2: On-demand Layers

### Skills Reuse
- **TDD pattern**: Reuse red→green→refactor cycle from Day 4
- **Agent loop pattern**: Reuse existing Loop struct + Run method
- **Audit logging**: Reuse existing AuditLog append-only pattern

### No External Deps
- Stdlib JSON-RPC only (no MCP SDK)
- Ollama HTTP client already built
- Gemma 4 prompt-only tool-call (Tier 3, no native function calling)

---

## P3: Deterministic Layers

### Test Commands
```bash
# Go test (from warvis/ dir)
cd /home/jang/Workspace/warvis-findEvil/warvis && go test ./... -short

# Go build
cd /home/jang/Workspace/warvis-findEvil/warvis && go build -o bin/warvis ./cmd/warvis

# Go lint
cd /home/jang/Workspace/warvis-findEvil/warvis && golangci-lint run ./...

# Kill switch test 1 (existing, Day 3)
make -C /home/jang/Workspace/warvis-findEvil/harness/find-evil kill-switch-check-1

# Kill switch test 2 (Day 5) — to be created
make -C /home/jang/Workspace/warvis-findEvil/harness/find-evil kill-switch-check-2

# Kill switch test 3 (Day 5) — to be created
make -C /home/jang/Workspace/warvis-findEvil/harness/find-evil kill-switch-check-3
```

### Gate Strategy (Day 5)
1. **Lint**: `golangci-lint run ./...`
2. **Test**: `go test ./...` (all packages, all tests pass)
3. **Build**: `go build -o bin/warvis ./cmd/warvis` (no errors)
4. **Kill switch**: `make kill-switch-check-1` (Day 3 regression check)
5. **Kill switch**: `make kill-switch-check-2` (Day 5 tool calls)
6. **Kill switch**: `make kill-switch-check-3` (Day 5 JSON autonomy)

---

## P4: Measurable Layers

### Baseline Eval (Day 4 Artifacts)
- ✓ Agent loop tests pass (loop_test.go)
- ✓ Ollama client tests pass (client_test.go)
- ✓ Tool-call parser tests pass (tool_call_test.go)
- ✓ Prompt builder tests pass (prompt_test.go)

### Day 5 Acceptance Metrics
- `audit.jsonl` contains: case_opened → state_transition(INIT→TRACE) → tool_called(timeline.build) → tool_result → tool_called(log.query) → tool_result → gemma_response(state_complete)
- Gemma makes ≥2 autonomous tool calls (not hardcoded)
- Zero manual JSON fix-ups (or retried 3x then escalated)
- All tests pass + kill-switch-check-1 regression check passes

---

## P5: Collaboration Map

| Task | Agent | Notes |
|------|-------|-------|
| **Implementation (M5a–M5f)** | `direct` (warvis-orchestrator + TDD) | TDD red→green→refactor per milestone; no sub-agents needed |
| **Code review** | `contract-auditor` (if divergence detected) | Verify FSM→agent integration matches phase3-go-bridge.md spec |
| **Verification** | `warvis-verifier` (lint→test→security gates) | Run after all code complete |
| **HITL gates** | USER (if HIGH/CRITICAL risk detected) | Budget enforcement, injection defense, Gemma JSON parsing edge cases |

---

## P6: HITL Points

### HIGH Risk Work (require user confirmation before proceeding)
1. **Budget enforcement** (M5d) — Force state transition if LLM turns exceed 50 per state
   - Risk: Agent could get stuck or loop if budget logic is wrong
   - Mitigation: Mock Ollama + unit test with specific turn limits

2. **Tool output injection defense** (M5c) — Wrap tool output with "Forensic data (untrusted source)" prefix
   - Risk: If envelope wrapping fails, raw tool output could be interpreted as user message
   - Mitigation: Unit test envelope wrapping separately; don't fold into callTool

3. **Gemma JSON parsing failure** (3-retry + escalate) — Malformed JSON from Gemma
   - Risk: Infinite loop if retry logic is broken
   - Mitigation: Unit test retry exhaustion → escalate path; mock Ollama with invalid JSON

### CRITICAL Work (block immediately)
- None for Day 5 (all work is bounded by tests + mocks)

---

## P7: Dead Weight Check

**CLAUDE.md sections to review after Day 5**:
- If any agent delegation rules were NOT used this cycle, flag for possible removal next sprint
- Currently: all 15 agents in table are used or standby; no pruning needed

---

## Implementation Plan

### M5a — TRACE state agent loop entry
**File**: `internal/hunt/fsm.go` (TraceState type + Run method)
- Add `Run(ctx context.Context, loop *agent.Loop) error` to TraceState
- FSM.Transition("TRACE") → TraceState.Run(ctx, loop)
- Agent loop runs until state_complete or escalate

**File**: `cmd/warvis/main.go` (main hunt flow)
- After INITIALIZE→TRACE transition, spawn agent loop
- Pass FSM + MCP client + Ollama client to Loop.Run

### M5b — Tool output sanitization
**File**: `internal/agent/tool_call.go` (SanitizeToolOutput function)
- Extract JSON schema from tool result
- Redact sensitive fields (file paths, IP addresses, process names)
- Truncate to 500 chars
- Unit test: sanitize → verify length + no secrets

### M5c — Tool output injection defense
**File**: `internal/agent/loop.go` (callTool method)
- Wrap sanitized tool output: `"Forensic data (untrusted source): " + sanitized`
- Add to conversation history with this prefix
- Unit test: verify prefix is always present in history

### M5d — Budget enforcement
**File**: `internal/hunt/fsm.go` (checkBudgets function)
- Called after each LLM turn: increment CurrentLLMTurns, check if >= MaxLLMTurns (50)
- If exceeded: log audit event, force transition to EXPOSE or LOCK
- Unit test: mock Ollama for exactly 50 turns, verify force transition on turn 51

### M5e — Kill switch test 2
**File**: `harness/find-evil/Makefile` (kill-switch-check-2 target)
- Run `warvis hunt /evidence/test.img`
- Verify `audit.jsonl` contains:
  - Event type: tool_called (tool_name: timeline.build)
  - Event type: tool_result (result_preview from timeline.build)
  - Event type: tool_called (tool_name: log.query)
  - Event type: tool_result (result_preview from log.query)

### M5f — Kill switch test 3
**File**: `harness/find-evil/Makefile` (kill-switch-check-3 target)
- Run `warvis hunt /evidence/test.img` with Gemma
- Verify Gemma returns valid JSON action format (no manual fix needed)
- Verify 3+ consecutive tool invocations succeed without retry exhaustion

---

## File Implementation Priority

**Tier 0 — Must-have first**:
1. `internal/hunt/fsm.go` — Add TraceState.Run method + checkBudgets
2. `internal/agent/loop.go` — Integrate FSM budget checks, wrap tool output
3. `internal/agent/tool_call.go` — Add SanitizeToolOutput function

**Tier 1 — Integration**:
4. `cmd/warvis/main.go` — Spawn agent loop in TRACE state
5. `harness/find-evil/Makefile` — Add kill-switch-check-2 + kill-switch-check-3 targets

**Tier 2 — Tests + Verification**:
6. Update all *_test.go files with new test cases
7. Run full `go test ./...` + kill-switch tests

---

## Risk Assessment

| Component | Risk | Mitigation |
|-----------|------|-----------|
| **Budget enforcement** | Loop if budget check is wrong | Unit test turn-count exactly at limit + 1 over |
| **Tool output injection** | Wrapped output parsed as user input | Unit test envelope prefix in history |
| **Gemma JSON parsing** | Infinite retry if logic broken | Unit test retry-exhaustion → escalate path |
| **FSM state transition** | Loop in TRACE if state_complete never sent | Unit test mock Ollama for state_complete |
| **Audit logging** | Missing events in audit.jsonl | Unit test each event type individually |

---

## Success Criteria

✅ **All milestones complete**:
- M5a: TRACE state agent loop runs (no panic)
- M5b: Tool output sanitization works (500 chars, no secrets)
- M5c: Envelope prefix always present (verified in history)
- M5d: Budget enforcement triggers force-transition (at turn 50)
- M5e: kill-switch-check-2 PASSES (≥2 tool calls in audit.jsonl)
- M5f: kill-switch-check-3 PASSES (3+ Gemma tool invocations, valid JSON)

✅ **All gates pass**:
- `go test ./...` → all packages passing
- `golangci-lint run ./...` → no errors
- `make kill-switch-check-1` → regression check passes (Day 3 still works)
- `make kill-switch-check-2` → tool calls logged correctly
- `make kill-switch-check-3` → Gemma JSON parsing works

✅ **No breaking changes**:
- src/find_evil_mcp/ untouched
- Day 1–4 code unchanged (only new methods in fsm.go + loop.go)
- All existing tests still pass

---

## Timeline

- **Day 5 morning**: Harness evaluation + planning (this doc) ✓
- **Day 5 afternoon**: M5a–M5c implementation + unit tests
- **Day 5 evening**: M5d budget enforcement + M5e Makefile targets
- **Day 6**: M5f kill-switch test 3 + full verification + Day 6 work (M6a–M6f)

**Kill Switch Deadline**: 2026-05-09 23:59 KST (3 days after Day 5)

---

Generated by warvis-orchestrator (harness bootstrap phase) on 2026-05-06.
