# ITEM-212 FIND EVIL — R+1 Handover (Phase 0 → Phase 1)

- **Closed (UTC)**: 2026-04-29T22:52:33Z (handover write)
- **Author**: dev-node, Claude Code session under `/home/jang/Workspace/warvis-forRich`
- **Inbound card**: `output/handoff-cards/HANDOFF-212-find-evil.md` (warvis-sub, mission-control)
- **Meta-rule satisfied**: KICK-749 (R+1 inbound spec from build node)

## 1. Phase 0 Output Index

| Artifact | Path |
|---|---|
| Devpost evidence | `plans/ITEM-212-find-evil/evidence/devpost-20260429T225233Z.md` |
| Spec | `plans/ITEM-212-find-evil/spec.md` |
| Agent: orchestrator | `.claude/agents/find-evil/find-evil-orchestrator.md` |
| Agent: mcp-architect | `.claude/agents/find-evil/find-evil-mcp-architect.md` |
| Agent: mcp-implementer | `.claude/agents/find-evil/find-evil-mcp-implementer.md` |
| Agent: detector-engineer | `.claude/agents/find-evil/find-evil-detector-engineer.md` |
| Agent: evidence-curator | `.claude/agents/find-evil/find-evil-evidence-curator.md` |
| Agent: verifier | `.claude/agents/find-evil/find-evil-verifier.md` |
| Agent: pitch-writer | `.claude/agents/find-evil/find-evil-pitch-writer.md` |
| Harness: env sample | `harness/find-evil/env.sample` |
| Harness: docker-compose | `harness/find-evil/docker-compose.yml` |
| Harness: Makefile | `harness/find-evil/Makefile` |
| Harness: gates.yaml | `harness/find-evil/gates.yaml` |
| Harness: secrets policy | `harness/find-evil/secrets.policy.md` |
| Harness: secrets-scan hook | `harness/find-evil/scripts/secrets-scan.sh` |

All 7 agent files declare `model: claude-opus-4-7` per KICK-749.

## 2. Spec Findings to Escalate (mission-control)

1. **$22K is the total prize pool, not an MCP track allocation** — see evidence §"Prize Pool". Reframe expected value (1st=$10K, 2nd=$7.5K, 3rd=$4.5K).
2. **MCP server is one of four allowed architectures** — strategy still defensible (criterion #4 rewards architectural constraints), but framing as "MCP track" may overstate exclusivity.
3. **Reuse fit is partial** — `warvis-mcp` and `obsidian-mcp` are dev-only; SENTINEL-lite (Solidity) is largely OOS for DFIR. Honest reuse story needed in Devpost copy.
4. **SIFT Workstation dependency** — Phase 1 must decide between local OVA install vs. Docker proxy (open question Q2 in spec §9).
5. **Submission is local-deploy** — no hosting / cloud / payment integration needed; reduces NFR surface.

## 3. Next-Session First Command

```
@find-evil-orchestrator "Phase 1 진입.
- 입력: plans/ITEM-212-find-evil/spec.md, harness/find-evil/gates.yaml
- 목표: phase_1.required 게이트 통과 (MCP skeleton + 3 tools schema + capability handshake)
- 산출 시작점: find-evil-mcp-architect → mcp-schema/{case.open,timeline.build,iocs.scan}.json
- 게이트 실패 시 사용자에게 결정권 반환"
```

## 4. Risk Register (Top 3)

| # | Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|---|
| R1 | DFIR domain knowledge gap (we are Solidity-native) | High | High | Phase 1 절반은 Protocol SIFT NotebookLM 학습. 외부 DFIR 컨설턴트 reach-out을 mission-control에 요청. |
| R2 | Recall ≥ 0.80 미달 (D-7 gate) | Medium | High | curator가 fixture 정의를 conservative 하게(우리가 이길 수 있는 패턴 위주). spec.md §8 D-15 게이트 0.60에서 조기 cut. |
| R3 | SIFT VM 호스팅 부담으로 dev 환경 못 갖춤 | Medium | Medium | Phase 1 D+2까지 Docker proxy 결정 못하면 OVA 설치로 강제 전환. |

## 5. Kill Switch Triggers

- D-26 (2026-05-20): MCP skeleton + 3 tools 미동작 → scope cut to 4 tools max.
- D-15 (2026-06-01): recall < 0.60 → **submission kill**, 대체 ITEM 4w 재배치.
- D-7  (2026-06-08): 제출 패키지 8종 중 3건 이상 결손 → best-effort 제출, D-1 push 금지.
- D-1  (2026-06-14): mission-control review BLOCK → 차기 hackathon 이월.

## 6. Cross-Node Sync Need

- Dev frontmatter 키 갱신 필요 (`@forrich-cross-node-sync`):
  - `dev_active_builds*` += ITEM-212 (Phase 0 closed → Phase 1 inbound)
  - `last_updated_dev` = 2026-04-30
- Research/공통 frontmatter, 본문 H2, `roadmap.md`는 **read-only** (KICK-375 위반 시 high severity).
- 핸드오프 카드 `output/handoff-cards/HANDOFF-212-find-evil.md` status: pending → accepted (sync 에이전트가 갱신).

## 7. Settings.local.json Diff Proposal (apply-after-approval)

Adds Bash/Edit allow-list for find-evil paths only.

```diff
 {
   "permissions": {
     "allow": [
       "Bash(mkdir -p .claude/skills/repo-quality-gate .claude/skills/slither-scan .claude/skills/bet-lifecycle)",
       "Bash(mv skills/repo-quality-gate/SKILL.md .claude/skills/repo-quality-gate/SKILL.md)",
       "Bash(mv skills/slither-scan/SKILL.md .claude/skills/slither-scan/SKILL.md)",
       "Bash(mv skills/bet-lifecycle/SKILL.md .claude/skills/bet-lifecycle/SKILL.md)",
-      "Bash(rm -rf skills/.claude skills/repo-quality-gate skills/slither-scan skills/bet-lifecycle)"
+      "Bash(rm -rf skills/.claude skills/repo-quality-gate skills/slither-scan skills/bet-lifecycle)",
+      "Bash(make -C harness/find-evil *)",
+      "Bash(docker compose -f harness/find-evil/docker-compose.yml *)",
+      "Bash(harness/find-evil/scripts/secrets-scan.sh)",
+      "Edit(harness/find-evil/**)",
+      "Edit(.claude/agents/find-evil/**)",
+      "Edit(plans/ITEM-212-find-evil/**)",
+      "Edit(src/find_evil_mcp/**)",
+      "Edit(docs/find-evil/**)"
     ]
   },
   "mcpServers": { ... unchanged ... }
 }
```

> Per kickoff §6: this diff is **proposal only** — apply only after user approval.

## 8. Required `.gitignore` additions (proposal)

```
# ITEM-212 FIND EVIL
harness/find-evil/.env
harness/find-evil/.env.*
!harness/find-evil/env.sample
harness/find-evil/logs/*.jsonl
docs/find-evil/demo.mp4
repos/find-evil-fixtures/**/raw/**
src/find_evil_mcp/**/__pycache__/
```

(Already in repo gitignore: `.env`, `.env.*` — sufficient for env files but not for logs/fixtures/cache.)

## 9. Open Questions Carried (mission-control reply requested)

See spec.md §9 (Q1–Q5). Phase 1 cannot proceed past D+2 without Q1 (team size) and Q2 (SIFT hosting).

## 10. Definition-of-Done for THIS Session

- [x] Step 1 evidence captured (Devpost overview + rules + resources)
- [x] Step 2 spec.md authored
- [x] Step 3 7 agents authored, all `model: claude-opus-4-7`
- [x] Step 4 harness/find-evil/ scaffolded; spec-check target wired
- [x] Step 5 this handover-r1.md
- [ ] User approval on settings.local.json diff (§7)
- [ ] git commit with prefix `item-212(phase-0):`
