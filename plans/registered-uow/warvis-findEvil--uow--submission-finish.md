---
note_type: uow
artifact_type: uow
id: BET-BIG-warvis-findEvil-warvis-findevil-bet-cc74
project_id: warvis-findEvil
title: warvis-findEvil Bet — submission finish decomposition
status: active
phase: shape
size: big
appetite_hours: 240
related_bet: BET-BIG-warvis-findEvil-warvis-findevil-bet-cc74
pitch_ref: PITCH-WARVIS-FINDEVIL-PITC-B74C
created_at: 2026-05-09T12:15:47Z
updated_at: 2026-05-09T15:25:03Z
source_session: a8c9ef4f-c212-4b46-b342-d8d57658382c
---

# UoW Decomposition: warvis-findEvil Bet

## Scope
Drive the SANS FIND EVIL hackathon submission package to completion by finishing the remaining mandatory artifacts and submission polish while preserving existing kill-switch, pytest, ruff, and Go-test gates.

## Milestones
1. M1: Repository polish complete (metadata + README license badge).
2. M2: SIFT VM cold-start evidence captured and README quickstart corrected.
3. M3: Devpost submission form draft is paste-ready.
4. M4: Demo video script/recording/upload evidence captured by D-7.
5. M5: Optional Volatility3 real firing residual resolved or explicitly documented as follow-up.

## Tasks

### T1-repo-polish — Repository metadata and README polish
- Priority: P3
- Size: small
- Status: open
- Action: Set GitHub About/topics and add README license badge without altering product behavior.
- Progress: README license badge exists and clone URL now matches the canonical repository; GitHub About/topics readback remains external.
- Done signal: GitHub metadata readback recorded; README has license badge.

### T2-sift-vm-cold-start — Fresh SIFT Workstation clone-to-hunt validation
- Priority: P1
- Size: large
- Status: open
- Action: Provision official SIFT Workstation VM, run documented quickstart from a fresh clone, fix missing README steps, and save full terminal log evidence.
- Done signal: `repos/find-evil-fixtures/cases/sans-starter/sift-vm-cold-start.log` is tracked and tests assert `git clone`, `warvis hunt`, and non-zero `audit.jsonl` evidence.
- Stop condition: If official SIFT VM cannot be provisioned, record STOP note rather than substituting another OS.

### T3-devpost-submission-draft — Devpost form draft from existing narrative sources
- Priority: P2
- Size: medium
- Status: done
- Action: Convert current Devpost page and rules snapshot into field-by-field paste-ready form content.
- Done signal: `docs/find-evil/devpost-form-draft.md` contains every required field heading and paste-ready content with video URL placeholder.

### T4-demo-video — ≤5 minute demo video showing agent self-correction
- Priority: P1
- Size: large
- Status: open
- Action: Update demo script, record/upload unlisted video, and write video URL into Devpost form draft.
- Dependencies: D-7 recording window around 2026-06-08; T3 form draft exists; 26B trace and comprehensive mock trace available.
- Done signal: Unlisted video URL is recorded in `docs/find-evil/devpost-form-draft.md`; local `docs/find-evil/demo.mp4` handling is confirmed gitignored.

### T5-vol3-real-firing-optional — Optional Volatility3 real firing evidence residual
- Priority: P4
- Size: medium
- Status: open
- Action: Attempt one trace where the agent loop actually invokes volatility3 pslist on the SANS memory image, or document honest miss if time-box expires.
- Dependencies: T1-T4 complete; buffer remains before D-5 freeze.
- Done signal: Either `tool_result` evidence spanning at least 30s exists, or `accuracy-report.md` honest caveat documents the miss and next step.

## Verification Strategy
- `python -m pytest tests/ -q` passes after each shipped code/doc-test change.
- `ruff check tests/` passes for test-only lint gate.
- `( cd warvis && go test ./... )` passes for Go orchestrator integrity.
- `make -C harness/find-evil kill-switch-check` remains 5/5 PASS after any code or mock change.
- `docs/find-evil/devpost-form-draft.md` contains demo URL before final submission package is considered complete.
- SIFT VM cold-start log exists and contains `git clone`, `warvis hunt`, and non-zero `audit.jsonl` evidence.
