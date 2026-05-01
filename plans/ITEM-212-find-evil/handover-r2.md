# ITEM-212 FIND EVIL — Phase 2 Handover (R+2)

- **Issued**: 2026-04-30
- **From**: dev-node Phase 1 session (`74b68cbb-27ab-44fc-a23b-c8ac02bc9360`)
- **To**: dev-node Phase 2 session (start ≥ 2026-05-21 per spec §10)
- **Verdict carried forward**: Phase 1 PASS (verify-phase-1.md)

## 1. Phase 1 Artifact Index

| Artifact | Path | Owner |
|---|---|---|
| Spec | `plans/ITEM-212-find-evil/spec.md` | Phase 0 |
| R+1 inbound | `plans/ITEM-212-find-evil/handover-r1.md` | Phase 0 |
| Status log | `plans/ITEM-212-find-evil/status.md` | continuous |
| Phase 1 verifier report | `plans/ITEM-212-find-evil/verify-phase-1.md` | Phase 1 verifier |
| Architecture | `docs/find-evil/architecture.md` | architect |
| Tool schemas (3) | `harness/find-evil/mcp-schema/{case.open,timeline.build,iocs.scan}.json` | architect |
| MCP server | `src/find_evil_mcp/{__init__.py,server.py,schema_loader.py,tools/*.py}` | implementer |
| Tests (27 PASS) | `tests/{test_schema.py,test_server.py}` | implementer |
| Conformance harness | `harness/find-evil/scripts/mcp_handshake_check.py` | dev-node |
| Gate runner | `harness/find-evil/Makefile` (build/test/mcp-conformance filled) | dev-node |
| pyproject (root) | `pyproject.toml` | implementer |

## 2. Phase 2 Entry Conditions

- All `gates.yaml.phase_1.required` green (re-verify with `make -C harness/find-evil spec-check build test mcp-conformance` before proceeding).
- `.venv-find-evil/` reusable (drop and rebuild if Python or `mcp` package version drifts).
- Mission-control Q1/Q2 still unconfirmed — Phase 1 ran on assumed `team=solo`, `SIFT=Docker proxy`. Phase 2 entry MUST re-check. If Q2 returns OVA-only, Docker compose stub becomes scrap.

## 3. First Phase 2 Command

```
@find-evil-orchestrator "Phase 2 진입.
- 입력: plans/ITEM-212-find-evil/spec.md §4 (남은 6 도구), gates.yaml phase_2.required.
- 목표: 9개 도구 schema 완비 + detector-engineer로 SIFT 호출 와이어링 +
  evil fixture 기반 recall ≥ 0.60.
- 호출 순서: architect (6 schemas 추가) → implementer (subprocess wrapping +
  outputSchema 검증) → detector-engineer (plaso/yara/volatility3/zeek 실호출) →
  evidence-curator (SANS sample + 자체 합성 evil fixture) → verifier (별도 lane).
- 가정값 재확인 필수: Q1 team, Q2 SIFT 호스팅, Q4 license."
```

## 4. Top Risks Forwarded

| # | Risk | Source | Phase 2 owner action |
|---|---|---|---|
| 1 | SIFT subprocess 미통합 — D-15 recall ≥ 0.60 게이트가 통합 실패를 첫 노출 | verify-phase-1.md | detector-engineer가 D-21 first-light 더미 fixture로 plaso/yara 호출 1회 성공 확보 |
| 2 | `outputSchema`가 mcp `Tool` 메타에 노출 안 됨 (현재 None) — 오케스트레이터 측 검증 누락 시 malformed JSON 통과 | implementer dump | implementer가 server.py에서 `outputSchema` 필드 채우거나 schema_loader.validate_output 강제 |
| 3 | `case.open`이 항상 `/cases/<id>/` 생성 → 다중 탐색 시 디스크 누적 | verifier 체크 | `case.list` / `--no-persist` 추가 또는 TTL 정책 설계 |
| 4 | Mission-control Q1/Q2/Q4 미회신 누적 14일 — Phase 2 빌드 windowing 어긋날 위험 | spec.md §9 | orchestrator가 Phase 2 D+0 hard-block에 회신 확인 게이트 추가 |
| 5 | conftest.py 덮어쓰기로 R55-W1 SaaS axis 테스트 collect 실패 (sqlalchemy 미설치) — find-evil 격리 venv 사용으로 우회 중 | git status | Phase 2에서 R55-W1 axis 합칠 때 conftest 분리 (예: `tests/find-evil/`) |

## 5. Risk / Kill Delta

- spec.md §8 D-15 게이트 (≥6 도구 + recall ≥ 0.60)는 Phase 2 외 다른 변경 없음.
- Phase 1에서 mission-control Q-회신 미확보 — kill 조건 발동은 아니지만 Q4(license) 미정 시 SPDX 헤더가 MIT 가정 상태로 잠금됨. Apache-2.0 결정 시 일괄 sed 필요.

## 6. Hand-back Signals (mission-control)

- Phase 1 PASS evidence: `plans/ITEM-212-find-evil/verify-phase-1.md`.
- 핸드오프 카드 status update는 `@forrich-cross-node-sync` 단독 권한.
- Devpost 어떤 형태로도 외부 송출 금지.
