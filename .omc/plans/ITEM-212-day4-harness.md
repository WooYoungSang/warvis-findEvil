# Harness Config: ITEM-212-day4-ollama-agent

## Always-on
- claude_md: exists(encyclopedia — 82 lines, 6 procedural sections)
- local_agents: 15 agents discovered (.claude/agents/)
  - Primary for Day 4: warvis-maker (implementation), warvis-verifier (gates)
- Warning: CLAUDE.md is focused (not over-detailed) — safe for this work

## On-demand
- skills_available: forge, bet-kickoff, build-report, pattern-extract, ship-or-cut, spec-gap-check
- reuse: None (Ollama/Gemma 4 integration is novel)
- new_skill_candidate: "ollama-agent-loop" (if this pattern repeats in later phases)

## Deterministic
- test_cmd: `cd warvis && go test ./internal/agent/... -v`
- lint_cmd: `cd warvis && go vet ./... && golangci-lint run ./internal/agent/...`
- eval_harness: exists (make kill-switch-check targets)
- gate_strategy: lint → test → security → kill-switch-check-1

## Measurable
- baseline_eval: kill-switch-check-1 (PASSED in Day 3)
- observability: structured logging (audit.jsonl, state.json)
- trace_plan: devos_record_evidence per milestone

## Collaboration Map
- impl: direct (small 5-file module, TDD-friendly)
- test: direct (unit tests via go test)
- security: pattern scan (no secrets, no unsafe)

## HITL Points
- HIGH risk: None (pure Go stdlib, no external deps beyond Ollama HTTP)
- CRITICAL: None

## Dead Weight Check
- CLAUDE.md agent delegation table is comprehensive — reuse warvis-maker for TDD
- No unnecessary rules detected

---

## Implementation Strategy

**Impl approach**: Direct TDD (red→green→refactor) for 5 files:
1. `pkg/ollama/client.go` — Ollama HTTP client (M4a)
2. `internal/agent/prompt.go` — System prompt builder (M4b)
3. `internal/agent/tool_call.go` — Action parsing + 3-retry (M4c)
4. `internal/agent/loop.go` — Agent loop skeleton + dispatch (M4d, M4e, M4f)
5. Unit tests for each

**Test cmd per milestone**:
- M4a: `go test ./pkg/ollama -run TestClient` (mock HTTP)
- M4b: `go test ./internal/agent -run TestPrompt` (schema validation)
- M4c: `go test ./internal/agent -run TestParse` (3 retry scenarios)
- M4d–M4f: `go test ./internal/agent/...` (full loop integration)

**Verification gates**:
1. `golangci-lint run ./pkg/ollama ./internal/agent` (lint)
2. `go test ./internal/agent/... -v -cover` (unit tests, 60%+ coverage)
3. Security: grep for unsafe patterns (none expected)
4. Ollama handshake: `ollama list` (verify gemma4:26b available)

**Kill-switch alignment**:
- M4 is prerequisite for M5 (TRACE agent loop)
- M5 gates kill-switch-check-2 & 3 (tool-call JSON parsing + autonomy)
- M4 itself is not gated (internal scaffolding)
