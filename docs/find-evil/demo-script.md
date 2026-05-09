# W.A.R.V.I.S. FIND EVIL Demo Script (≤5 minutes)

**Goal:** show architectural constraints, agent self-correction, and a reproducible audit trail for the SANS FIND EVIL submission.

This is a recorded screencast plan, not a live-only demo. Use the already captured traces under `repos/find-evil-fixtures/cases/sans-starter/` so the video is deterministic and honest.

---

## Pre-flight

```bash
# Build binary used in the demo
( cd warvis && go build -o bin/warvis ./cmd/warvis )

# Optional: pre-warm the model before recording live snippets
ollama serve
ollama run gemma4:26b-a4b-it-q4_K_M 'ready'

# SANS starter evidence path expected by case.open contract
ls -lh /evidence/sans-starter/base-wkstn-05-memory.img
```

Required evidence paths for screen capture:

- 26B real-hunt trace: `repos/find-evil-fixtures/cases/sans-starter/real-hunt-trace-26b/audit.jsonl`
- Full deterministic FSM trace: `repos/find-evil-fixtures/cases/sans-starter/comprehensive-mock-trace/audit.jsonl`
- Accuracy caveats: `docs/find-evil/accuracy-report.md`
- Architecture diagram: `docs/find-evil/architecture.md` §11

---

## Timeline

### 0:00–0:15 — Problem statement

Narration:

> AI threats strike in minutes. W.A.R.V.I.S. is a local forensic IR agent whose architecture, not prompt discipline, prevents it from escaping its role.

Show:

```bash
head -20 docs/find-evil/devpost-form-draft.md
```

---

### 0:15–0:45 — Architecture constraint

Show the Mermaid/FSM section in `docs/find-evil/architecture.md` and explain:

- One Go orchestrator owns the Hunt FSM: `INITIALIZE → TRACE → SCAN → EXPOSE → LOCK`.
- Each state has compile-time allowed tools.
- Python MCP tools expose structured forensic actions, not a generic shell.
- Every LLM response, tool call, tool result, and transition is written to `audit.jsonl`.

Suggested command:

```bash
grep -n "INITIALIZE\|TRACE\|SCAN\|EXPOSE\|LOCK" docs/find-evil/architecture.md | head -20
```

---

### 0:45–2:00 — Real 26B trace: self-correction instead of fabrication

Show that the 26B run escalated when critical context was missing instead of inventing forensic data.

```bash
AUDIT=repos/find-evil-fixtures/cases/sans-starter/real-hunt-trace-26b/audit.jsonl
jq -c 'select(.event == "gemma_response") | {event, current_state, action_type, reason}' "$AUDIT"
```

Narration:

> This is the self-correction moment. The agent does not fabricate a case identifier or fake findings. It emits an escalation with the reason in the audit trail, preserving analyst trust.

Also show the SANS memory image target used for this lane:

```bash
ls -lh /evidence/sans-starter/base-wkstn-05-memory.img
```

---

### 2:00–3:30 — Deterministic full FSM traversal

Show the comprehensive mock trace exercising all five states and SCAN memory tools.

```bash
MOCK=repos/find-evil-fixtures/cases/sans-starter/comprehensive-mock-trace/audit.jsonl
jq -r 'select(.event == "state_transition") | [.timestamp, .from, .to, .reason] | @tsv' "$MOCK"
jq -r 'select(.event == "tool_called") | [.timestamp, .tool_name] | @tsv' "$MOCK"
```

Call out:

- `INITIALIZE → TRACE → SCAN → EXPOSE → LOCK`
- `memory.process_list` and `memory.malfind` are reached only in SCAN.
- This trace proves loop/FSM/MCP dispatch shape, while the accuracy report honestly documents the remaining Volatility3 last-hop caveat.

---

### 3:30–4:30 — Audit integrity and honest accuracy caveat

Show hash-chain and caveat docs:

```bash
python -m pytest tests/test_real_hunt_trace.py -q

grep -n "honest\|caveat\|vol3\|26B\|real hunt" docs/find-evil/accuracy-report.md | head -20
```

Narration:

> We do not claim production-grade IR accuracy without real-sample measurement. The submission separates measured synthetic recall, real SANS trace behavior, and known residuals.

---

### 4:30–5:00 — Positioning and close

Show Valhuntir-aware positioning and repo link:

```bash
grep -n "Valhuntir\|smallest possible IR agent\|GitHub" docs/find-evil/valhuntir-comparison.md docs/find-evil/devpost-form-draft.md | head -20
```

Narration:

> Valhuntir is the comprehensive DFIR reference. W.A.R.V.I.S. is deliberately smaller: a local binary, a strict FSM, MCP tool boundaries, and reproducible kill-switch evidence for airgapped or regulated response.

End on:

```text
https://github.com/WooYoungSang/warvis-findEvil
```

---

## Recording checklist

- Video length ≤ 5 minutes.
- Show `real-hunt-trace-26b` escalation/self-correction.
- Show `comprehensive-mock-trace` full FSM traversal.
- Show accuracy caveat; do not overclaim Volatility3 last-hop evidence.
- Save local recording as `docs/find-evil/demo.mp4` (gitignored).
- Upload unlisted video and paste URL into `docs/find-evil/devpost-form-draft.md`.

---

## Troubleshooting

| Error | Cause | Fix |
|---|---|---|
| `localhost:29134` refused | Ollama not running | Start `ollama serve` and pre-warm Gemma 4 |
| `/evidence/...` missing | SIFT evidence symlink/copy absent | Restore `/evidence/sans-starter/base-wkstn-05-memory.img` |
| Hunt pauses at TRACE | Model timeout or missing context | Use recorded 26B audit to show escalation as self-correction |
| Mock trace differs | Regenerated trace changed timestamps | Re-run trace tests and keep invariant-focused assertions |
