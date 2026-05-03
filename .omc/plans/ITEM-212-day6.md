# Plan: ITEM-212-day6-scan-status — SCAN + Kill Switch Tests 4–5

**UoW**: ITEM-212-day6-scan-status  
**Project**: warvis (Go bridge, Phase 3)  
**Scope**: Implement SCAN/EXPOSE/LOCK states, warvis status CLI, kill-switch-check-{4,5}  
**Timeline**: 2026-05-06 (Day 6), 6–8 hours  
**Kill Switch Deadline**: 2026-05-09 23:59 KST

---

## Harness

- **Config**: `.omc/plans/ITEM-212-day6-harness.md`
- **Test Cmd**: `cd warvis && go test ./... -short` (all must PASS)
- **Lint Cmd**: `cd warvis && go vet ./...`
- **Gate Strategy**: test → lint → kill-switch-check-{1,2,3,4,5}
- **Reused Skills**: None

---

## Scope

**Create**:
- ScanState implementation in `internal/hunt/fsm.go` (minimal logic, audit logging)
- ExposeState stub in `internal/hunt/fsm.go` (SCAN→EXPOSE transition)
- LockState stub in `internal/hunt/fsm.go` (EXPOSE→LOCK transition)
- AuditLog JSON-Lines validation logic in `internal/hunt/audit.go`
- `warvis status` subcommand in `cmd/warvis/main.go`
- kill-switch-check-4 and kill-switch-check-5 targets in `harness/find-evil/Makefile`

**Modify**:
- `internal/hunt/fsm.go` (add Scan/Expose/Lock methods, state registration)
- `internal/hunt/state.go` (already has ScanState/ExposeState/LockState structs)
- `cmd/warvis/main.go` (add status subcommand)
- `harness/find-evil/Makefile` (add test targets)

**Forbidden**:
- src/find_evil_mcp/ (Python MCP server) — ABSOLUTE NO-TOUCH
- External Go MCP SDK / Agent framework usage

---

## Milestones

### M6a: SCAN State Skeleton
**Task**: Implement ScanState in `internal/hunt/fsm.go`
- Add `func (f *FSM) ScanState(ctx context.Context) error` method
- Transition TRACE→SCAN, set state.CurrentState = "SCAN"
- Allow tool calls: iocs.scan, memory.process_list, memory.malfind, net.flow_summary
- Audit log: state_changed event with timestamp
- No agent loop logic (stub with immediate state_complete)

**Validation**: 
```bash
cd warvis && go test ./... -short -run TestScanState
```

**Risk**: MEDIUM (core FSM logic)

---

### M6b: EXPOSE + LOCK Stubs
**Task**: Add EXPOSE and LOCK state stubs
- Add `func (f *FSM) ExposeState(ctx context.Context) error` (SCAN→EXPOSE)
- Add `func (f *FSM) LockState(ctx context.Context) error` (EXPOSE→LOCK)
- Both: log state_changed, transition to next, return nil (no agent logic)

**Validation**:
```bash
cd warvis && go test ./... -short -run TestExposeState -run TestLockState
```

**Risk**: LOW (stubs only)

---

### M6c: JSONL Integrity Validator
**Task**: Add JSONL validation to `internal/hunt/audit.go`
- Implement `func (log *AuditLog) ValidateJSONLIntegrity() error`
- Read all lines from audit.jsonl, parse each as JSON
- Return error on malformed JSON, truncation, or corruption
- Used in kill-switch-check-4

**Validation**:
```bash
cd warvis && go test ./... -short -run TestValidateJSONL
```

**Risk**: LOW (utility function)

---

### M6d: warvis status CLI
**Task**: Add `warvis status <case_id>` subcommand
- Read `/cases/<case_id>/state.json`
- Output JSON: `{case_id, current_state, started_at, updated_at, llm_turns, tool_calls}`
- Used in kill-switch-check-5

**Implementation**:
```go
// cmd/warvis/main.go: status subcommand
func statusCmd(caseID string) {
    state := loadStateJSON(caseID) // read from FIND_EVIL_CASES_ROOT
    output := map[string]interface{}{
        "case_id": state.CaseID,
        "current_state": state.CurrentState,
        "started_at": state.StartedAt,
        "updated_at": state.UpdatedAt,
        "llm_turns": state.LLMTurns,
        "tool_calls": state.ToolCalls,
    }
    fmt.Println(mustMarshalJSON(output))
}
```

**Validation**:
```bash
cd warvis && go test ./... -short -run TestStatusCmd
FIND_EVIL_CASES_ROOT=/tmp/test warvis status <test-case-uuid> | jq -e '.current_state'
```

**Risk**: MEDIUM (new CLI surface, must match JSON contract)

---

### M6e: Kill-Switch Test 4
**Task**: Add `kill-switch-check-4` target to `harness/find-evil/Makefile`
- Run `warvis hunt /evidence/test.img`
- Extract case_id from output
- Run `jq . /cases/<case_id>/audit.jsonl` on ALL lines
- PASS if all lines valid JSON-Lines (no truncation/corruption)

**Implementation**:
```makefile
kill-switch-check-4: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 4: audit.jsonl JSONL integrity ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" $(WARVIS_BIN) hunt /evidence/test.img); \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
	  test -n "$$CASE_ID" || { echo "FAILED: no case_id"; exit 1; }; \
	  while IFS= read -r line; do \
	    echo "$$line" | jq . > /dev/null || { echo "FAILED: malformed JSON in audit.jsonl"; exit 1; }; \
	  done < $(CASES_TMP)/$$CASE_ID/audit.jsonl; \
	  echo "PASSED: all audit.jsonl lines are valid JSON"
```

**Validation**: `make -C harness/find-evil kill-switch-check-4`

**Risk**: MEDIUM (shell scripting)

---

### M6f: Kill-Switch Test 5
**Task**: Add `kill-switch-check-5` target to `harness/find-evil/Makefile`
- Run `warvis hunt /evidence/test.img`
- Extract case_id from output
- Run `warvis status <case_id>`
- Assert output contains "current_state" and its value is non-empty

**Implementation**:
```makefile
kill-switch-check-5: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 5: warvis status JSON output ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" $(WARVIS_BIN) hunt /evidence/test.img); \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
	  test -n "$$CASE_ID" || { echo "FAILED: no case_id"; exit 1; }; \
	  STATUS=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) $(WARVIS_BIN) status $$CASE_ID); \
	  echo "$$STATUS" | jq -e '.current_state' > /dev/null || { echo "FAILED: missing current_state in output"; exit 1; }; \
	  echo "PASSED: warvis status outputs valid JSON with current_state"
```

**Validation**: `make -C harness/find-evil kill-switch-check-5`

**Risk**: MEDIUM (new CLI integration)

---

## Done When

- [ ] M6a: ScanState compiles, tests PASS
- [ ] M6b: ExposeState + LockState stubs compile, tests PASS
- [ ] M6c: JSONL validator compiles, tests PASS
- [ ] M6d: warvis status CLI works, tests PASS
- [ ] M6e: kill-switch-check-4 Makefile target runs and PASSES
- [ ] M6f: kill-switch-check-5 Makefile target runs and PASSES
- [ ] `go test ./... -short` — all 30+ tests PASS
- [ ] `go vet ./...` — 0 errors
- [ ] `make -C harness/find-evil kill-switch-check` (tests 1–5) — all PASS
- [ ] No regression in Tests 1–3
- [ ] `/cases/<case_id>/audit.jsonl` valid JSON-Lines (Test 4)
- [ ] `warvis status <case_id>` outputs valid JSON with current_state (Test 5)

---

## Status Log

**2026-05-06 14:00 KST**: Plan created, harness evaluated, ready for M6a.
**2026-05-06 16:45 KST**: M6a–M6f implementation COMPLETE.
- M6a: SCAN state tools allowed (iocs.scan, memory.*, net.flow_summary) ✓
- M6b: EXPOSE + LOCK stubs (state transitions working) ✓
- M6c: JSONL validator (ValidateJSONLIntegrity) ✓
- M6d: warvis status CLI command ✓
- M6e: kill-switch-check-4 Makefile target ✓
- M6f: kill-switch-check-5 Makefile target ✓
- All 40+ unit tests PASS
- kill-switch-check-1 PASS (regression)
- kill-switch-check-4 PASS (JSONL validation)
- kill-switch-check-5 PASS (warvis status)
