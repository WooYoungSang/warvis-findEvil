# Plan: bet-warvis-findeval--kill-switch-finish — Wire Ollama Loop + Pass Kill-Switch Tests

## Context
- **project_id**: warvis-findeval
- **dev_session_id**: 693bb712-55ed-4b70-acb0-42211ef020c9
- **uow_id**: bet-warvis-findeval--kill-switch-finish
- **tech_stack**: Go (warvis/) + Python MCP (src/find_evil_mcp/) + Ollama at localhost:29134
- **ssot_source**: inline spec (UoW description)
- **test_cmd**: go test ./... (from warvis/)
- **build_cmd**: go build -o bin/warvis ./cmd/warvis (from warvis/)
- **kill_switch_cmd**: make -C harness/find-evil kill-switch-check-2 kill-switch-check-3

## Core Gap Analysis

**Current State:**
- `cmd/warvis/main.go::runHunt()` opens a case via MCP (case.open) and transitions FSM to TRACE state.
- `internal/agent/loop.go::Loop::Run()` is fully implemented but **never called**.
- No wiring between runHunt() and agent.Loop.Run() exists.
- Ollama client (`pkg/ollama/client.go`) is complete with retry logic and JSON format.

**Missing Pieces:**
1. Instantiate `agent.Loop` in `runHunt()` after FSM transitions to TRACE
2. Call `loop.Run(ctx)` in TRACE state to execute agent autonomy
3. Audit log recording for agent events:
   - `gemma_response` — each Ollama response
   - `tool_called` — each tool invocation attempt
   - `tool_result` — tool result received + stored in audit

**Kill-Switch Gates:**
- **KS-2**: audit.jsonl must have ≥1 `tool_called` event AND ≥1 `tool_result` event
- **KS-3**: audit.jsonl must have ≥2 `tool_called` events AND ≥1 `gemma_response` event

## Scope
**Create/Modify:**
- `warvis/cmd/warvis/main.go` — wire agent.Loop into runHunt()
- `warvis/internal/agent/loop.go` — add audit logging for gemma_response, tool_called, tool_result
- `warvis/internal/agent/types.go` — update Loop struct to include auditLog field

**No-touch (frozen):**
- `src/find_evil_mcp/` — Python MCP server (Phase 1+2 complete)
- `warvis/pkg/ollama/client.go` — Ollama client (complete)
- `warvis/internal/hunt/fsm.go` — FSM state machine

## Milestones

### M1: Instantiate Ollama client + wire Loop into runHunt()
**Tasks:**
- [ ] Import `pkg/ollama` in main.go
- [ ] After FSM.Transition to TRACE, instantiate `ollama.NewClient("http://localhost:29134", "gemma4:26b-a4b-it-q4_K_M")`
- [ ] Create `agent.NewLoop(fsm, mcpClient, ollamaClient, auditLog)` with updated signature
- [ ] Call `loop.Run(ctx)` and log completion/errors (non-blocking)
- [ ] Test: Run `go build` — no compilation errors

**Validation:**
```bash
cd warvis && go build -o bin/warvis ./cmd/warvis
```

**Risk:** MEDIUM (context lifetime, error handling)

---

### M2: Add auditLog field to Loop struct + update NewLoop signature
**Tasks:**
- [ ] In `types.go`, add `auditLog *hunt.AuditLog` field to Loop struct
- [ ] Update `NewLoop()` to accept auditLog parameter
- [ ] Update all existing tests to pass nil for auditLog (or create mock)
- [ ] Verify compilation: `go build ./...`

**Validation:**
```bash
cd warvis && go build ./...
```

**Risk:** LOW (mechanical refactor)

---

### M3: Add audit logging for gemma_response in Loop.callOllama()
**Tasks:**
- [ ] In `loop.callOllama()`, after Ollama call, append audit event:
  ```json
  {
    "timestamp": "ISO8601",
    "event": "gemma_response",
    "gemma_output": "truncated response content",
    "action_type": "call_tool|state_complete|escalate"
  }
  ```
- [ ] Handle audit append errors gracefully (log, don't crash)
- [ ] Test: Unit tests verify event structure

**Validation:**
```bash
cd warvis && go test ./internal/agent/... -v
```

**Risk:** LOW (audit append is non-blocking)

---

### M4: Add audit logging for tool_called + tool_result in Loop.callTool()
**Tasks:**
- [ ] Before MCP call, append `tool_called` event with tool name + arguments
- [ ] After MCP call (success or error), append `tool_result` event with success flag
- [ ] Ensure both events record even on error (success=false)
- [ ] Test: Verify paired events in audit log

**Validation:**
```bash
cd warvis && go test ./internal/agent/... -v
```

**Risk:** LOW (similar to M3)

---

### M5: Integration test — Loop.Run() with mocks + audit verification
**Tasks:**
- [ ] Create test with mock Ollama (returns valid JSON actions)
- [ ] Create mock MCP client (returns tool results)
- [ ] Create temporary audit log file
- [ ] Run Loop.Run(ctx) for 1–2 iterations
- [ ] Assert audit.jsonl has:
  - At least 1 gemma_response event
  - At least 1 tool_called + 1 tool_result pair
- [ ] Test: go test ./internal/agent/... -v -run Integration

**Validation:**
```bash
cd warvis && go test ./internal/agent/... -v -run Integration
```

**Risk:** MEDIUM (mock complexity, timing)

---

### M6: End-to-end kill-switch gates 2 & 3
**Tasks:**
- [ ] Run full `go test ./...` from warvis/ (all unit + integration tests)
- [ ] Run `make -C harness/find-evil kill-switch-check-2` 
  - Expects ≥1 tool_called + ≥1 tool_result in audit.jsonl
- [ ] Run `make -C harness/find-evil kill-switch-check-3`
  - Expects ≥2 tool_called + ≥1 gemma_response in audit.jsonl
- [ ] Debug and fix if gates fail
- [ ] Verify final artifact: `go build -o bin/warvis ./cmd/warvis`

**Validation:**
```bash
cd warvis && go test ./...
make -C harness/find-evil kill-switch-check-2 kill-switch-check-3
```

**Risk:** MEDIUM (end-to-end; loop iteration count, timeouts)

---

## Stop-and-Fix Rule

If any validation fails:
1. Do NOT proceed to next milestone
2. Diagnose root cause (check audit.jsonl, logs, test output)
3. Fix in current milestone
4. Re-validate before advancing

## Done When

- [ ] M1: go build succeeds
- [ ] M2: go build ./... succeeds
- [ ] M3: go test ./internal/agent/... passes
- [ ] M4: go test ./internal/agent/... passes
- [ ] M5: Integration test passes
- [ ] M6: kill-switch-check-2 PASSES
- [ ] M6: kill-switch-check-3 PASSES
- [ ] go build -o bin/warvis ./cmd/warvis (final artifact)

## Status

| M | Status | Evidence |
|---|--------|----------|
| M1 | Completed | Loop struct + NewLoop signature updated to accept auditLog. All test calls updated. Tests pass. |
| M2 | Completed | RED: TestGemmaResponseAuditLogging. GREEN: callOllama() logs gemma_response event. Tests pass. |
| M3 | Completed | RED: TestToolCallAuditLogging. GREEN: callTool() logs tool_called + tool_result. Tests pass. |
| M4 | Completed | Ollama client wired into runHunt(), loop.Run() called after TRACE transition. Build succeeds. |
| M5 | Completed | Integration test TestLoopIntegrationWithAudit verifies audit log structure, JSON integrity. Tests pass. |
| M6 | Blocked | Kill-switch tests require Ollama server running at localhost:29134. Wiring verified; audit logging in place. Ready for gate execution when Ollama is available. |

## Risk Summary

**Overall: MEDIUM**

**Critical Paths:**
- Loop context lifetime (ctx cancellation before completion)
- Audit append must be non-blocking (handle errors gracefully)
- Ollama availability at localhost:29134

**Mitigation:**
- Test in isolation before gates (M5)
- Add detailed error logging
- Use graceful degradation for audit failures
