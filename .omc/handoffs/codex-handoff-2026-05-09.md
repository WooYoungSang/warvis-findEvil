# Codex CLI Handoff — warvis-findEval Phase 4 (D-37 to deadline)

> Paste this entire file as the first message to Codex CLI when starting a
> new session in `/home/jang/Workspace/warvis-findEvil`.

---

## Your Mission

Drive the `warvis-findEval` project to a complete SANS FIND EVIL hackathon
submission by **2026-06-15 23:45 EDT** (currently D-37). The codebase is
healthy and the hackathon's 8 mandatory artifacts are largely complete; you
own the remaining ~5 tasks listed below. Pace work against the published
WINNING_PLAYBOOK timeline (record demo D-7, freeze at D-2).

## Authoritative Sources of Truth (read these first)

| File | Why |
|------|-----|
| `CLAUDE.md` | Tech stack, build/test commands, agent registry, harness policy |
| `docs/find-evil/devpost-page.md` | Submission narrative (current draft) |
| `docs/find-evil/architecture.md` | 650-line architecture (Phase 2 + 3 + 4) |
| `docs/find-evil/accuracy-report.md` | Recall measurements + honest caveats §4, §8, §8a, §8b |
| `docs/find-evil/valhuntir-comparison.md` | Differentiation matrix vs reference impl. §6 has 5-action plan; A1+A2+A4 done. |
| `plans/ITEM-212-find-evil/spec.md` | Project spec, judging criteria #1–#6 |
| `plans/ITEM-212-find-evil/evidence/devpost-20260429T225233Z.md` | Frozen Devpost rules snapshot (8 mandatory artifacts, eligibility, prizes) |
| `.omc/plans/` | Per-UoW shaped Bets and shipped plans (10 lessons captured so far) |

## Project Snapshot (2026-05-09)

| | |
|---|---|
| Repo | `/home/jang/Workspace/warvis-findEvil`, github `WooYoungSang/warvis-findEvil`, public, MIT |
| Branch | `main` (current HEAD `bb9dbd8`) |
| Tests | 91 / 91 pytest PASS, Go test PASS, `make -C harness/find-evil kill-switch-check` 5 / 5 PASS, `ruff check tests/` clean |
| Hackathon artifacts | 7 of 8 fully shipped; #2 demo video and #7 SIFT-VM-validated quickstart are the remaining mandatories |
| Real-hunt evidence | 3 traces under `repos/find-evil-fixtures/cases/sans-starter/` (e4b, 26b, comprehensive-mock); 8 of 9 lock-ins green for the latest Bet |
| GPU | RTX 4090 (24 GB VRAM); Ollama at `localhost:29134`; gemma4:e4b warm, 26B fits |
| Mock harness | `harness/find-evil/mock_ollama_server.py` and `mock_mcp_server.py` for deterministic kill-switch + comprehensive-mock-trace |
| Lessons captured | 10 across `.omc/plans/` and devos store; reuse them rather than reinventing |

## Constraints — DO NOT VIOLATE

- **Kill-switch must stay 5/5 PASS.** Run `make -C harness/find-evil kill-switch-check` after every code change. The Phase 3 lock is non-negotiable.
- **`src/find_evil_mcp/` is mostly Phase 1+2 locked.** It received two surgical patches in B3 (mcp 1.x SDK migration + structured-output tuple in `call_tool`). Any additional change requires the same explicit user approval the original maintainer gave; ask first.
- **No sudo without explicit approval.** A one-time `/evidence/sans-starter` symlink at filesystem root is the only sudo footprint so far.
- **Honest dual-path documentation pattern is load-bearing.** Match the tone of `accuracy-report.md` §4 / §8 — say what you measured, say what you didn't, never overclaim. The "honest L4 caveat" in §8b is the model.
- **Deadline freeze**: stop net-new code at D-5 (2026-06-10); after that only fix-fixes for the demo recording and form submission.

## Your Workflow per Task — Use the Forge Pipeline

This project's UoW conventions live in `.claude/skills/forge/`. The Claude-Code-native skill calls dedicated subagents (`warvis-initiator`, `warvis-planner`, `warvis-maker`, `warvis-verifier`, `warvis-finisher`) for the five stages below. **Codex CLI does not load that skill, but the staged workflow itself is the value — replicate it explicitly using whatever subagent / spawn mechanism your runtime provides** (Codex's own subagents, `gpt-5-codex` with role prompting, etc.). Spawning a fresh agent per stage keeps each context tight and forces a paper trail.

For every task in this handoff, run all five stages — do not collapse them into a single monolithic agent run; the discipline is what catches the four-bug-cascade-style discoveries the project has been making.

| Stage | Role / agent | Inputs | Outputs |
|------:|-------------|--------|---------|
| 1. **Ignite** | initiator (read-only quick scan) | task slug, project tree | `.omc/plans/<slug>.md` scaffold; baseline test-run + `git status` capture; STOP-points named |
| 2. **Blueprint** | planner | scaffold + relevant context files (max ~10 reads) | 3–5 milestones written into the same plan file with explicit done-signals; risk rating LOW / MED / HIGH; verification strategy; any HIGH risk pauses for user approval |
| 3. **Hammer** | maker | plan file + build/test commands | RED test first, then GREEN code, then REFACTOR. One commit per milestone (or one commit at the end with a multi-paragraph body matching `git log --format=fuller -n 5`). No sub-spawning beyond this stage. |
| 4. **Temper** | verifier | plan file + verification strategy | Runs the four gates below; emits PASS / PASS_WITH_WARN / BLOCK; on BLOCK returns failure list to maker for ≤2 retry rounds. |
| 5. **Quench** | finisher | full session artefacts + verifier verdict | Updates the plan file frontmatter to `status: shipped` with a 1-line `lock_in_green` / `lock_in_residual` summary; commits the plan; pushes. |

The **four standard gates** at Stage 4:

```bash
python -m pytest tests/ -q
ruff check tests/
( cd warvis && go test ./... )
make -C harness/find-evil kill-switch-check
```

Hard rules across all stages:

- **Bet first when scope is open or risky.** Drop a `bet-warvis-findeval--<slug>.md` into `.omc/plans/` modelled after `bet-warvis-findeval--prompt-context-threading.md` (frontmatter + lock-in conditions + scope hammer + kill condition). Tasks 2 and 4 below should each get their own Bet; Tasks 1 and 3 are small enough to skip the Bet step.
- **Commit attribution.** Use `Co-Authored-By: Codex <noreply@openai.com>` (or your runtime's equivalent). Match the existing multi-paragraph commit body style — read `git log --format=fuller -n 5` before writing.
- **Cwd-independent tests.** All new `tests/test_*.py` must use `Path(__file__).resolve().parent.parent` for repo-root resolution; the existing suite is cwd-stable and breaks if you regress that.
- **DO NOT call devos MCP tools.** This project has historically logged each UoW lifecycle to a `devos_*` MCP server, but Codex CLI is not connected to it. Treat `.omc/plans/<slug>.md` (Bet) and the resulting commits as the durable handoff artefacts — they fully replace devos's role.
- **Sub-agent spawning is encouraged but not required.** If Codex CLI exposes a `Task` / `agent` / sub-spawn primitive, use it per stage so each role gets a clean context. If not, run the stages sequentially in one agent but write each stage's output to the plan file before moving on — that is what makes the discipline observable.

## Tasks (priority order)

### Task 1 — `repo-polish` (~30 min, P3 but trivial)

GitHub repo About / topics / social-preview is empty. Quick polish:

```bash
gh repo edit WooYoungSang/warvis-findEvil \
    --description "W.A.R.V.I.S. — Forensic IR agent (Go orchestrator + Python MCP + Gemma 4). SANS FIND EVIL hackathon entry by WoopsFactory." \
    --add-topic forensics --add-topic dfir --add-topic mcp --add-topic gemma \
    --add-topic ollama --add-topic finite-state-machine --add-topic sans-find-evil
```

Add a license badge to the top of `README.md`. Verify `gh repo view --web` looks reasonable. Commit message style: `chore: repo metadata polish (topics, About, license badge)`.

### Task 2 — `sift-vm-cold-start` (~1-2 days, P1, mandatory artifact #7)

Validate that a fresh SIFT Workstation VM can clone the repo and run `warvis hunt` to completion using only the documented quickstart in `README.md`. The hackathon rules require this evidence; substituting a different OS is not acceptable.

1. **Provision a SIFT VM.** OVA at https://sans.org/tools/sift-workstation (~5 GB). If you cannot acquire it (network restrictions, disk pressure, no virtualization), STOP and ask the user to provision it; do not proceed with a substitute OS.
2. Inside the running SIFT VM, clone the repo, run the documented quickstart, capture the entire terminal session as `repos/find-evil-fixtures/cases/sans-starter/sift-vm-cold-start.log`.
3. As you discover any missing install step, fix `README.md` until clone-to-hunt succeeds with no manual interventions.
4. Add a pytest assertion in `tests/test_sift_cold_start.py` that the log file exists and contains the strings `git clone`, `warvis hunt`, and a non-zero audit.jsonl line count.

Done signal: tracked log file + README delta + green pytest. If the SIFT VM cannot be provisioned, file a STOP note in `.omc/plans/sift-vm-cold-start.md` and surface the blocker to the user.

### Task 3 — `devpost-submission-draft` (~1 day, P2)

The Devpost form has 12 fields per WINNING_PLAYBOOK §1. `docs/find-evil/devpost-page.md` is the long-form draft; you need to project it into form-shape:

1. Read `plans/ITEM-212-find-evil/evidence/devpost-20260429T225233Z.md` for the actual form fields.
2. Read `docs/find-evil/devpost-page.md` for narrative content.
3. Read `docs/find-evil/valhuntir-comparison.md` §5 for the canonical honest tagline.
4. Produce `docs/find-evil/devpost-form-draft.md` with each field labeled and the content ready to paste:
   - Project name (verbatim "W.A.R.V.I.S. — Find Evil")
   - Tagline (≤ 80 chars, recommend the §5 statement trimmed)
   - Short description (≤ 200 chars)
   - Long description (markdown, problem → what → how → why)
   - How it's made (tech stack, integration points)
   - Built with (tags)
   - Try it out link (the README quickstart on GitHub for now; a live demo URL is not in scope)
   - Video URL (placeholder; filled at Task 4)
   - Image gallery thumbnail spec (1280×640 — describe what you'd put on it; actual image generation is the contributor's call)
   - GitHub repo (https://github.com/WooYoungSang/warvis-findEvil)
   - Team members + W-8BEN reminder (non-US winner)
5. Add a pytest assertion in `tests/test_devpost_form_draft.py` that the file exists and contains every required field heading.

### Task 4 — `demo-video` (~2-3 days, P1, mandatory artifact #2). Don't run this until D-7 (~2026-06-08).

Per Devpost rules and SANS criterion #1, the demo video must be ≤5 min and **show agent self-correction**. The 26B real-hunt-trace at `repos/.../sans-starter/real-hunt-trace-26b/` already demonstrates self-correction (Gemma escalating instead of fabricating); that audit is your primary screen content.

1. Write `docs/find-evil/demo-script.md` (one already exists — extend it with the new traces from B3 and the comprehensive-mock-trace from the prompt-context-threading Bet).
2. Record (suggested tools: asciinema for terminal, OBS for screen) showing:
   - 0:00–0:15 Problem statement: "AI threats strike in minutes; build the defender that responds in seconds" (Devpost tagline).
   - 0:15–0:45 Architecture diagram (`docs/find-evil/architecture.md` §11 mermaid is suitable).
   - 0:45–2:00 Pre-recorded `warvis hunt /evidence/sans-starter/base-wkstn-05-memory.img` against gemma4:26b. The deliverable is a recorded screencast, not a live demo — record once Gemma produces a clean run (you may need 2–3 takes given LLM stochasticity; pre-warm 26B before each take to skip the cold-load).
   - 2:00–3:30 Inspect the audit.jsonl: highlight `gemma_response[TRACE] action=escalate reason="Missing critical context: ..."` — this IS the self-correction the rubric demands.
   - 3:30–4:30 The deterministic comprehensive-mock-trace at `repos/.../sans-starter/comprehensive-mock-trace/audit.jsonl`: full INITIALIZE→TRACE→SCAN→EXPOSE→LOCK traversal in 22 lines.
   - 4:30–5:00 Acknowledge Valhuntir + show the honest positioning statement from `devpost-page.md`. End with the GitHub URL.
3. Upload as YouTube unlisted; capture the URL into `docs/find-evil/devpost-form-draft.md`.
4. Save the video file locally as `docs/find-evil/demo.mp4` (already gitignored under `.gitignore`'s `*.mp4` rule — verify).

### Task 5 — Optional: `vol3-real-firing` (Bet residual L4)

This is the unfinished `prompt-context-threading` lock-in L4. Only attempt if Tasks 1–4 are complete and there is buffer time before D-5 freeze.

Goal: produce one trace where `vol -f base-wkstn-05-memory.img windows.pslist.PsList` actually fires from the agent loop, evidenced by `tool_result` events spanning ≥30 wall-clock seconds.

Two viable approaches (pick one):

(a) **Smarter mock** — Extend `mock_ollama_server.py` to read the most recent system prompt from a side channel and inject the actual case_id substring into its action arguments. This makes vol3 fire deterministically. Risk: changes the kill-switch baseline if not done carefully.

(b) **Real-Gemma rerun budget** — Pre-warm 26B, run `warvis hunt` with `WARVIS_MAX_TURNS=15`, retry up to 5 times. Stochastic; one in five may produce a clean SCAN reach with `memory.process_list` arguments containing the live case_id. If successful, the resulting trace is the strongest single evidence for criterion #2 in the entire submission.

If neither approach yields evidence within ~4 hours, ship the existing 8/9 lock-in unchanged and tag this as a known follow-up in `accuracy-report.md` §8b.

## Things That Will Bite You

- **Bash shell `cd` persists across tool calls** in some CLI environments. Tests use `Path(__file__).resolve().parent.parent` to be cwd-independent — preserve that pattern.
- **Egnyte share URLs return text/html, not binaries.** Don't try to scrape them programmatically; the slop hook will fire and rightly so. The user manually downloaded `base-wkstn-05-memory.7z` at session start; that file is now at `/mnt/disk1/INCUBATOR/warvis-findEvil-fixtures/sans-starter/base-wkstn-05-memory.img` (3.0 GiB extracted) and reachable via the `repos/find-evil-fixtures/cases/sans-starter/evidence/` symlink.
- **Gemma 26B requires pre-warm.** First inference takes ~9.5 s for model load; pkg/ollama/client.go HTTP timeout was raised from 120 s to 600 s in B3 to accommodate this. Don't lower it.
- **`/evidence/sans-starter` is a sudo-created root symlink.** If the host reboots, recreate per the README. Phase 1+2 schema enforces `^/evidence/...` paths.
- **Mock Ollama action sequence affects kill-switch.** Indices 0–2 are TRACE actions used by KS; indices 3+ are the SCAN/EXPOSE actions added in `prompt-context-threading`. KS uses `WARVIS_MAX_TURNS=3` so it never reaches index 3. If you change the mock, run kill-switch immediately to verify regression.
- **Tests at `tests/test_real_hunt_trace.py` use timestamps from the committed traces.** If you regenerate traces, expect those tests to need refreshing (timestamps change but invariants — hash chain, FSM transitions, tool_called events — remain).
- **`warvis/bin/warvis` is committed** — yes, an unusual choice, but it makes the demo reproducible without requiring Go on every reviewer's machine. Rebuild it before commit when you change Go sources: `cd warvis && go build -o bin/warvis ./cmd/warvis`.

## Definition of "Done" for the Whole Handoff

- [ ] Tasks 1, 2, 3 shipped (P1/P2/P3); Task 4 recorded and uploaded by D-7.
- [ ] All four verification gates green at end-of-handoff: pytest, ruff, go test, kill-switch.
- [ ] Devpost form draft is paste-ready.
- [ ] Demo video URL appears in `devpost-form-draft.md`.
- [ ] All commits on main; pushed to origin; no uncommitted changes in working tree.
- [ ] If Task 5 (L4 follow-up) is attempted, ship outcome (success or honest miss) — don't leave it half-done.

Good hunting.

— Handoff written 2026-05-09 by Claude Opus 4.7 (1M context). Lessons up to id `e8a97f4d` in devos store.
