---
name: bet-kickoff
description: Validate Bet readiness and confirm spec→contract→test→code chain exists. Gate before implementation start.
argument-hint: "<bet_id>"
level: 2
---

Validate Bet **{{PROMPT}}** readiness for implementation.

## Trigger

`/forge:bet-kickoff <bet_id>` or Bet state transition to Build phase

## Steps

1. **Read Bet + Pitch docs** via Obsidian-MCP or filesystem
   - Verify Bet has: appetite, problem statement, acceptance criteria, scope/hill chart
   - Verify Pitch has: design rationale, constraint summary
   - Check for: no_touch fields, escalation rules

2. **Confirm Obsidian-MCP connectivity**
   - Call devos_health_check() → must pass
   - Verify vault path accessible and readable

3. **Verify spec→contract→test→code chain**
   - Spec exists: FR/NFR/ADR files for this bet
   - Contract exists: OpenAPI, JSON Schema, or type stubs defined
   - Test scaffold exists: at least one test file created (even if empty)
   - Implementation skeleton exists OR package structure created

4. **Check no_touch and escalation rules**
   - Read Bet frontmatter: any no_touch fields listed?
   - Read Bet frontmatter: any escalation rules (HIGH → pause)?
   - Output: "All clear" or list of fields requiring care

5. **Run spec-auditor for initial gap scan**
   - Invoke spec-auditor agent to identify critical gaps
   - If HIGH gaps found: output "NOT_READY: fix gaps first"
   - If only MEDIUM/LOW gaps: output "READY with warnings"

6. **Output: READY / NOT_READY + missing pieces + next action**

```
# Bet Kickoff Report: {bet_id}

Status: READY ✓

Spec:      FR-123, NFR-456, ADR-789 ✓
Contract:  contracts/shapeup-api.openapi.json ✓
Tests:     tests/unit/lifecycle/ scaffold created ✓
Code:      src/context_devos/lifecycle/ package exists ✓

Gaps found by spec-auditor: MEDIUM-2, LOW-1 (non-blocking)

No-touch fields: auth.tokens (handle with care)
Escalation rules: if DB schema modified → pause for review

Next action: Start red phase (tdd-red agent) for first criterion
```

OR if NOT_READY:

```
Status: NOT_READY ✗

Missing:
- Contract file (OpenAPI) — create contracts/shapeup-api.openapi.json
- Test scaffold — mkdir -p tests/unit/lifecycle/

Critical gaps (HIGH-3):
- Bet.hill_chart field not in contract schema
- Error code E_APPETITE_EXCEEDED missing from error envelope
- Endpoint PUT /api/bets missing from API contract

Next action: Fix HIGH gaps first, then re-run bet-kickoff
```

## Rules

- Gate implementation until READY
- Critical gaps (HIGH) block → NOT_READY
- Warnings (MEDIUM/LOW) allow proceed → READY with caution
- Non-blocking missing pieces (test scaffold, package dir) can be auto-created
