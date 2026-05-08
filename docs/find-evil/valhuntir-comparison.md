# Valhuntir vs warvis-findEval — Differentiation Matrix

> Reference impl: https://github.com/AppliedIR/Valhuntir (Steve Anson, SANS instructor; v0.6.1, 323 commits, MIT)
> This file: /home/jang/Workspace/warvis-findEvil/docs/find-evil/valhuntir-comparison.md
> Purpose: precision differentiation against the reference. Foundation for narrative-pivot rewrite.

## 1. Executive Summary

**The gap is real.** Valhuntir is a multi-MCP platform with 90+ tools across 8 backends, 22K+-record forensic RAG, 3,700+ Hayabusa Sigma rules, a HMAC-signed examiner-portal workflow, and 7 months of compounding commits. As a SANS-instructor reference, it sets the de-facto judging baseline.

We cannot — and do not try to — match Valhuntir on breadth. Instead we focus on three narrow axes Valhuntir does not occupy:

1. **Single Go binary, no Python runtime, no Docker** — `warvis hunt` runs offline on a fresh SIFT VM in seconds.
2. **5-state Hunt FSM with per-state tool whitelist enforced in Go** — architecturally guaranteed self-restraint, not gateway-mediated discipline.
3. **Budget-preserving resume + kill-switch test harness** — every hunt is reproducible, interruptible, and verifiable in CI.

This is not a slogan. The matrix below is honest about where Valhuntir wins.

## 2. Architecture Compare

| Dimension | Valhuntir | warvis-findEval | Winner |
|-----------|-----------|-----------------|:------:|
| Orchestrator language | Python 3.10+ | Go (single static binary) | us (deployability) |
| MCP backends | 8 (sift-gateway:4508) | 1 (find-evil-mcp stdio) | Valhuntir (breadth) |
| Total forensic tools | 90+ | 9 | Valhuntir |
| Evidence parsers | 15 (EVTX/MFT/Registry/memdump/...) | lite-mode emulation | Valhuntir |
| Detection content | Hayabusa Sigma 3,700+ rules | none | Valhuntir |
| Forensic knowledge base | RAG 22K+ records | none | Valhuntir |
| Agent runtime | Claude Code / Desktop / Cherry / LibreChat | self-hosted Gemma 4 (Ollama 26B local) | mixed* |
| Audit | HMAC-signed ledger + 5 jsonl streams | audit.jsonl with hash chain | Valhuntir |
| Approval workflow | DRAFT → password-gated examiner approve | none (FSM lock-state instead) | Valhuntir on HITL |
| Deterministic state model | gateway-level discipline (prompt-shaped) | hardcoded 5-state FSM in Go (compile-time) | us (architectural) |
| Resume | unspecified | budget-preserving with state.json | us |
| Kill-switch test harness | unspecified | `make kill-switch-check` 5/5 PASS | us |
| Cold start | Python + 8 stdio subprocesses + gateway | single binary launch | us (offline) |
| API key required | Claude API in default path | none (Ollama local) | us (airgap-ready) |

\*Mixed: Valhuntir's choice maximizes intelligence (Claude); ours maximizes airgap-readiness.

## 3. Six-Criteria Comparison (SANS, equal weight 16.7% each)

| # | Criterion | Valhuntir | warvis-findEval | Edge |
|:-:|-----------|-----------|-----------------|:----:|
| 1 | Autonomous Execution Quality | DRAFT-approve examiner gate; Claude-grade reasoning | FSM-bounded budget + retry; Gemma 4 reasoning | Valhuntir |
| 2 | IR Accuracy | 22K-record RAG + 3,700 Sigma rules + 15 parsers | synthetic case-001 lite-mode 100%; real samples NOT measured | Valhuntir |
| 3 | Breadth and Depth of Analysis | 90+ tools, 8 backends, OpenCTI + OpenSearch | 9 tools, single backend | Valhuntir |
| 4 | Constraint Implementation | gateway-level discipline + DRAFT workflow | compile-time tool whitelist + 5-state FSM + sandbox per case | **us** |
| 5 | Audit Trail Quality | HMAC ledger + 5 jsonl + per-MCP audit | audit.jsonl + hash chain (no HMAC) | Valhuntir |
| 6 | Usability and Documentation | Examiner Portal + production docs + 7-month commit base | 8/8 Phase 4 docs + 5-min demo script + offline-runnable README | mixed |

**Reading the table honestly**: Valhuntir wins 4 of 6, ties 1 (criterion 6), and loses only criterion 4. Two-thirds of judging weight goes against us. Our path is therefore not "win the matrix" but "be the strongest narrow alternative".

## 4. What Valhuntir Doesn't Have — Three Defensible Differentiators

We selected these three from a candidate set of seven (offline-only, Go binary, FSM determinism, resume safety, kill-switch harness, sub-second cold start, hardcoded whitelist). Each is verifiable from public artifacts in this repo.

### 4.1 Single Go Binary, Offline, No Python Runtime

`warvis/bin/warvis` is a statically-linked Go binary (7.8 MB). On a fresh SIFT VM:
- No `pip install`. No Docker. No gateway.
- No outbound network in default path (Ollama is `localhost:29134`).
- Demonstrable advantage in airgapped IR (e.g., classified or regulated environments where Claude API egress is forbidden).

Valhuntir's default path requires Claude API egress; even the LibreChat option needs a hosted LLM endpoint. Our binary runs offline once Ollama is initialized locally. This is not a marketing claim—it is architectural.

### 4.2 5-State Hunt FSM — Compile-Time Tool Whitelist

In `warvis/internal/hunt/state.go`, each state declares its allowed tools as a Go literal:

| State | Allowed tools | LLM autonomy |
|-------|---------------|:------------:|
| INITIALIZE | case.open | 0% |
| TRACE | timeline.build, log.query | 70% |
| SCAN | iocs.scan, memory.{process_list, malfind}, net.flow_summary | 90% |
| EXPOSE | verify.cross_check, report.append | 60% |
| LOCK | (terminal) | 0% |

`IsToolAllowed(state, tool)` is enforced at every call site in `warvis/internal/agent/loop.go`. The agent **physically cannot** call SCAN tools while in TRACE state — this is criterion #4 (Constraint Implementation) at compile time.

Valhuntir's discipline is gateway-level (a Python service can be modified at runtime); ours is part of the binary you ship. This difference is meaningful for regulated environments where runtime code modification is forbidden.

### 4.3 Budget-Preserving Resume + Kill-Switch Test Harness

`warvis hunt --case-id <uuid> --phase TRACE` reloads `state.json` and the FSM with budgets **not reset**. A hunt interrupted by SIGTERM resumes without re-burning the LLM-turn quota.

`make kill-switch-check` runs 5 tests that each prove a specific terminality property:
1. INITIALIZE→TRACE transition completes
2. tool_called + tool_result events ≥ 1
3. gemma_response autonomy ≥ 1, tool_called ≥ 2
4. audit.jsonl all-lines valid JSON
5. `warvis status` outputs valid JSON

Result: 5/5 PASS as of 2026-05-07. This is reproducible CI evidence of agent terminality — Valhuntir's repo does not appear to expose an equivalent gate. The differentiation here is about verification: Valhuntir's quality depends on prompt discipline; ours includes a machine-checkable harness.

## 5. Honest Positioning Statement (for devpost-page)

warvis-findEval does not try to outclass Valhuntir on tool count, RAG depth, or detection-rule volume; we do not claim production-grade IR accuracy without real-sample measurement, and we acknowledge Valhuntir as the comprehensive reference. We focus instead on a strictly narrower question: *what is the smallest possible IR agent whose architecture — not its prompt — guarantees it cannot escape its forensic role?* Our answer is a single 7.8 MB Go binary, a 5-state FSM with compile-time tool whitelists, budget-preserving resume, and a kill-switch test harness that makes terminality reproducible in CI. For airgapped or regulated environments where Claude API egress is forbidden, this is a deployable answer; for open settings where breadth dominates, Valhuntir is the reference.

## 6. 38-Day Action Items (Derived from this Analysis)

| # | Action | Why (criterion impact) | Effort |
|:-:|--------|-----------------------|:------:|
| A1 | Integrate at least one real DFIR sample (SANS starter case data at https://sansorg.egnyte.com/fl/HhH7crTYT4JK) and produce a measured run with real Gemma 4 + at least one real binary tool (yara) | Lifts criterion #2/#3 from "lite-only" to "measured-on-real" | 5–7 d |
| A2 | Add `reasoning` field to gemma_response audit events (currently truncated 200 chars; expand to full JSON action object including reason) | Strengthens criterion #1 + #5 with explicit decision trace | 2–4 h |
| A3 | Record 5-minute demo video showing self-correction (e.g., verify.cross_check rejecting a finding and triggering re-scan) | Mandatory artifact #2 + criterion #1 evidence | 2–3 d |
| A4 | Pivot devpost-page narrative to "smallest IR agent whose architecture guarantees forensic role" using §5 statement; rewrite README opening to mirror | Differentiation against Valhuntir baseline | 1 d |
| A5 | Validate fresh SIFT VM end-to-end: clone, `make`, `warvis hunt`, audit inspection — record the trace as `evidence/sift-vm-cold-start.md` | Criterion #6 + project requirement "must run on SIFT Workstation" | 1–2 d |

## 7. Caveats — What This Document Is Not

- It is not a feature scorecard. SANS judges humans with a rubric, not a spreadsheet.
- It is not a marketing comparison. Valhuntir is honestly better on 4 of 6 criteria.
- It is not a full Valhuntir audit. We extracted public README facts, not source-code review.
- It is not stable. Valhuntir released v0.6.1 three weeks ago; the gap may widen.

The narrow-differentiation thesis (Section 1) is contingent on action A1 actually producing a measured-on-real trace. If it does not, our case for criterion #4 stands alone — and one out of six criteria is unlikely to win against the reference.
