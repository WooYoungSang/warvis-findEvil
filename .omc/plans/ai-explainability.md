# Plan: ai-explainability — Expand gemma_response audit event

## Context
- **project_id**: warvis-findEvil
- **dev_session_id**: ba8ed909-f8e5-4171-8da1-9e454bc5e281
- **tech_stack**: Go 1.x, pytest (Python MCP locked)
- **ssot_source**: inline spec (UoW ai-explainability)
- **key_constraint**: Kill-switch test (5/5 PASS) non-negotiable by 2026-05-09

## Scope

### Files to Modify (Targeted)
1. `warvis/internal/agent/loop.go` (lines 117-129 → callOllama audit append)
2. `warvis/internal/agent/loop_test.go` (new test: audit event schema validation)

### Files NOT to Modify
- `src/find_evil_mcp/` (Python MCP — Phase 1+2 locked, DO NOT BREAK)
- `warvis/internal/agent/parser.go` (Action struct — no changes needed)
- `warvis/internal/hunt/fsm.go` (FSM interface — no changes needed)
- Other warvis/* files unless signature change required
- `harness/find-evil/*` except kill-switch regression run

### No-Go Scope
- Do not change audit event structure in backward-incompatible way
- Do not modify JSONL serialization
- Do not alter FSM state machine logic

## Approach

**Parameter vs Internal Call Decision:**
- Chose **internal call**: `l.fsm.CurrentState().Name()` directly in `callOllama()`
- Rationale: Loop struct already holds fsm reference; avoids parameter bloat; cleanest for audit logging location
- No signature change to `callOllama()` required — keeps interface stable

## Milestones

### M1: Write failing test (RED)
- **Task**: Add test in `loop_test.go` that expects 5 new audit event fields
  - `tool_name`, `arguments`, `reason`, `raw_output`, `current_state`
  - Test MUST fail because code not yet implemented
  - Use temp audit log file pattern from existing tests (TestGemmaResponseAuditLogging, TestToolCallAuditLogging)
- **Validation**: `cd warvis && go test ./... -run TestExpandedGemmaResponseAudit -v`
  - Expected: FAIL (missing fields in audit event map)
- **Risk**: LOW (test-only, no production code touched)
- **Est. Time**: 10 min

### M2: Implement audit event expansion (GREEN)
- **Task**: Modify `callOllama()` in loop.go (lines 117-129)
  - Replace hardcoded 4-field map with 9-field map:
    ```go
    _ = l.auditLog.Append(map[string]interface{}{
        "timestamp":     time.Now().UTC().Format(time.RFC3339),
        "event":         "gemma_response",
        "action_type":   action.Type,
        "tool_name":     action.ToolName,         // NEW: directly from parsed Action
        "arguments":     action.Arguments,         // NEW: directly from parsed Action
        "reason":        action.Reason,            // NEW: untruncated reasoning
        "raw_output":    truncated,                // RENAMED: was gemma_output
        "current_state": l.fsm.CurrentState().Name(),  // NEW: FSM state at decision time
    })
    ```
  - Verify all Action fields populated correctly (done by ParseAction already)
  - Verify FSM interface provides CurrentState().Name()
- **Validation**: `cd warvis && go test ./... -run TestExpandedGemmaResponseAudit -v`
  - Expected: PASS (all 5 new fields + renamed field present)
- **Risk**: MEDIUM (first mutation to audit schema; must verify backward compat)
- **Est. Time**: 15 min

### M3: Verify no regressions (REFACTOR + INTEGRATION)
- **Task**: Run full Go test suite + kill-switch gate
  - `cd warvis && go test ./...` → all tests PASS
  - `cd warvis && go build -o bin/warvis ./cmd/warvis` → success
  - `make -C harness/find-evil kill-switch-check` → 5/5 PASS (non-negotiable)
  - `python -m pytest` → 47/47 PASS (Python MCP untouched, should be green)
- **Validation**: All gates green before proceeding
- **Risk**: HIGH (kill-switch must hold; Python MCP must stay intact)
- **Est. Time**: 20 min (most time on kill-switch validation)

### M4: Optional documentation (if energy permits)
- **Task**: Add 1-line note in `docs/find-evil/architecture.md` §15 (Audit & Persistence)
  - Example: "§15.1 gemma_response events now include tool_name, arguments, reason, and current_state for forensic decision reconstruction."
- **Validation**: Documentation matches code change
- **Risk**: LOW (purely informational)
- **Est. Time**: 5 min

## Stop-and-Fix Rule

**If any milestone validation fails:**
1. Do NOT proceed to next milestone
2. Analyze root cause:
   - Test failures → check Action parsing, FSM interface
   - Build failures → check Go syntax, import paths
   - Kill-switch failures → REVERT and escalate (non-negotiable)
3. Fix in place, re-validate, continue

**Kill-switch policy:**
- If `make kill-switch-check` fails, STOP immediately and revert all changes
- Do not attempt to fix kill-switch separately; it gates the entire UoW

## Verification Gates (In Order)

1. **M1 validation (RED test)**: `go test ./... -run TestExpandedGemmaResponseAudit -v` → FAIL
2. **M2 validation (GREEN test)**: `go test ./... -run TestExpandedGemmaResponseAudit -v` → PASS
3. **M3a (Go full suite)**: `go test ./...` → all PASS
4. **M3b (Go build)**: `go build -o bin/warvis ./cmd/warvis` → success
5. **M3c (Kill-switch)**: `make -C harness/find-evil kill-switch-check` → 5/5 PASS
6. **M3d (Python regression)**: `python -m pytest` → 47/47 PASS
7. **M4 (optional docs)**: `grep "decision reconstruction\|tool_name\|arguments" docs/find-evil/architecture.md`

## Done When

- [ ] M1 test written, fails as expected
- [ ] M2 code implemented, test passes
- [ ] M3a: `go test ./...` → all PASS
- [ ] M3b: `go build` → success
- [ ] M3c: kill-switch → 5/5 PASS (non-negotiable)
- [ ] M3d: Python pytest → 47/47 PASS (no regression)
- [ ] M4: Optional doc note added (if time)
- [ ] devos_verify_dev_session → PASS

## Status

| M | Task | State | Evidence |
|---|------|-------|----------|
| 1 | Write RED test | waiting | - |
| 2 | Implement GREEN | waiting | - |
| 3 | Integration gates | waiting | - |
| 4 | Doc note (opt) | waiting | - |

## Additional Notes

**Audit Schema Backward Compatibility:**
- JSONL appends only (no breaking change to existing lines)
- New fields are additive; old audit logs remain valid
- Forensic tools can safely ignore new fields or add them to schema iteratively

**FSM Access Pattern:**
- `l.fsm` is guaranteed non-nil in callOllama() (set in NewLoop constructor)
- `CurrentState()` uses RLock internally; safe for concurrent read
- `CurrentState().Name()` returns enum string ("INITIALIZE", "TRACE", "SCAN", "EXPOSE", "LOCK")

**Risk Mitigation:**
- All changes isolated to audit event logging (non-blocking append)
- No change to action parsing, tool execution, or FSM logic
- Kill-switch validates entire hunt protocol end-to-end

---

**Status**: shipped (2026-05-08, Go agent tests PASS, kill-switch 5/5 PASS, python pytest 47/47 PASS, lesson_id=ab387a91-3dc7-408e-aac5-8ca1cf1f292f)
**Risk Level**: MEDIUM (mitigated — first Go change since Phase 3 lock, regression gate held)
