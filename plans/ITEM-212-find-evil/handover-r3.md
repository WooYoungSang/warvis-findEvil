# ITEM-212 FIND EVIL — Phase 2 Step 3 / Phase 3 Handover (R+3)

- **Issued**: 2026-04-30
- **From**: dev-node Phase 2 partial session (`874d4ed7-5825-4280-b913-d2e41e710212`)
- **Verdict**: PASS_WITH_DEFERRED (verify-phase-2-partial.md)
- **Scope completed this session**: Phase 2 Step 1 (6 schemas) + Step 2 (handlers + outputSchema). Detector + fixtures + recall measurement DEFERRED.

## 1. State of the Build

| Asset | Path | Status |
|---|---|---|
| MCP schemas (9) | `harness/find-evil/mcp-schema/*.json` | complete (3 P1 + 6 P2) |
| Tool handlers (9) | `src/find_evil_mcp/tools/*.py` | complete; subprocess stubs only |
| Server registration | `src/find_evil_mcp/server.py` | 9 tools w/ outputSchema |
| Schema loader | `src/find_evil_mcp/schema_loader.py` | `get_output_schema()` exposed |
| Conformance harness | `harness/find-evil/scripts/mcp_handshake_check.py` | EXPECTED_TOOLS = 9 |
| Tests | `tests/test_schema.py` (39), `tests/test_server.py` (17) | 56/56 PASS |
| Architecture doc | `docs/find-evil/architecture.md` | 9-tool handshake snippet |
| Verify reports | `plans/ITEM-212-find-evil/verify-phase-{1,2-partial}.md` | both PASS-grade |

## 2. Outstanding Phase 2 Work (DEFERRED)

| Item | Owner | Effort | Blocking gate |
|---|---|---|---|
| Detector subprocess wiring (plaso, yara, volatility3, zeek/suricata, log2timeline, sigma) | `find-evil-detector-engineer` | 3-4h | `evil_detection_recall_min` |
| Curated evil fixtures (SANS samples + synthesised memory/pcap/log) under `repos/find-evil-fixtures/` | `find-evil-evidence-curator` | 2-3h | `evil_detection_recall_min` |
| Recall/precision measurement harness — `make accuracy` | dev-node + detector-engineer | 1h | `evil_detection_recall_min` |
| `slither_high == 0`, `aderyn_high == 0` — gate is N/A for Python project; verifier flagged as such. **Action**: verifier proposes amending gates.yaml in Phase 3 to gate solidity-only, OR keep as N/A with explicit doc note. | verifier | 0.5h |  |

Detector wiring entry command:
```
@find-evil-detector-engineer "Phase 2 Step 3 detector wiring.
- subprocess whitelist: plaso, log2timeline, yara, volatility3, zeek, suricata, sigma.
- 입력: 기존 9 handler stub들 (src/find_evil_mcp/tools/*.py) — TODO 마커 따라 SIFT 호출 채울 것.
- SIFT 호스트: harness/find-evil/docker-compose.yml (Q2 가정값).
- 게이트: tests/test_server.py 회귀 + 신규 detector_*.py 통합 테스트로 placeholder fixture 1개 통과.
- Out-of-scope: real fixtures (evidence-curator). 합성 mini fixture 1개로 wiring smoke만."
```

## 3. Phase 3 Entry Conditions

Per spec.md §10 Phase 3 (2026-06-02 → 2026-06-08): demo video, accuracy report, architecture diagram, README. Submission package draft + recall ≥ 0.80 ship-or-cut.

Required before Phase 3 entry:
- Phase 2 deferred 4건 모두 PASS (recall ≥ 0.60).
- Mission-control Q1/Q2/Q4 회신 확보 (이번 세션도 미수신 — 14일+ 누적).

## 4. Risks Forwarded

| # | Risk (carried) | New status | Action |
|---|---|---|---|
| 1 | SIFT subprocess 미통합 | unchanged — DEFERRED | Phase 2 Step 3 (위) |
| 2 | outputSchema 미노출 | **RESOLVED** in this session | regression test 추가됨 (`test_server_tools_have_output_schema`) |
| 3 | case.open 디스크 누적 | unchanged | Phase 2 Step 3 또는 별도 maintenance UoW |
| 4 | mission-control Q-회신 14일+ | **escalated** | orchestrator가 Phase 3 D+0 hard-block에 강제 |
| 5 | tests/conftest.py R55-W1 충돌 | unchanged | tests 분리 별도 UoW |
| 6 (new) | gates.yaml `slither_high`/`aderyn_high`이 Python 프로젝트에 N/A — 평가자 혼동 가능 | new | verifier 제안: gates.yaml에 `applies_when: has_solidity` 조건 추가 또는 docs/find-evil/architecture.md에 N/A 명시 |
| 7 (new) | report.append hash chain은 SHA-256 단방향 — 다중-orchestrator 동시 append 시 race | new | 단일 process 가정으로 충분; 멀티-tenant는 OOS |

## 5. Hand-back Signals

- Phase 2 partial PASS evidence: `plans/ITEM-212-find-evil/verify-phase-2-partial.md`.
- 핸드오프 카드 status update는 `@forrich-cross-node-sync` 단독 권한.
- Devpost 외부 송출 금지.
