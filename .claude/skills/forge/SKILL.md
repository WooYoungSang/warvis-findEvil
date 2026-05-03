---
name: forge
description: Full UoW lifecycle via warvis-* pipeline — Ignite→Blueprint→Hammer→Temper→Quench
argument-hint: "<project_id> <uow_id> [obsidian=<vault_path>] [-- <inline spec>]"
level: 3
---

Run the warvis-* pipeline to fully implement UoW **{{PROMPT}}**.

## Arguments

Parse `{{PROMPT}}` as follows:

```
<project_id>  <uow_id>  [obsidian=<vault_path>]  [-- <uow_spec>]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `project_id` | yes | Project slug (e.g. `warvis-ignis`) |
| `uow_id` | yes | Unit of Work ID (e.g. `V3-003`) |
| `obsidian_vault_path` | no | Absolute path to Obsidian vault root. If omitted, finisher skips vault update (non-blocking). |
| `uow_spec` | no | Inline spec after `--`. Used when `.omc/plans/<uow_id>.md` does not exist. |

**Spec resolution order:** `.omc/plans/<uow_id>.md` → inline `--` spec → ask user then stop.

## Use When

- UoW ID and project are known and implementation should run end-to-end automatically
- devos MCP server is reachable (`devos_health_check` passes)
- You want TDD → verify → lesson in a single unattended run

## Do Not Use When

- No devos MCP server — call warvis-* agents directly instead
- Multiple UoWs concurrently — run a separate `forge` session per UoW

---

## Pipeline

```
Stage 1  Ignite     warvis-initiator  → devos_start_dev_session
Stage 2  Blueprint  warvis-planner    → devos_plan_dev_session
Stage 3  Hammer     warvis-maker      → devos_record_evidence × N
                                      + devos_advance_dev_session
Stage 4  Temper     warvis-verifier   → devos_verify_dev_session
Stage 5  Quench     warvis-finisher   → devos_prepare_lesson
                                      + devos_end_dev_session
```

Each stage runs sequentially. Output of each stage feeds the next.

---

## Stage 0 — Preflight

1. Extract `project_id`, `uow_id`, `obsidian_vault_path`, `uow_spec` from `{{PROMPT}}`.
2. Load spec: read `.omc/plans/<uow_id>.md` if it exists; fall back to inline spec; stop and ask if neither is available.
3. Call `devos_health_check()` — abort immediately on failure.

---

## Stage 1 — Ignite  (`warvis-initiator`, model=haiku)

Inputs passed to agent:
- `project_id`, `uow_id`, `uow_spec`

Agent performs:
- Read `CLAUDE.md` / `AGENTS.md` for build and test commands
- `git status` + baseline test run (failures recorded, not blocking)
- `devos_start_dev_session(project_id, uow_id)` → returns `dev_session_id`
- Create `.omc/plans/<uow_id>.md` scaffold if absent

Outputs: `dev_session_id`, `build_cmd`, `test_cmd`, `baseline_pass`

---

## Stage 2 — Blueprint  (`warvis-planner`, model=sonnet)

Inputs: `project_id`, `dev_session_id`, `uow_id`, `uow_spec`

Agent performs:
- Explore codebase (max 10 reads)
- Decompose spec into 3–7 milestones (each milestone = one TDD cycle with a clear done-signal)
- Write milestones + verification strategy into `.omc/plans/<uow_id>.md`
- `devos_plan_dev_session(project_id, dev_session_id, uow_id, milestone_strategy, verification_strategy, task_specs)`

> **Risk Gate:** HIGH/CRITICAL risk → pause for user confirmation before proceeding.
> If confirmation unavailable, call `devos_block_dev_session` and stop.

---

## Stage 3 — Hammer  (`warvis-maker`, model=sonnet)

Inputs: `project_id`, `dev_session_id`, `uow_id`, `build_cmd`, `test_cmd`

Per-milestone TDD cycle:

```
RED      → Write failing test; confirm it FAILS (not errors)
GREEN    → Implement minimum code to make it pass
REFACTOR → Remove duplication; full test suite must pass
RECORD   → devos_record_evidence(project_id, dev_session_id, uow_id,
               bundle_type="test",
               artifacts={milestone, test_output, files_changed})
```

After all milestones:
```
devos_advance_dev_session(project_id, dev_session_id, uow_id)
```

Constraints:
- **No sub-agents.** Direct execution only.
- **No RED skip.** Never implement without a failing test first.
- **No `|| true` suppression** of test failures.

---

## Stage 4 — Temper  (`warvis-verifier`, model=sonnet)

Inputs: `project_id`, `dev_session_id`, `uow_id`
(reads verification strategy from `.omc/plans/<uow_id>.md`)

Gate sequence: Lint → Type Check → Unit Tests → Integration Tests → Security (if HIGH/CRITICAL)

```
devos_verify_dev_session(project_id, dev_session_id, uow_id, verify_scope=[...])
```

Verdicts: `PASS` | `PASS_WITH_WARN` | `BLOCK`

- `BLOCK` → return to Hammer for fixes; retry up to **2 times**
- After 2 blocked retries → `devos_block_dev_session(...)` and report to user

---

## Stage 5 — Quench  (`warvis-finisher`, model=haiku)

Inputs: `project_id`, `dev_session_id`, `uow_id`, `obsidian_vault_path` (optional), verdict (must be PASS or PASS_WITH_WARN)

Agent performs:
1. Collect session summary from `.omc/plans/<uow_id>.md` + `git diff`
2. `devos_prepare_lesson(project_id, dev_session_id, uow_id, title, summary, retrieval_tags, compaction_pipeline=True)`
3. `devos_end_dev_session(project_id, dev_session_id, uow_id, final_summary)`
4. Update `.omc/plans/<uow_id>.md` → `status: shipped`
5. If `obsidian_vault_path` provided: patch UoW note via obsidian-mcp (**best-effort, non-blocking**)

---

## Blocker Handling

| Condition | Action |
|-----------|--------|
| `devos_health_check` fails | Abort; ask user to check devos MCP server |
| No UoW spec found | Abort; ask user to provide `-- <spec>` or create plan file |
| Risk HIGH/CRITICAL at Blueprint | Wait for user confirmation |
| Temper BLOCK after 2 retries | `devos_block_dev_session`; report failures to user |
| Unexpected exception at any stage | Report stage name + error; do not continue |

---

## Completion Checklist

- [ ] `devos_health_check` passed
- [ ] `dev_session_id` obtained (Ignite)
- [ ] All milestones marked `[x]` in `.omc/plans/<uow_id>.md` (Hammer)
- [ ] `devos_record_evidence` called for every milestone (Hammer)
- [ ] `devos_advance_dev_session` called (Hammer)
- [ ] `devos_verify_dev_session` → PASS or PASS_WITH_WARN (Temper)
- [ ] `devos_prepare_lesson` called (Quench)
- [ ] `devos_end_dev_session` called (Quench)
- [ ] `.omc/plans/<uow_id>.md` status = `shipped`

Original task:
{{PROMPT}}
