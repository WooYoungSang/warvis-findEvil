# Comprehensive Mock Hunt Trace — Full FSM Traversal

Third sibling trace (alongside `../real-hunt-trace/` with gemma4:e4b and
`../real-hunt-trace-26b/` with gemma4:26b). Captured by the
`prompt-context-threading` Bet to deterministically prove the agent
loop's state-advancement infrastructure end-to-end.

## Configuration

| Field | Value |
|-------|-------|
| Date (UTC) | 2026-05-09 10:10 |
| Evidence | `/evidence/sans-starter/base-wkstn-05-memory.img` |
| LLM | **mock** (`harness/find-evil/mock_ollama_server.py`, deterministic) |
| MCP server | `python -m find_evil_mcp.server` (Phase 1+2 + B3 patches) |
| Forensic backend | volatility3 2.28.0 (B2) — see "What this trace does NOT prove" |
| Budget | `WARVIS_MAX_TURNS=10` |

## What this trace proves (Bet lock-in conditions L1, L2, L3, L5)

22-line audit.jsonl with the **complete Hunt FSM traversal**:

```
INITIALIZE
   │ case_opened
   ▼
TRACE
   │ timeline.build  (tool_called, tool_result success=true)
   │ log.query       (tool_called, tool_result success=true)
   │ state_complete  → Transition()
   ▼
SCAN
   │ memory.process_list  (tool_called, tool_result success=true)
   │ memory.malfind       (tool_called, tool_result success=true)
   │ state_complete  → Transition()
   ▼
EXPOSE
   │ state_complete  → Transition()
   ▼
LOCK   ← state.json terminal value
   │ state_complete attempt → Cannot advance from LOCK (terminal)
   ▼ loop_completed
```

This trace **deterministically demonstrates**:

- **L1 — case_id threading**: BuildSystemPrompt embeds `case_id:
  dd1950af-…` in the system prompt for every turn (the Bet's primary fix).
- **L2 — TRACE → SCAN transition recorded in audit.jsonl**: the
  `state_transition` event with `from=TRACE, to=SCAN` is at line 10. The
  same trace also records SCAN → EXPOSE and EXPOSE → LOCK.
- **L3 — memory.\* tool_called events in SCAN state**: both
  `memory.process_list` and `memory.malfind` are recorded with
  `current_state=SCAN`.
- **L5 — hash chain integrity**: every `entry_hash` chains to its
  `prior_hash` across 22 events; clean termination via `loop_completed`.

## What this trace does NOT prove (Bet lock-in L4 — partial)

The `tool_result` events for `memory.process_list` and `memory.malfind`
both show `success: true`, **but the implied vol3 subprocess invocation
did not actually run**. Evidence of non-invocation:

- All five SCAN-state events share the wall-clock timestamp `10:10:11Z`.
  `vol -f 3GiB.img windows.pslist.PsList` takes 30–60 s on this hardware
  (B2 smoke test data); five events in one wall-clock second is
  impossible if vol3 fires.
- `memory_process_list.py` and `memory_malfind.py` (Phase 1+2 locked)
  short-circuit to an empty result when `case_id` is not threaded
  through to the tool's argument map (the schema-validated argument
  bag). Mock Ollama emits hard-coded action arguments without case_id
  awareness.

So the Bet's L4 ("`tool_used="volatility3"` appears in audit") is
**unsatisfied** for this comprehensive-mock trace. The infrastructure
*from* warvis CLI *through* find_evil_mcp *to* memory.* tool dispatch
is fully exercised; the **last hop into vol3 subprocess** is the
remaining gap, attributable to mock vs. real-Gemma case_id propagation
into tool arguments — not to the prompt-context-threading change itself.

The B2 smoke test (`smoke-vol-windows-info.txt`) anchors that vol3
*does* parse this exact image when it is invoked — only the agent-loop
→ vol3 last hop has not been demonstrated end-to-end in a single trace.

## Files

| File | Description |
|------|-------------|
| `audit.jsonl` | 22-line hash-chained audit log of the full FSM traversal |
| `state.json` | Final FSM state: `LOCK` |
| `hunt-stdout.json` | warvis CLI stdout |
| `hunt-stderr.log` | Agent loop stderr including "Advanced TRACE -> SCAN" lines |

## Reproduction

```bash
# 1. /evidence/ symlink (one-time, sudo)
sudo mkdir -p /evidence
sudo ln -snf /mnt/disk1/INCUBATOR/warvis-findEvil-fixtures/sans-starter \
            /evidence/sans-starter
sudo chown -h "$USER:$USER" /evidence/sans-starter

# 2. Mock Ollama
fuser -k 19134/tcp 2>/dev/null
python3 harness/find-evil/mock_ollama_server.py 19134 >/dev/null 2>&1 &
MOCK_PID=$!

# 3. Hunt with deterministic mock
mkdir -p /tmp/warvis-comprehensive
FIND_EVIL_CASES_ROOT=/tmp/warvis-comprehensive \
FIND_EVIL_SERVER_CMD=".venv-find-evil/bin/python -m find_evil_mcp.server" \
OLLAMA_URL=http://localhost:19134 \
OLLAMA_MODEL=mock \
WARVIS_MAX_TURNS=10 \
WARVIS_HUNT_TIMEOUT_SECONDS=600 \
warvis/bin/warvis hunt /evidence/sans-starter/base-wkstn-05-memory.img

kill $MOCK_PID
```

The audit byte content is deterministic up to UUIDs and timestamps.
