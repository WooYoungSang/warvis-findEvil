# Plan: real-hunt-trace — Live warvis Hunt vs Real SANS Memory Dump (B3)

**UoW ID**: `real-hunt-trace` (B3 of 3-step split for P0 real-sample-integration)
**dev_session_id**: `ad9517db-8e52-49d1-b37b-2b30c675640f`
**Project**: warvis-findEval

## Context

B1 acquired the SANS `base-wkstn-05-memory.img` (3.0 GiB Windows 7 dump);
B2 installed volatility3 2.28.0 and verified parsing of that exact image
(SystemTime equality with the dc3dd capture log).

B3's goal: produce the first real-LLM, real-tool, real-evidence audit
trail for the warvis hunt pipeline. Foundation evidence for SANS judging
criteria #1 (Autonomous Execution Quality) + #5 (Audit Trail Quality).

## Scope

- M1: warvis CLI evidenceArgs heuristic — recognize `*memory*.img` /
  `*memory*.raw` as `mem_path` (1-line type addition + filename probe).
- M2: Go test + kill-switch regression gate (Phase 3 lock).
- M3: Run real hunt with `WARVIS_MAX_TURNS=8` budget. Real Gemma + real
  vol3 + real Phase 1+2 server.
- M4: Copy trace to `repos/.../sans-starter/real-hunt-trace/` (gitignore
  allow-list).
- M5: `accuracy-report.md` §8 Real Hunt Trace section (honest dual-path).
- M6: `tests/test_real_hunt_trace.py` + full regression.

## Out-of-spec issues discovered + resolved during B3

- **Phase 1+2 server.py never ran live**: `mcp` SDK API drift since 0.x.
  With explicit user approval, applied two surgical fixes
  (`stdio_server` context manager + `(content, structured)` tuple return
  in `call_tool`). pytest 56 pre-existing tests still pass; no regression.
- **`/evidence/` schema requires sudo symlink at filesystem root**:
  Phase 1+2 input schema enforces `^/evidence/...` pattern. Created
  one-time `/evidence/sans-starter` → `/mnt/disk1/INCUBATOR/.../sans-starter`
  symlink with sudo. No code change needed; preserves Phase 1+2 honesty.
- **Gemma 4 26B exceeded 120s HTTP timeout** in `pkg/ollama/client.go`
  on this hardware. Substituted same-family `gemma4:e4b` (8B Q4_K_M).
  Documented in trace README and accuracy-report.md.

## Verification Strategy

- pytest 75/75 PASS (13 new + 62 prior)
- Go test ./... PASS
- kill-switch 5/5 PASS (Phase 3 lock held)
- ruff clean
- audit.jsonl all valid JSON; hash chain continuous
- gemma_response events have non-empty `reason` field
- ≥ 1 tool_called + ≥ 1 tool_result
- accuracy-report.md mentions "Real Hunt Trace" + "base-wkstn-05" + caveats

## What B3 demonstrates

- Criterion #1: real Gemma 4 reasoning with explicit self-correction
  ('The previous tool call failed due to...').
- Criterion #4: per-state tool whitelist exercised on every turn.
- Criterion #5: hash-chained audit JSONL with explainability schema.

## What B3 does NOT measure (honest caveats)

- No SCAN state reached (timeout on turn 4) — vol3 was not invoked from
  the agent loop in this trace; B2 smoke anchors vol3 capability.
- Gemma hallucinated `case_id`; documented prompt limitation, not fixed.
- No precision/recall (SANS does not publish ground truth labels).
- LLM substitution: e4b not 26b on this hardware.

## Risk

**MED-HIGH executed; outcome LOW-MED**. Multiple surfaced bugs (mcp 1.x
drift, schema /evidence/ requirement, model timeout) all resolved within
B3 scope without violating Phase 3 kill-switch lock.

---

**Status**: shipped (2026-05-09, 75/75 tests PASS, kill-switch 5/5, real-hunt-trace audit.jsonl committed, lesson_id=b1ec131b-401e-43ed-840a-323b89b1f41c)
**Risk Level**: MED (mitigated — multiple discovered bugs surgically fixed with audit + tests)
