# ITEM-212 FIND EVIL — Spec (Phase 0 / R+1 inbound)

- **ITEM**: ITEM-212 FIND EVIL
- **Source card**: `output/handoff-cards/HANDOFF-212-find-evil.md` (warvis-sub, mission-control owned)
- **Evidence**: `plans/ITEM-212-find-evil/evidence/devpost-20260429T225233Z.md`
- **Today**: 2026-04-30
- **Submission deadline**: 2026-06-15 23:45 EDT (D-46)
- **Build window**: 2026-05-13 → 2026-06-15 (4.5 weeks)
- **Appetite**: **4 weeks** (D5 ship-or-cut at 2026-06-08)
- **Owner**: dev-node (this repo). Mission-control owns Devpost submission.

## 1. Problem Statement

SANS-organized **FIND EVIL!** hackathon asks contestants to extend **Protocol SIFT** with an agentic incident-response system that operates autonomously on forensic case data (disk images, memory captures, logs, network captures, remote endpoints). Submissions must demonstrate self-correction, accuracy validation, and architectural — not prompt-based — constraint enforcement.

The judging rubric (in priority order) rewards:

1. Autonomous execution quality (tiebreaker)
2. IR accuracy
3. Analytical breadth/depth
4. **Architectural constraints** > prompt guardrails
5. Audit trail
6. Usability/docs

## 2. Hypothesis (1-line)

> **An MCP-only tool surface is the cheapest path to defensible "architectural constraints" credit (criterion #4); winning depends on accuracy + autonomous self-correction (criteria #1–#2), which we get from a Claude Code orchestrator + curated SIFT toolset wrapped behind a narrow MCP server.**

## 3. Prize Reality Check (correction vs. card)

The $22,000 figure is the **total prize pool**, not an MCP-track-specific bucket:

| Place | Cash | Extras |
|---|---|---|
| 1 | $10,000 | SANS Summit pass + hotel + course + webcast |
| 2 | $7,500 | SANS Summit pass + hotel + course + presentation |
| 3 | $4,500 | course |

There is no separate MCP track. MCP is one of four allowed architectures. **Action**: open-question Q4 to mission-control to confirm KICK-740 framing.

## 4. Solution Sketch

```
            ┌───────────────────────────────────────────┐
            │  Claude Code (or OpenClaw) orchestrator   │
            │  — autonomous IR loop, self-correction    │
            └──────────────────┬────────────────────────┘
                               │ MCP (stdio + http)
            ┌──────────────────▼────────────────────────┐
            │  find-evil-mcp  (this build)              │
            │  Narrow tool surface — NO shell, NO eval  │
            │  Tools:                                   │
            │   • case.open(image|mem|pcap|logdir)      │
            │   • timeline.build(case_id)               │
            │   • iocs.scan(case_id, ruleset)           │
            │   • memory.process_list(case_id)          │
            │   • memory.malfind(case_id)               │
            │   • net.flow_summary(pcap_id)             │
            │   • log.query(case_id, q)                 │
            │   • report.append(finding)                │
            │   • verify.cross_check(finding_id)        │
            └──────────────────┬────────────────────────┘
                               │ subprocess (whitelisted)
            ┌──────────────────▼────────────────────────┐
            │  SIFT toolchain  (Volatility3, plaso,     │
            │  Zeek, Suricata, log2timeline, YARA, …)   │
            └───────────────────────────────────────────┘
```

Constraint design:

- MCP server **never** exposes `shell`, `exec`, `read_file`, `write_file` outside `/cases/<id>/report/`.
- Each tool returns **structured JSON**, never raw stdout — forces analytical narrative (criterion #3).
- `verify.cross_check` is the agent-side self-correction primitive (criterion #1).

## 5. Reuse Inventory & Reality Check

| Asset | Fit | Notes |
|---|---|---|
| `warvis-mcp` (devos_*) | **High** for build lifecycle, **None** for product surface. | Use for our own dev session; do NOT ship in submission. |
| `obsidian-mcp` | **None** for product surface. | Internal only. |
| SENTINEL-lite (Slither/Aderyn) | **Low** — Solidity static analysis is OOS for DFIR. | One stretch tool: `iocs.scan_smartcontract` for crypto-incident bonus. |
| Existing `.claude/agents/find-evil/*` (TBD) | n/a | This session creates them. |
| Korean residency | **OK** — eligible. | |

> **Honest note**: The original "재사용 자산" framing in the kickoff message is optimistic. Phase 0 must surface this gap so mission-control can adjust expectations.

## 6. Evaluation Criteria → Artifact Mapping

| # | Criterion | Our artifact |
|---|---|---|
| 1 | Autonomous execution quality | `verify.cross_check` MCP tool + agent retry loop + recorded trace |
| 2 | IR accuracy | Curated evil fixtures + accuracy report (precision/recall) |
| 3 | Analytical breadth/depth | Tool surface covers disk + memory + log + network |
| 4 | Architectural constraints | MCP server allow-list; capability handshake test |
| 5 | Audit trail | Structured JSON-Lines trace per case; signed hash chain on `report.append` |
| 6 | Usability/docs | Demo video, README, architecture diagram, dataset doc, accuracy report, deployment guide |

## 7. Non-Goals (scope wall)

- ❌ Re-implementing SIFT tools (we wrap, never replace).
- ❌ Generic shell/exec MCP tools — defeats criterion #4.
- ❌ Cloud deployment, multi-tenant, billing — submission is local-deploy.
- ❌ Smart-contract-only fixtures (DeFi audits are OOS for FIND EVIL).
- ❌ GPU inference (NFR: GPU=0).
- ❌ Auto-submit to Devpost (mission-control gate).
- ❌ Stripe/non-LemonSqueezy payments (n/a here, but NFR holds).

## 8. Kill Conditions (hard gates)

| Date | Gate | Trigger | Action |
|---|---|---|---|
| 2026-05-20 (D-26) | **Phase 1 ship**: MCP skeleton w/ `case.open` + `timeline.build` returns JSON on sample fixture | Fail → reduce scope to 4 tools max, drop `verify.cross_check` |  |
| 2026-06-01 (D-15) | **Phase 2 ship**: ≥6 tools live, recall ≥ 0.60 on internal fixtures | Fail → kill submission, redirect 4w to next-best ITEM |  |
| 2026-06-08 (D-7) | **Phase 3 ship-or-cut**: full submission package draft, recall ≥ 0.80, demo video shot | Fail → submit best-effort and accept Honorable-Mention odds; do NOT push to D-1 |  |
| 2026-06-14 (D-1) | mission-control review | Any blocking finding → defer to next hackathon |  |

## 9. Open Questions (mission-control)

- Q1: Team or solo entry? (max 5; partner availability?)
- Q2: SIFT Workstation VM hosting — local Docker proxy acceptable, or full OVA on dev machine?
- Q3: Confirm $22K = total pool understanding, not MCP-only.
- Q4: License choice MIT vs Apache 2.0 — repo policy preference?
- Q5: Acceptable to use Claude Code (our orchestrator) or must support OpenClaw too?

## 10. Phase Plan (4 weeks, post-Phase-0)

| Phase | Window | Output |
|---|---|---|
| Phase 0 | this session | spec + agents + harness + R+1 handover |
| Phase 1 | 2026-05-13 → 2026-05-20 | MCP skeleton (3 tools) + case fixture loader |
| Phase 2 | 2026-05-21 → 2026-06-01 | All 9 tools + accuracy fixtures + agent loop |
| Phase 3 | 2026-06-02 → 2026-06-08 | Demo video, accuracy report, architecture diagram, README |
| Phase 4 | 2026-06-09 → 2026-06-14 | Buffer + mission-control review + submission package |
| Submit | 2026-06-15 (mission-control) | Devpost upload (NOT this node) |
