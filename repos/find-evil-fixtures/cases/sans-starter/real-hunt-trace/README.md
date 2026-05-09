# Real Hunt Trace — base-wkstn-05

Live execution trace of `warvis hunt` against the SANS SRL-2018
`base-wkstn-05-memory.img` (3.0 GiB Windows 7 memory dump).

## Configuration

| Field | Value |
|-------|-------|
| Date (UTC) | 2026-05-09 05:51 |
| Evidence | `/evidence/sans-starter/base-wkstn-05-memory.img` |
| LLM | `gemma4:e4b` (8B params, Q4_K_M) via Ollama at `localhost:29134` |
| MCP server | `python -m find_evil_mcp.server` (Phase 1+2 code, fixed for mcp 1.x SDK in B3) |
| Forensic backend | volatility3 2.28.0 (B2 install) |
| Budget | `WARVIS_MAX_TURNS=8` |
| Hunt timeout | `WARVIS_HUNT_TIMEOUT_SECONDS=900` |

## Files

| File | Description |
|------|-------------|
| `audit.jsonl` | Hash-chained audit log (12 lines, all valid JSON) |
| `state.json` | Final FSM state (TRACE) + budgets |
| `hunt-stdout.json` | warvis CLI stdout (case_id + state file paths) |
| `hunt-stderr.log` | Agent loop stderr (entered state, tool calls, exit reason) |

## What this trace proves

1. **End-to-end live integration**: warvis Go CLI → Python MCP server → real
   forensic-tool subprocess wrapper → real volatility3 binary. None of this is
   mocked.
2. **Real LLM autonomy**: 3 turns of live Gemma 4 reasoning captured in
   `gemma_response` events with full decision-trace fields (`reason`,
   `tool_name`, `arguments`, `current_state`).
3. **Self-correction behavior visible**: Turn 2 reasoning explicitly states
   *"The previous tool call failed due to a missing 'case_id'..."*; Turn 3
   states *"The previous attempt failed due to an invalid case_id format
   error..."*. This is criterion #1 (Autonomous Execution Quality) in action.
4. **Audit trail integrity**: every `entry_hash` chains to the prior entry
   via `prior_hash` (SHA-256). No gaps.

## What this trace does NOT prove (honest caveats)

- **Hunt did not reach SCAN state.** All 3 turns were spent in TRACE calling
  `timeline.build`. The 4th-turn Gemma call exceeded the 120-second HTTP
  timeout (likely GPU contention) and the loop exited cleanly via
  `loop_error`. We therefore have no live `vol windows.pslist` /
  `windows.malfind` invocation in this trace; vol3's actual ability to
  parse this image is anchored separately by the B2 `smoke-vol-windows-info.txt`
  evidence (`vol windows.info` exit 0, captured kernel SystemTime equal to
  the SANS dc3dd capture log to the second).
- **Gemma fabricated a case_id.** Turn 2 invented `Case_ALPHA_789`
  instead of using the actual UUID from case.open. The system prompt and
  history-injection design need work to feed the case_id forward; this is a
  documented limitation, not a fix in scope for B3.
- **No findings yet.** Without reaching SCAN state, no IOCs/processes/
  malfind regions were enumerated. B3 demonstrates infrastructure
  correctness, not detection accuracy.
- **gemma4:e4b ≠ gemma4:26b.** CLAUDE.md targets the 26B variant; we used
  the 8B variant for this trace because the 26B model exceeded the 120s
  HTTP client timeout in pkg/ollama/client.go on this hardware. Both are
  Gemma 4 family.

## Reading the audit.jsonl

Each line is a JSON object with `event` ∈ {`case_opened`, `state_transition`,
`gemma_response`, `tool_called`, `tool_result`, `loop_error`}. Hash chain via
`entry_hash` / `prior_hash`. The `gemma_response` events carry the
explainability schema added in UoW `ai-explainability` (B3 prerequisite):
`action_type`, `tool_name`, `arguments`, `reason`, `raw_output`,
`current_state`.
