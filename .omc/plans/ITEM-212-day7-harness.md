# Harness Config: ITEM-212-day7-integration

## Always-on
- claude_md: exists (map quality, 82 lines)
- local_agents: 28 agents discovered
- CLAUDE.md contains comprehensive agent routing; no encyclopedia warning

## On-demand
- skills_available: [update-config, keybindings-help, simplify, fewer-permission-prompts, loop, schedule, claude-api, omc-reference, pattern-extract, forge, spec-gap-check, bet-kickoff, ship-or-cut, build-report, harness-bootstrap, harness-audit, strategic-compact, oh-my-claudecode:self-improve, oh-my-claudecode:omc-reference, oh-my-claudecode:sciomc, oh-my-claudecode:plan, oh-my-claudecode:wiki, oh-my-claudecode:learner, oh-my-claudecode:ultrawork, oh-my-claudecode:ralph, oh-my-claudecode:release, oh-my-claudecode:omc-teams, oh-my-claudecode:ultraqa, oh-my-claudecode:configure-notifications, oh-my-claudecode:external-context, oh-my-claudecode:deep-interview, oh-my-claudecode:trace, oh-my-claudecode:ask, oh-my-claudecode:omc-setup, oh-my-claudecode:cancel, oh-my-claudecode:autoresearch, oh-my-claudecode:verify, oh-my-claudecode:skill, oh-my-claudecode:hud, oh-my-claudecode:autopilot, oh-my-claudecode:ralplan, oh-my-claudecode:debug, oh-my-claudecode:team, oh-my-claudecode:ai-slop-cleaner, oh-my-claudecode:skillify, oh-my-claudecode:deepinit, oh-my-claudecode:project-session-manager, oh-my-claudecode:omc-doctor, oh-my-claudecode:ccg, oh-my-claudecode:mcp-setup, oh-my-claudecode:writer-memory, oh-my-claudecode:deep-dive, init, review, security-review]
- reuse: [forge (warvis lifecycle), oh-my-claudecode:ultraqa (integration testing), oh-my-claudecode:verify (quality gates)]
- new_skill_candidate: integration-test-orchestration (Day 7 milestones M7a-M7f repeat in future SANS work)

## Deterministic
- test_cmd: `cd warvis && go test ./... -short`
- lint_cmd: `cd warvis && go vet ./...`
- eval_harness: exists (kill-switch-check 1-5)
- gate_strategy: [lint → test-short → kill-switch-check (integrated) → security scan]

## Measurable
- baseline_eval: missing (first integration test run)
- observability: structured audit.jsonl (append-only JSON lines per FSM transition)
- trace_plan: devos_record_evidence per milestone M7a-M7f

## Collaboration Map
- impl: direct (orchestrator executes M7a-M7f sequentially, TDD per milestone)
- test: direct (integration tests in warvis/internal/integration/ + harness/find-evil/Makefile)
- security: gate3 pattern scan (no hardcoded secrets in test fixtures)

## HITL Points
- HIGH risk: timeout enforcement (120s tool call, 600s state, 1800s hunt total) — test under load
- HIGH risk: resume capability (--case-id --phase) — verify prior state fully restored
- CRITICAL: src/find_evil_mcp/ never modified (absolute constraint)

## Dead Weight Check
- CLAUDE.md: contains "Never break src/find_evil_mcp/" policy (keep, load-bearing)
- Agent delegation table: remains relevant for Phase 4 (keep)

## Summary
**Option B: Direct Execution**
- Harness quality: HIGH (CLAUDE.md map, 28 agents, test gates ready)
- Implementation path: M7a → M7b → M7c → M7d → M7e → M7f sequentially with TDD
- No sub-agent delegation needed (Phase 3 Go bridge sufficiently scoped)
- Devos session: start → plan → advance milestones → verify → end
- Completion: M7f delivers verify-phase3.md with kill-switch pass/fail verdict + Phase 4 handover condition

---

**Harness Config approved for ITEM-212-day7-integration**
**Execution start: M7a (Integration test suite)**
