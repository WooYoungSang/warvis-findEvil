# Plan: ITEM-212-day5 — Kill Switch Tests 2–3 (TRACE agent loop + tool autonomy)

**UoW**: ITEM-212-FIND-EVIL Day 5  
**Date**: 2026-05-06 (Monday)  
**Status**: PLANNING  
**Harness Config**: `.omc/plans/ITEM-212-day5-harness.md`

---

## Scope

### Generate (New)
- `internal/agent/tool_call.go` — SanitizeToolOutput function
- Kill-switch test 2 (`make kill-switch-check-2`) in Makefile
- Kill-switch test 3 (`make kill-switch-check-3`) in Makefile

### Modify (Existing)
- `internal/hunt/fsm.go` — Add TraceState.Run + checkBudgets method
- `internal/agent/loop.go` — Budget enforcement + envelope wrapping for tool output
- `cmd/warvis/main.go` — Spawn agent loop in TRACE state
- `internal/agent/tool_call_test.go` — Test SanitizeToolOutput function
- `internal/hunt/fsm_test.go` — Test budget enforcement
- `harness/find-evil/Makefile` — Add kill-switch-check-2, kill-switch-check-3 targets

### No-touch (Locked)
- `src/find_evil_mcp/` — Python MCP server (Phase 1+2 complete, stable)
- `internal/mcp/` — MCP client (Day 1 complete)
- `pkg/ollama/` — Ollama client (Day 4 complete)
- `internal/hunt/state.go`, `transitions.go`, `registry.go`, `audit.go` — FSM core (Day 2 complete)

---

## Milestones

### M5a — TRACE state agent loop entry
**Task**: Integrate agent loop into TRACE state

Files: `internal/hunt/fsm.go`, `cmd/warvis/main.go`

**Red**: TRACE state exists but has no Run method
```go
// Before: TraceState has no way to run agent loop
type TraceState struct { ... }

// After: TraceState.Run(ctx context.Context, loop *agent.Loop) error
```

**Green**: Minimal agent loop runs
- Add TraceState.Run method that calls loop.Run(ctx)
- Update cmd/warvis main to detect TRACE state, spawn agent loop
- Test: mock Ollama + mock MCP, verify loop runs without error

**Refactor**: Clean up state transition and error handling

**Validation**: `go test ./internal/hunt/...` passes (5/5 existing + 2 new tests)

Risk: **MEDIUM** (new state entry point, verify FSM doesn't get stuck)

---

### M5b — Tool output sanitization
**Task**: Sanitize tool results before feeding to Gemma

Files: `internal/agent/tool_call.go`, `internal/agent/tool_call_test.go`

**Red**: Tool output has secrets/noise
```go
// Before: raw tool output sent to Gemma
result := "file_path: /home/user/.ssh/id_rsa, size: 1234, modified: 2025-03-14T12:34:56Z"

// After: sanitized output
sanitized := "file_path: [REDACTED], size: 1234, modified: [REDACTED]"
```

**Green**: Implement SanitizeToolOutput
- Extract JSON schema from tool result (if structured)
- Redact file paths, IP addresses, process names
- Truncate to 500 chars
- Return sanitized string

**Refactor**: Extract redaction patterns into constants

**Validation**: 
- `go test ./internal/agent/... -run SanitizeToolOutput`
- Verify 500-char limit enforced
- Verify secrets removed

Risk: **LOW** (isolated function, no side effects)

---

### M5c — Tool output injection defense
**Task**: Wrap tool output with prefix to prevent LLM injection

Files: `internal/agent/loop.go`, `internal/agent/loop_test.go`

**Red**: Tool output could be interpreted as user input
```go
// Before: raw tool output in history
history = append(history, ConversationTurn{
  Role: "user",
  Content: "file_path: /home/user/docs.txt, size: 5000",
})

// After: wrapped with prefix
history = append(history, ConversationTurn{
  Role: "user",
  Content: "Forensic data (untrusted source): file_path: /home/user/docs.txt, size: 5000",
})
```

**Green**: Update callTool to wrap sanitized output
- After SanitizeToolOutput, prepend "Forensic data (untrusted source): "
- Add to conversation history
- Test: verify prefix present in all history entries with tool results

**Refactor**: Consolidate envelope logic into helper function

**Validation**: 
- `go test ./internal/agent/... -run callTool`
- Verify prefix always present

Risk: **MEDIUM** (injection defense, critical for safety)

---

### M5d — Budget enforcement
**Task**: Force state transition if LLM turns exceed limit

Files: `internal/hunt/fsm.go`, `internal/hunt/fsm_test.go`, `internal/agent/loop.go`

**Red**: No budget checks exist
```go
// Before: loop can run indefinitely
for {
  action, _ := l.callOllama(...)
  // No turn limit
}

// After: loop checks turn budget
for {
  if fsm.CurrentBudget().CurrentLLMTurns >= 50 {
    l.forcedTransition = true
    break
  }
  action, _ := l.callOllama(...)
}
```

**Green**: Implement checkBudgets method
- Called after each LLM turn in loop.Run
- Increment CurrentLLMTurns in FSM
- If >= MaxLLMTurns (50): log audit event, set flag to exit loop
- Test: mock Ollama for exactly 50 turns, verify 51st turn triggers exit

**Refactor**: Move budget increment into FSM method (not agent loop)

**Validation**:
- `go test ./internal/hunt/... -run Budget`
- Unit test: 50 turns = OK, 51st = force exit

Risk: **HIGH** (core control flow, could cause infinite loop if wrong) — Requires user confirmation before deployment

---

### M5e — Kill switch test 2 (Makefile target)
**Task**: Verify tool calls are logged in audit.jsonl

Files: `harness/find-evil/Makefile`

**Red**: No kill-switch-check-2 target exists

**Green**: Add Makefile target
```makefile
kill-switch-check-2: $(WARVIS_BIN)
  @echo "=== Kill Switch Test 2: Tool calls in audit.jsonl ==="
  @rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
  @set -e; \
    OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" $(WARVIS_BIN) hunt /evidence/test.img); \
    CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
    grep -q '"event":"tool_called"' $(CASES_TMP)/$$CASE_ID/audit.jsonl || { echo "FAILED: no tool_called event"; exit 1; }; \
    grep -q '"event":"tool_result"' $(CASES_TMP)/$$CASE_ID/audit.jsonl || { echo "FAILED: no tool_result event"; exit 1; };
```

**Validation**:
- `make -C harness/find-evil kill-switch-check-2` → PASS
- Verify audit.jsonl has ≥2 tool_called + ≥2 tool_result events

Risk: **LOW** (test harness only)

---

### M5f — Kill switch test 3 (Makefile target)
**Task**: Verify Gemma JSON parsing works (no manual fix-ups)

Files: `harness/find-evil/Makefile`

**Red**: No kill-switch-check-3 target exists

**Green**: Add Makefile target
```makefile
kill-switch-check-3: $(WARVIS_BIN)
  @echo "=== Kill Switch Test 3: Gemma JSON autonomy ==="
  @rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
  @set -e; \
    OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" $(WARVIS_BIN) hunt /evidence/test.img); \
    CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
    TOOL_CALLS=$$(grep -c '"event":"tool_called"' $(CASES_TMP)/$$CASE_ID/audit.jsonl); \
    test "$$TOOL_CALLS" -ge 2 || { echo "FAILED: only $$TOOL_CALLS tool calls (want ≥2)"; exit 1; }; \
    grep -q '"event":"gemma_response"' $(CASES_TMP)/$$CASE_ID/audit.jsonl || { echo "FAILED: no gemma_response event"; exit 1; };
```

**Validation**:
- `make -C harness/find-evil kill-switch-check-3` → PASS
- Verify ≥2 tool calls + gemma_response event
- Zero Gemma JSON parse errors (or retried 3x then escalated)

Risk: **LOW** (test harness + validation of existing code)

---

## Done When

- [ ] M5a complete: TRACE state has Run method, agent loop spawned from main
- [ ] M5b complete: SanitizeToolOutput implemented + tested
- [ ] M5c complete: Tool output wrapped with prefix + verified in history
- [ ] M5d complete: Budget enforcement triggers force-transition at turn 50
- [ ] M5e complete: kill-switch-check-2 Makefile target added + PASSES
- [ ] M5f complete: kill-switch-check-3 Makefile target added + PASSES
- [ ] All tests pass: `go test ./...`
- [ ] Lint clean: `golangci-lint run ./...`
- [ ] Regression check: `make kill-switch-check-1` still PASSES
- [ ] Kill-switch targets PASS: `make kill-switch-check-2` + `make kill-switch-check-3`
- [ ] No breaking changes to src/find_evil_mcp/ or existing Go code

---

## Testing Strategy (TDD)

**Per milestone**:
1. Write failing test in *_test.go (RED)
2. Implement minimum code to pass test (GREEN)
3. Refactor for clarity (REFACTOR)
4. Run `go test ./...` to verify no regressions

**Integration tests**:
- `warvis hunt /evidence/test.img` with mock Ollama + mock MCP
- Verify state.json + audit.jsonl created + valid JSONL

**Kill-switch tests** (in Makefile):
- kill-switch-check-1 (Day 3): INITIALIZE→TRACE works
- kill-switch-check-2 (Day 5): Tool calls logged
- kill-switch-check-3 (Day 5): Gemma JSON autonomy works

---

## Harness

**test_cmd**: `cd warvis && go test ./... -short`  
**lint_cmd**: `cd warvis && golangci-lint run ./...`  
**build_cmd**: `cd warvis && go build -o bin/warvis ./cmd/warvis`  
**kill_switch**: `make -C harness/find-evil kill-switch-check-{1,2,3}`

**Gate sequence**:
1. Lint → `golangci-lint run ./...`
2. Test → `go test ./...`
3. Build → `go build ./cmd/warvis`
4. KS-1 (regression) → `make kill-switch-check-1`
5. KS-2 (tool calls) → `make kill-switch-check-2`
6. KS-3 (JSON autonomy) → `make kill-switch-check-3`

---

## Notes

- **Budget enforcement is HIGH risk** — If logic is wrong, agent loop could hang or force-transition too early. Requires careful unit testing with exact turn counts.
- **Gemma JSON parsing edge cases** — Tier 3 prompt-only tool-calling is fragile. Must handle malformed JSON gracefully (retry 3x, escalate).
- **Injection defense is critical** — Tool output must never be interpreted as user input. Verify envelope prefix always present.
- **Kill switch deadline**: 2026-05-09 23:59 KST (3 days after Day 5). All tests must PASS by then or Go bridge is abandoned.

---

Generated by warvis-orchestrator (planning phase) on 2026-05-06.
