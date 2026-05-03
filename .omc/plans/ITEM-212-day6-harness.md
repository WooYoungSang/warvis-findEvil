# Harness Config: ITEM-212-day6-scan-status

**Date**: 2026-05-06 (Day 6)  
**Status**: READY FOR DIRECT IMPLEMENTATION  
**Risk Level**: MEDIUM (Go bridge, 4 days to kill-switch deadline 2026-05-09)

---

## Always-on

- **CLAUDE.md**: exists (encyclopedia style, 80+ lines) with clear agent delegation table
- **Local Agents**: 10 agents found (warvis-orchestrator, warvis-initiator, warvis-verifier, python-generic-engineer, etc.)
- **Assessment**: Well-structured, agents defined for domain-specific work. No dead weight identified for this UoW.

---

## On-demand

**Skills Available**:
- pattern-extract, forge, spec-gap-check, bet-kickoff, ship-or-cut, build-report
- oh-my-claudecode: ultrawork, ralph, team, autopilot, etc.

**Reuse**: None — this UoW is Phase 3 Go bridge implementation (not matching prior patterns).  
**New Skill Candidate**: SCAN state orchestration pattern (iocs.scan + memory.* + net.* coordination) may be reusable post-hackathon.

---

## Deterministic

**Test Commands**:
- `go test ./... -short` (warvis/ dir, 30+ tests, all currently PASSING)
- `make -C harness/find-evil kill-switch-check-1` (existing, PASSING)
- `make -C harness/find-evil kill-switch-check-2` (existing, PASSING)
- `make -C harvis/find-evil kill-switch-check-3` (existing, PASSING)

**Lint Commands**:
- `go vet ./...` (warvis/)
- `golangci-lint run` (warvis/)

**Eval Harness**: Exists (kill-switch-check-{1,2,3} targets in Makefile; Tests 4–5 to be added).

**Gate Strategy**: test → lint → kill-switch-checks (1-5 sequential)

---

## Measurable

**Baseline Eval**: 
- kill-switch-check-1: PASSING (INITIALIZE→TRACE transition)
- kill-switch-check-2: PASSING (Tool calls in audit.jsonl)
- kill-switch-check-3: PASSING (Gemma JSON parsing)
- kill-switch-check-4: NOT YET (audit.jsonl JSONL validation)
- kill-switch-check-5: NOT YET (warvis status JSON output)

**Observability**: Structured JSONL logging in /cases/<case_id>/audit.jsonl (working).

**Trace Plan**: devos_record_evidence after each milestone completion.

---

## Collaboration Map

**Implementation**: DIRECT (no sub-agents)  
- Rationale: UoW scope is well-defined, 6 milestones, <10 files modified, orchestrator skill sufficient.

**Testing**: DIRECT  
**Security**: Gate 3 pattern scan (always) — no critical secrets exposure risk.

---

## HITL Points

**HIGH Risk Actions**:
- Modify internal/hunt/fsm.go (core FSM logic) — verify tests pass after each edit
- Add/modify kill-switch-check-{4,5} Makefile targets — validation essential

**CRITICAL Actions**: None identified for this UoW.

**Approval**: Use `Option B (direct implementation)` — no user approval needed between milestones; only final verification gate.

---

## Dead Weight Check

**CLAUDE.md Agent Catalog**: Over-specified for this UoW (warvis-planner, warvis-maker, warvis-verifier, warvis-finisher not needed — orchestrator handles all). No action required (agents useful for future UoWs); keep as-is.

---

## Summary

- **Harness Quality**: Excellent (encyclopedic CLAUDE.md, 10 agents, 3 kill-switch tests passing).
- **Execution Path**: Direct (no delegation).
- **Expected Blockers**: Go/Ollama environment setup; devos MCP connectivity (optional, best-effort).
- **Time Budget**: ~6–8 hours (M6a–M6f, TDD cycle per milestone).
