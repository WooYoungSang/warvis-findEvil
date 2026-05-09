---
name: forge
description: Run the forge pipeline (Ignite → Blueprint → Hammer → Temper → Quench) for a UoW. The main session is the coordinator; each stage spawns one warvis-* specialist sub-agent. Sub-agents cannot spawn sub-agents in Claude Code, so orchestration MUST live in the main thread.
argument-hint: "<project_id> <uow_id> [obsidian=<vault_path>] [risk_authorized] [-- <inline spec>]"
level: 3
---

Forge a UoW into shipped form. **You (the main session) are the forge coordinator.** You spawn one warvis-* specialist sub-agent per stage in sequence and pass data forward. You do NOT write code, run tests, or modify files yourself except for the small Stage 0 harness/plan scaffolds and recovery actions explicitly listed below.

## Why main-thread orchestration

Claude Code does not allow sub-agents to spawn further sub-agents (https://code.claude.com/docs/en/sub-agents). Therefore the forge pipeline must run in the main session — there is no spawnable "orchestrator" intermediate. The legacy `warvis-orchestrator` agent is deprecated; redirect any caller to `/forge`.

## Arguments

Parse `{{PROMPT}}`:

```
<project_id>  <uow_id>  [obsidian=<vault_path>]  [risk_authorized]  [-- <uow_spec>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `project_id` | yes | Project slug |
| `uow_id` | yes | UoW ID |
| `obsidian=<path>` | no | Absolute vault root. Omitted → vault patch skipped (non-blocking). |
| `risk_authorized` | no | If present, HIGH risk auto-proceeds. CRITICAL is never auto-authorized. |
| `-- <uow_spec>` | no | Inline UoW spec used when `.omc/plans/<uow_id>.md` does not exist. |

**Spec resolution order:** `.omc/plans/<uow_id>.md` → inline `-- <spec>` → ask user then stop.

## Forge Metaphor

| Stage | Meaning | Specialist | Model | devos transition |
|---|---|---|---|---|
| 🔥 **Ignite** | 화로에 불을 피운다 — session start + baseline | `warvis-initiator` | haiku | `→ START` |
| 📐 **Blueprint** | 설계도를 그린다 — milestone breakdown + verification strategy | `warvis-planner` | **opus** | `START → PLAN` |
| 🔨 **Hammer** | 두드려 형태를 잡는다 — TDD red→green→refactor per milestone | `warvis-maker` | sonnet | `PLAN → IMPLEMENT` |
| ⚗️  **Temper** | 담금질로 강도를 본다 — lint / test / security gates | `warvis-verifier` | **opus** | `IMPLEMENT → VERIFY` |
| 💧 **Quench** | 식혀 굳힌다 — lesson capture + session end | `warvis-finisher` | haiku | `VERIFY → END` |

The main session is the smith holding the tongs. Specialist `model:` frontmatter is authoritative; the spawn-call `model` argument should match.

---

## Pipeline (main session executes top-to-bottom)

### Stage 0 — Pre-Forge

Main does directly (no sub-agent):

1. Parse args. If `project_id` or `uow_id` missing → ask user, stop.
2. Resolve spec: `.omc/plans/<uow_id>.md` → inline `--` spec → ask, stop.
3. Quick harness scan via Bash: `ls .claude/agents/`, `ls .claude/skills/`, project root for `package.json` / `pyproject.toml` / `go.mod` / `Makefile`, `git rev-parse --abbrev-ref HEAD` (best-effort).
4. `devos_health_check({ project_id })` — record result; failure is best-effort, not blocking.
5. Write or update `.omc/plans/<uow_id>-harness.md` with: 4-layer assessment, confirmed `test_cmd` / `lint_cmd` / `gate_strategy`, reusable specialists, HITL points.

### Stage 1 — 🔥 Ignite (spawn `warvis-initiator`)

```
Agent(
  subagent_type = "warvis-initiator",
  model = "haiku",
  prompt = """
You are running Stage 1 (Ignite). Read your agent definition.

Inputs: project_id, uow_id, harness_config_path, uow_spec, working_dir.

Tasks:
1. Run baseline test + lint (record exit codes; don't block).
2. devos_start_dev_session(project_id, uow_id) → dev_session_id.
3. Create or refresh .omc/plans/<uow_id>.md scaffold from spec.

Return JSON: { stage, dev_session_id, build_cmd, test_cmd, lint_cmd, baseline_pass, baseline_notes, plan_path, harness_config_path, blockers }
"""
)
```

### Stage 2 — 📐 Blueprint (spawn `warvis-planner` with **opus**)

```
Agent(
  subagent_type = "warvis-planner",
  model = "opus",
  prompt = """
You are running Stage 2 (Blueprint). Read your agent definition.

Inputs: project_id, dev_session_id, uow_id, harness_config_path, plan_path, uow_spec.

Tasks:
1. Explore codebase (max 10 reads).
2. Decompose UoW into 3–7 milestones; each is one TDD cycle.
3. Confirm verification strategy from harness gate_strategy.
4. Update plan_path with milestones + Done When + Risk.
5. devos_plan_dev_session(...).

Return JSON: { stage, dev_session_id, milestones[], risk_level, verification_steps[], scope_change_proposed, plan_path, blockers }
"""
)
```

**Risk Gate (main thread):**
- `LOW` / `MEDIUM` → auto-proceed.
- `HIGH` AND `risk_authorized=true` → auto-proceed.
- `HIGH` AND not authorized → ask user; if denied, `devos_block_dev_session` and stop.
- `CRITICAL` → `devos_block_dev_session` immediately and stop.
- `scope_change_proposed=true` → ask user before proceeding.

### Stage 3 — 🔨 Hammer (spawn `warvis-maker`)

```
Agent(
  subagent_type = "warvis-maker",
  model = "sonnet",
  prompt = """
You are running Stage 3 (Hammer). Read your agent definition.

Inputs: project_id, dev_session_id, uow_id, build_cmd, test_cmd, milestones[].

Per milestone: RED → GREEN → REFACTOR → devos_record_evidence.
After all milestones: devos_advance_dev_session.

Constraints:
- No RED skip.
- No `|| true` test suppression.
- Stage-internal helpers (small-diff-implementer / tdd-red / tdd-green / tdd-refactor) ALLOWED — those are direct tool calls inside your stage, not new forge stages.

CRITICAL: Past makers have terminated mid-flight after the last code edit
or during a long-running install, BEFORE the final devos calls. Do NOT
emit a "Now ..." or "Waiting for ..." final utterance. After every
milestone, IMMEDIATELY call devos_record_evidence. After the last
milestone, IMMEDIATELY call devos_advance_dev_session. Then return JSON.

Return JSON: { stage, dev_session_id, milestones_completed, milestones_total, evidence_refs[], final_test_output, final_typecheck_output, final_lint_output, files_changed[], blockers }
"""
)
```

### Stage 4 — ⚗️  Temper (spawn `warvis-verifier` with **opus**)

```
Agent(
  subagent_type = "warvis-verifier",
  model = "opus",
  prompt = """
You are running Stage 4 (Temper). Read your agent definition.

Inputs: project_id, dev_session_id, uow_id, verification_steps[], evidence_refs[].

Gate sequence: Lint → Type check → Unit tests → Integration tests → Security scan (HIGH/CRITICAL only).

Call devos_verify_dev_session with verify_scope and evidence_ref. Do NOT auto-repair.

Return JSON: { stage, dev_session_id, verdict: "PASS|PASS_WITH_WARN|BLOCK", verify_report, failed_gates[], blockers }
"""
)
```

**Recovery (main thread):**
- `PASS` / `PASS_WITH_WARN` → continue to Stage 5.
- `BLOCK` and retries < 2 → re-spawn `warvis-maker` with `failed_gates` and the previous evidence in the prompt; then re-spawn `warvis-verifier`. Increment retry counter.
- `BLOCK` and retries ≥ 2 → `devos_block_dev_session` and stop.

### Stage 5 — 💧 Quench (spawn `warvis-finisher`)

Only if verdict is `PASS` or `PASS_WITH_WARN`.

```
Agent(
  subagent_type = "warvis-finisher",
  model = "haiku",
  prompt = """
You are running Stage 5 (Quench). Read your agent definition.

Inputs: project_id, dev_session_id, uow_id, obsidian_vault_path?, verdict, plan_path, files_changed[], evidence_refs[].

Tasks:
1. Collect summary from plan + files_changed.
2. devos_prepare_lesson(...) → lesson_id.
3. devos_end_dev_session(...) with final_summary.
4. Update plan_path Status → "shipped" with completion timestamp.
5. If obsidian_vault_path: patch UoW note in vault (best-effort).

Return JSON: { stage, dev_session_id, lesson_id, final_summary, vault_patch_status, plan_status_updated, blockers }
"""
)
```

---

## Final Output (main reports to user)

```json
{
  "uow_id": "<id>",
  "project_id": "<id>",
  "dev_session_id": "<id>",
  "verdict": "PASS|PASS_WITH_WARN|BLOCK|KILLED",
  "harness_config_path": ".omc/plans/<uow_id>-harness.md",
  "plan_path": ".omc/plans/<uow_id>.md",
  "stages_completed": ["pre-forge", "ignite", "blueprint", "hammer", "temper", "quench"],
  "evidence_refs": [],
  "lesson_id": "<id>|null",
  "milestones_completed": <n>,
  "files_changed": [],
  "harness_growth_suggestions": [],
  "notes": ""
}
```

## Failure Recovery (main thread responsibilities — VALIDATED)

| Situation | Main thread action |
|---|---|
| Specialist mid-flight termination (most common — ends with "Now ..." utterance) | Inspect disk via Bash. If work largely done, call `devos_record_evidence` + `devos_advance_dev_session(force=true)` directly from main, then continue to next stage. If incomplete, re-spawn the same specialist with progress in the prompt. If still incomplete after 1 retry, report blocker. |
| devos session in wrong state (e.g. PLAN when verify needed) | `devos_advance_dev_session(force=true)`. Record force=true in next evidence note. |
| Long-running install during Hammer | Run install in background from main if specialist died holding it; wait via `until` loop or task notification; resume specialist with installed state. |
| Temper BLOCK ×2 | `devos_block_dev_session` and report failed gates to user. |
| CRITICAL risk discovered any stage | Immediate `devos_block_dev_session`. No further stages. Wait for user. |

## Reuse of Stage 0 Artifacts

If `.omc/plans/<uow_id>-harness.md` or `<uow_id>.md` already exist from a previous failed attempt, reuse them (re-validate against current spec). Pass to initiator with a note to refresh rather than overwrite.

## Use When

- UoW is well-specified with stable inputs.
- devos MCP server is reachable (best-effort otherwise).
- You want a single command to drive Ignite→Blueprint→Hammer→Temper→Quench.

## Do Not Use When

- The UoW is exploratory — run a planning skill first.
- Multiple UoWs need *true* concurrent forging — open separate Claude Code sessions per UoW. Within one session, multiple `/forge` calls execute serially or barrier-style at each stage.

---

## Reference

- Specialist agents: `.claude/agents/warvis-{initiator,planner,maker,verifier,finisher}.md` (model assignments authoritative there)
- Hammer-internal helpers: `.claude/agents/{small-diff-implementer,tdd-red,tdd-green,tdd-refactor}.md`
- devos lifecycle: `devos_start_dev_session` → `_plan_` → `_advance_` → `_record_evidence` → `_verify_` → `_prepare_lesson` → `_end_`
- Deprecated: `warvis-orchestrator` agent (redirect-only stub)
- SDK constraint: https://code.claude.com/docs/en/sub-agents — sub-agents cannot spawn sub-agents
