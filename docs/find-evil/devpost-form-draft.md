# Devpost Form Draft — W.A.R.V.I.S. Find Evil

Paste-ready draft for the SANS FIND EVIL Devpost submission. Source material: `docs/find-evil/devpost-page.md`, `docs/find-evil/valhuntir-comparison.md` §5, and the frozen Devpost rules snapshot in `plans/ITEM-212-find-evil/evidence/devpost-20260429T225233Z.md`.

## Project name

W.A.R.V.I.S. — Find Evil

## Tagline

Smallest IR agent whose architecture enforces its forensic role

Note: ≤ 80 chars. This trims the canonical Valhuntir-aware positioning statement while preserving the core claim.

## Short description

W.A.R.V.I.S. is a local forensic IR agent: one Go binary, a 5-state FSM, MCP forensic tools, Gemma 4 via Ollama, reproducible JSONL audit logs, and CI kill-switch checks.

Note: ≤ 200 chars.

## Long description

Digital forensics teams need faster triage without giving an autonomous agent unrestricted shell access. W.A.R.V.I.S. asks a narrow question: what is the smallest possible IR agent whose architecture — not its prompt — guarantees it cannot escape its forensic role?

W.A.R.V.I.S. answers with a single Go orchestrator that drives a deterministic Hunt FSM:

1. **INITIALIZE** opens a case and hashes evidence.
2. **TRACE** builds event timelines and queries structured logs.
3. **SCAN** hunts for indicators of compromise and memory artifacts.
4. **EXPOSE** cross-checks findings before reporting.
5. **LOCK** seals the audit trail and finalizes findings.

Each state declares its allowed tools at compile time, so the agent physically cannot call SCAN tools while it is in TRACE, or reporting tools before LOCK. Gemma 4 through Ollama chooses the next investigation action, but the binary enforces the boundary. Every LLM decision, tool call, tool result, and state transition is written to immutable `audit.jsonl` for review.

The project is intentionally honest about scope. Valhuntir is the comprehensive DFIR reference with far broader tool coverage, RAG depth, and detection-rule volume. W.A.R.V.I.S. does not try to outclass that breadth. It focuses on airgapped or regulated environments where a reviewer needs a deployable local binary, deterministic state boundaries, resume without re-burning LLM budget, and reproducible kill-switch tests.

Current evidence includes synthetic recall measurement, real SANS starter-case traces, structured audit logs, and deterministic mock traces that exercise the full INITIALIZE→TRACE→SCAN→EXPOSE→LOCK path. The demo video will show the agent self-correcting by escalating instead of fabricating when critical context is missing, then inspect the audit trail that proves the decision.

## How it's made

- **Go orchestrator**: `warvis/cmd/warvis` owns the Hunt FSM, state transitions, tool whitelist enforcement, pause/resume, and the user-facing `warvis hunt` command.
- **Python MCP server**: `src/find_evil_mcp/` exposes forensic tools for case opening, timeline building, log queries, IOC scanning, memory scanning, network scanning, cross-checking, and report appends.
- **Local LLM path**: Gemma 4 runs through Ollama at `localhost:29134`, so the agent can operate without cloud API egress after the model is available.
- **Audit and evidence integrity**: each hunt writes JSONL events for LLM responses, tool calls, tool results, state changes, and hash-linked forensic records.
- **Verification harness**: pytest covers documentation and evidence invariants; Go tests cover orchestrator packages; `make -C harness/find-evil kill-switch-check` proves the agent reaches controlled terminality and preserves structured audit output.
- **Hackathon integration points**: SANS starter evidence is represented under `repos/find-evil-fixtures/cases/sans-starter/`; SIFT Workstation cold-start validation remains the required deployment proof before final submission.

## Built with

- Go 1.21+
- Python 3.10+
- Model Context Protocol (MCP)
- Gemma 4
- Ollama
- JSON-RPC 2.0
- JSONL audit logs
- pytest
- ruff
- Go test
- SANS SIFT Workstation target

Suggested Devpost tags: `forensics`, `dfir`, `mcp`, `gemma`, `ollama`, `finite-state-machine`, `sans-find-evil`, `incident-response`, `audit-logs`.

## Try it out link

https://github.com/WooYoungSang/warvis-findEvil#quick-start

Use the README quickstart on GitHub for now. No live hosted demo URL is in scope because W.A.R.V.I.S. is designed for local forensic execution over case evidence.

## Video URL

TODO: Add unlisted YouTube URL after Task 4 demo recording.

Recording requirement reminder: ≤5 minutes, show live terminal execution and agent self-correction.

## Image gallery thumbnail spec

1280×640 thumbnail concept: left side shows the 5-state Hunt FSM (INITIALIZE → TRACE → SCAN → EXPOSE → LOCK) with compile-time whitelist shields around each state; right side shows a terminal with `warvis hunt /evidence/sans-starter/base-wkstn-05-memory.img` and an `audit.jsonl` excerpt highlighting `action=escalate` as self-correction. Use the tagline: “Architecture, not prompt discipline.”

## GitHub repo

https://github.com/WooYoungSang/warvis-findEvil

License: MIT.

## Team members + W-8BEN reminder

Team member: WoopsFactory / WooYoungSang.

Administrative reminder: Korean residents are eligible per the captured rules snapshot. If selected for a prize, prepare the non-US winner tax paperwork path, including W-8BEN where required by the sponsor/payment processor.
