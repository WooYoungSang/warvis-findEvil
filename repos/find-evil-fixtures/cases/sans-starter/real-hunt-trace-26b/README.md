# Real Hunt Trace — base-wkstn-05 (Gemma 4 26B variant)

Sibling trace to `../real-hunt-trace/` (which used `gemma4:e4b`, the 8B
variant). Same evidence, same MCP server, same Go bridge — different LLM.
The point of having both is to show how reasoning quality scales with
model size on the *same* missing-context failure mode.

## Configuration

| Field | Value |
|-------|-------|
| Date (UTC) | 2026-05-09 17:00 |
| Evidence | `/evidence/sans-starter/base-wkstn-05-memory.img` |
| LLM | **`gemma4:26b-a4b-it-q4_K_M`** (25.8B params, Q4_K_M, 19.5 GB VRAM) |
| Hardware | RTX 4090 (24 GB VRAM) |
| HTTP timeout | 600s (raised from 120s in pkg/ollama/client.go for 26B-class models) |
| Pre-warm | model loaded into VRAM before hunt (`/api/generate` 1-token call, 9.5s total) |
| Budget | `WARVIS_MAX_TURNS=8` |

## What 26B did

1. `case_opened` — same as 8B trace.
2. `state_transition` INITIALIZE → TRACE.
3. **Turn 1** `gemma_response`: chose `timeline.build`. tool_called →
   tool_result success.
4. **Turn 2** `gemma_response`: instead of fabricating a case_id (as 8B
   did), 26B returned `action: "escalate"` with reason:

   > *"Missing critical context: The 'case_id' parameter is required for
   > all tool operations, but no case_id has been provided in the initial
   > context or environment. I cannot proceed with investigation without
   > a valid identifier for the [scope]."*

5. `loop_error` (escalation propagated as agent loop exit).

## Why this is stronger evidence than the 8B trace

8B (e4b) **hallucinated a UUID** (`Case_ALPHA_789`) when it noticed
case_id was missing. It tried 3 times with the fabricated value before
the loop timed out.

26B **detected the missing context and escalated honestly**. This is the
behavior of a senior analyst — the SANS judging rubric explicitly compares
autonomous execution to a senior analyst (criterion #1). Refusing to
proceed when context is insufficient is a *strength*, not a failure.

The contrast across the two traces is itself evidence:
- Same root-cause prompt issue (case_id not threaded through history).
- 8B: fabricates → retries → times out.
- 26B: detects → escalates → exits cleanly.

The Hunt FSM treats `escalate` as a first-class action (see
`warvis/internal/agent/loop.go`), so the escalation event is recorded in
the audit trail without crashing the loop. Hash chain integrity preserved.

## Honest caveats (same as e4b trace)

- **No SCAN state.** Both traces stopped in TRACE. Vol3's ability to
  parse `base-wkstn-05-memory.img` is anchored separately by the B2
  smoke test (`vol windows.info` → SystemTime equal to dc3dd capture log
  to the second).
- **No findings.** Neither trace enumerates IOCs / processes / malfind.
- **No precision/recall.** SANS does not publish ground truth labels.
- **Underlying prompt limitation.** Both traces failed for the same
  reason (case_id not in agent conversation history); B3 demonstrates
  the *infrastructure* end-to-end correctness, the prompt fix is a
  follow-up UoW.

## Files

| File | Description |
|------|-------------|
| `audit.jsonl` | 7-line hash-chained audit log |
| `state.json` | Final FSM state (TRACE) |
| `hunt-stdout.json` | warvis CLI stdout |
| `hunt-stderr.log` | Agent loop stderr including escalation message |
