# ITEM-212 FIND EVIL — Phase 3 Handover (R+4)

- **Issued**: 2026-04-30
- **From**: dev-node Phase 2 Step 3+4 session (`ea704702-6e63-4e04-9a28-9f8f0d2b9c23`)
- **Verdict**: Phase 2 PASS (verify-phase-2-full.md) — caveats documented
- **Phase 3 window**: 2026-06-02 → 2026-06-08 (spec.md §10)

## 1. Phase 2 Final State

| Asset | Path | Status |
|---|---|---|
| MCP schemas (9) | `harness/find-evil/mcp-schema/*.json` | complete |
| Tool handlers (9, SIFT-wired + lite fallback) | `src/find_evil_mcp/tools/*.py` | complete |
| SIFT subprocess wrapper | `src/find_evil_mcp/sift_runner.py` | whitelist + metachar reject + timeout |
| Tests (90 PASS) | `tests/test_{schema,server,detector}.py` | green |
| Synthetic fixtures | `repos/find-evil-fixtures/cases/case-001/` (→ disk1) | 13 ground-truth findings |
| Recall harness | `harness/find-evil/scripts/accuracy.py` | lite-mode: 100% on synthetic |
| Conformance harness | `harness/find-evil/scripts/mcp_handshake_check.py` | 9 tools |
| Architecture doc | `docs/find-evil/architecture.md` | subprocess execution model documented |
| Verify reports | `plans/ITEM-212-find-evil/verify-phase-{1,2-partial,2-full}.md` | all PASS-grade |

## 2. Honest Caveats Carried into Phase 3

1. **Synthetic recall** — 100% lite-mode on hand-crafted 13-finding fixture. Real DFIR-sample recall (SANS evidence) is unmeasured. Phase 3 ship gate is recall ≥ 0.80 on real samples per spec.md §8 D-7.
2. **Lite scanners** — pure-Python emulation of yara/plaso/volatility3/zeek output. Sufficient for architecture validation, NOT for Phase 3 ship.
3. **Solidity gates N/A** — `slither_high==0`/`aderyn_high==0` from gates.yaml apply only to Solidity. Recommend a `gates.yaml` patch (`applies_when: language=solidity`) before Phase 3 verifier so the report does not show false N/A noise.

## 3. Phase 3 Required Artifacts (gates.yaml `phase_3.required` `submission_artifacts_8`)

- `README.md`
- `docs/find-evil/architecture.md` (exists; expand for submission)
- `docs/find-evil/dataset.md` (exists; expand with SANS sources)
- `docs/find-evil/accuracy-report.md` (NEW — recall on real samples)
- `docs/find-evil/devpost-page.md` (NEW — pitch-writer)
- `docs/find-evil/demo-script.md` (NEW — pitch-writer, ≤5 min)
- `harness/find-evil/logs/` (structured trace dir — already exists empty)
- `LICENSE` (NEW — MIT pending Q4 confirmation)

## 4. Phase 3 Entry Actions (mission-control + dev-node)

| Action | Owner | Blocking |
|---|---|---|
| Mission-control: deliver Q1/Q2/Q4 responses (16d+ pending) | mission-control | hard-block — Phase 3 D+0 |
| Source SANS/ICS-CERT forensic dataset (≥3 cases, ground-truth labels) | evidence-curator | recall ≥ 0.80 ship gate |
| Install plaso/yara/volatility3/zeek in dev environment OR finalize Docker proxy spec | dev-node | real-tool recall measurement |
| Patch `gates.yaml` to scope solidity gates | mission-control or verifier | clean Phase 3 verify report |
| Pitch-writer Phase 3 lane: devpost-page + demo-script + README polish | find-evil-pitch-writer | submission artifacts gate |

Phase 3 entry command:
```
@find-evil-orchestrator "Phase 3 진입.
- 입력: spec.md §10 Phase 3, gates.yaml phase_3.required, plans/ITEM-212-find-evil/handover-r4.md.
- 목표: 8 submission artifacts + recall ≥ 0.80 on real samples + LICENSE 결정.
- 호출 순서: evidence-curator (SANS sample 통합) → detector-engineer (real-tool recall) → pitch-writer (devpost/demo/README) → verifier (별도 lane).
- 가정값 재확인: Q1 team, Q2 SIFT 호스팅, Q4 license — Q-회신 미수령 시 D+0 hard-block."
```

## 5. Risk Ledger Snapshot (forwarded)

| # | Risk | Status |
|---|---|---|
| 1 | SIFT subprocess wiring | RESOLVED (sift_runner.py + 7 handlers) |
| 2 | outputSchema 미노출 | RESOLVED (Phase 2 partial) |
| 3 | case.open 디스크 누적 | open — defer to maintenance UoW |
| 4 | Mission-control Q1/Q2/Q4 회신 16d+ | escalated — Phase 3 D+0 hard-block |
| 5 | tests/conftest.py R55-W1 충돌 | open — Phase 3에서 tests/find-evil/ 분리 권장 |
| 6 | Solidity gates N/A noise | open — gates.yaml patch 권고 |
| 7 | report.append SHA-256 race | open — single-process 가정 충분 (멀티-tenant OOS) |
| 8 (new) | Synthetic-only recall — Phase 3 ship gate (≥0.80 real) 미충족 위험 | active — evidence-curator + real binary 통합 필요 |
| 9 (new) | Lite scanners가 production code path와 분기 — 평가자가 "real tool path 검증되지 않음"으로 감점 가능 | active — Phase 3 accuracy-report.md에 분기 명확히 표기 |

## 6. Hand-back Signals

- Phase 2 final PASS evidence: `plans/ITEM-212-find-evil/verify-phase-2-full.md`.
- 핸드오프 카드 status update는 `@forrich-cross-node-sync` 단독 권한.
- Devpost / Discord / 외부 송출 일체 금지.
