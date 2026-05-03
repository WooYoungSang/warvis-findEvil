# Plan: ITEM-212-day7-integration — Go Bridge Integration Testing & Hardening

## Harness
- config: .omc/plans/ITEM-212-day7-harness.md
- test_cmd: `cd warvis && go test ./... -short`
- lint_cmd: `cd warvis && go vet ./...`
- gate_strategy: [lint → test-short → kill-switch-check (integrated)]

## Scope
- **Create**: warvis/internal/integration/integration_test.go, timeout config, resume logic
- **Modify**: cmd/warvis/main.go (error handling, resume flags), harness/find-evil/Makefile (kill-switch-check unified target)
- **Verify**: all 5 kill-switch tests pass, no panics, audit trail valid
- **Don't touch**: src/find_evil_mcp/ (absolute constraint)

## Milestones

### M7a: Integration Test Suite
**Tasks**:
1. Create `warvis/internal/integration/integration_test.go` with full hunt flow test
2. Mock MCP server (case.open → case_id)
3. Mock Ollama (tool-call JSON responses)
4. Verify state.json written, audit.jsonl valid JSONL
5. Run: `cd warvis && go test ./internal/integration/... -short`

**Validation**: Integration test passes (INITIALIZE → TRACE with ≥2 mocked tool calls)
**Risk**: LOW

---

### M7b: Error Handling
**Tasks**:
1. Update `cmd/warvis/main.go` hunt command handler
2. Add file existence check (evidence path) → exit 1 with message
3. Add MCP server spawn error handling → logged + user message
4. Add Ollama connection error (http.Client timeout) → logged, NO panic
5. All errors use consistent format: "error: <message>"

**Validation**: `go test ./cmd/warvis/... -short` passes error handling tests
**Risk**: LOW

---

### M7c: Timeout Enforcement
**Tasks**:
1. Create `internal/config/config.go` with timeout constants:
   - DefaultToolTimeout = 120s
   - DefaultStateTimeout = 600s
   - DefaultHuntTimeout = 1800s
2. Update `internal/hunt/fsm.go`: wrap each tool call in context.WithTimeout(ctx, DefaultToolTimeout)
3. Update `internal/hunt/fsm.go`: wrap state duration in context.WithTimeout(ctx, DefaultStateTimeout)
4. Update `cmd/warvis/main.go`: wrap hunt in context.WithTimeout(ctx, DefaultHuntTimeout)
5. Verify context cancellation propagates (no goroutine leaks)

**Validation**: Timeout test passes (mock slow tool call, verify exit after 120s)
**Risk**: MEDIUM (ensure all goroutines respect context.Done())

---

### M7d: Resume Capability
**Tasks**:
1. Add `--case-id <uuid>` and `--phase <STATE>` flags to `cmd/warvis/main.go hunt` command
2. If `--case-id` provided:
   - Load `/cases/<case_id>/state.json`
   - Load `/cases/<case_id>/audit.jsonl` for history
   - If `--phase` provided, use it; else use state.json's current_state
3. Skip case.open (already done), resume agent loop from current state
4. Restore conversation history from audit.jsonl (last 20 exchanges)
5. Test: Resume from TRACE after ≥2 tool calls

**Validation**: Resume test passes (load prior state, budget counters preserved, history restored)
**Risk**: MEDIUM (corruption risk if state.json malformed)

---

### M7e: Makefile Integration Target
**Tasks**:
1. Update `harness/find-evil/Makefile`: add unified `kill-switch-check` target
2. Target runs: kill-switch-check-1 → kill-switch-check-2 → ... → kill-switch-check-5
3. Fail-fast: exit 1 if any test fails
4. Output: "✓ All 5 kill-switch tests PASSED" on success

**Validation**: `make -C harness/find-evil kill-switch-check` succeeds (all 5 tests green)
**Risk**: LOW

---

### M7f: Verification + Sign-off
**Tasks**:
1. Create `plans/ITEM-212-find-evil/verify-phase3.md` with:
   - Kill-switch test results (PASS/FAIL each)
   - External dependency notes (Test 2-3 require Ollama + Gemma)
   - Blocker list (if any)
   - Phase 4 handover condition: "Ready for demo + accuracy report"
2. Run full test suite: `cd warvis && go test ./... -short`
3. Run full integration: `make -C harness/find-evil kill-switch-check`
4. Final audit: check state.json + audit.jsonl format

**Validation**: verify-phase3.md written, all tests passing
**Risk**: LOW

---

## Done When
- [ ] All milestones M7a-M7f completed
- [ ] `cd warvis && go test ./... -short` passes (0 failures)
- [ ] `make -C harness/find-evil kill-switch-check` passes (all 5 tests green)
- [ ] No panics or crashes
- [ ] verify-phase3.md written with final verdict
- [ ] Go module ready for commit

---

## Implementation Strategy

**TDD Per Milestone**:
1. Write failing integration test (Red)
2. Implement minimum code (Green)
3. Refactor for clarity (Refactor)
4. Record evidence via devos_record_evidence

**Parallelizable**: M7b (error handling) and M7c (timeout) can run in parallel; M7d (resume) depends on M7c.

**Sequential**: M7a → (M7b, M7c) || M7d → M7e → M7f

---

Status: Ready for execution
Date: 2026-05-03 (Day 7 of 7)
Kill-switch deadline: 2026-05-09 23:59 KST
