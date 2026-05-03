---
applies_to: [warvis]
name: warvis-planner
description: >
  WARVIS stage 2 — Generate implementation plan from UoW spec. Calls devos_plan_dev_session.
  Writes .omc/plans/<uow_id>.md artifact. Reads project CLAUDE.md for commands.
  No approval gate except scope expansion beyond UoW boundaries. Works without Obsidian vault.
model: claude-sonnet-4-6
---

# warvis-planner

인라인 스펙 또는 볼트 SSOT에서 구현 계획을 생성하고 `.omc/plans/<uow_id>.md`에 저장한다.

## 인자

```
uow_id:         required
dev_session_id: required
project_id:     required
inline_spec:    optional (warvis-initiator handoff 또는 직접 전달)
no_go_scope:    optional — 구현 금지 영역
appetite_days:  optional — 예산
```

## 실행

### 1. 컨텍스트 로드

```bash
cat .omc/state/sessions/<dev_session_id>/init.md 2>/dev/null
```

`tech_stack`, `test_cmd`, `lint_cmd`, SSOT 내용 추출.

### 2. 코드베이스 영향 범위 파악 (빠른 탐색)

```bash
# 핵심 키워드로 관련 파일 찾기
grep -r "<uow_핵심_키워드>" src/ --include="*.py" --include="*.ts" --include="*.go" -l 2>/dev/null | head -15
ls src/ tests/ 2>/dev/null | head -20
```

### 3. 계획 파일 작성

`.omc/plans/<uow_id>.md`:

```markdown
# Plan: <uow_id> — <title>

## Context
- project_id: <id>
- tech_stack: <stack>
- ssot_source: inline | vault | plan_file

## Scope
- 생성: [파일 경로]
- 수정: [파일 경로]
- 금지 (no_go_scope): [항목]

## Milestones

### M1: <이름>
- Tasks:
  - [ ] <작업 (TDD: 테스트 먼저)>
- Validation: `<test_cmd> <test_path> -v`
- Risk: LOW | MEDIUM | HIGH

### M2: <이름>
...

## Stop-and-fix Rule
마일스톤 검증 실패 → 다음 마일스톤 이동 금지. 수정 후 재검증.

## Done when
- [ ] 모든 마일스톤 검증 통과
- [ ] `<lint_cmd>` 오류 없음
- [ ] devos_verify_dev_session PASS

## Status
| M | 상태 | 완료 | 증거 |
|---|------|------|------|
| M1 | 대기 | - | - |
```

### 4. devos 계획 등록

```
devos_plan_dev_session({
  project_id,
  dev_session_id,
  uow_id,
  task_specs: [{ task_id, title, action_statement }],
  milestone_strategy: ["M1: ...", "M2: ..."],
  verification_strategy: ["<test_cmd>", "<lint_cmd>"]
})
```

devos 실패 시: 로컬 계획 파일로 계속. 세션 없이도 구현 가능.

## 계획 원칙

- **최소 변경**: 요청 범위만. 추가 리팩터링 금지.
- **TDD 순서**: 테스트 먼저 → 구현 → 통과 확인.
- **마일스톤당 최대 30분** 예상 작업.

## 스코프 증가 판단 (유일한 게이트)

아래 경우에만 사용자 확인:
- 다른 모듈/서비스 수정 필요
- 데이터 마이그레이션 수반
- 공개 API 계약 변경

단순 테스트 추가, 관련 문서 업데이트 = 스코프 증가 아님. 자동 진행.

## 출력

```json
{
  "status": "succeeded",
  "dev_session_id": "<id>",
  "plan_path": ".omc/plans/<uow_id>.md",
  "milestone_count": 0,
  "risk_level": "LOW|MEDIUM|HIGH",
  "test_cmd": "<command>",
  "lint_cmd": "<command>"
}
```
