---
name: spec-gap-check
description: Scan spec→contract divergence and report structural gaps with remediation priority.
argument-hint: "[scope] [--contracts path/to/contracts]"
level: 2
---

Check spec↔contract alignment for **{{PROMPT}}**.

## Trigger

`/forge:spec-gap-check [scope]` or on spec update

## Steps

1. **Identify scope**: Default to active bet scopes; override with argument
   - Scope can be: bet_id (S20-005), UoW_id (V3-003), or module path (src/lifecycle/)

2. **Read FR/NFR/ADR for scope**
   - Find all spec files matching scope
   - Extract: required methods, fields, constraints, error codes
   - Build spec model (expected surface)

3. **Read contracts** (JSON Schema, OpenAPI, Python types)
   - Find all contract files matching scope
   - Extract: type definitions, endpoint signatures, error envelopes
   - Build contract model (actual surface)

4. **Diff spec intent vs contract surface**
   - Missing endpoint? (spec lists it, contract doesn't)
   - Missing field? (spec requires it, contract lacks it)
   - Type mismatch? (spec: int, contract: string)
   - Constraint violation? (spec: max 100, contract: max 1000)

5. **Output: structured gap table**

```markdown
# Spec Gap Check: {scope}

**Date**: {timestamp}
**Spec files**: FR-123 (endpoint), NFR-456 (performance)
**Contract files**: contracts/api.openapi.json, src/models/bet.py

## Gaps Found: {n}

| Gap | Severity | Spec Ref | Contract | Recommendation |
|-----|----------|----------|----------|-----------------|
| POST /api/bets missing | HIGH | FR-123:L12 | contracts/api.openapi.json lacks path | Add paths."/bets".post to contract |
| Bet.appetite missing | MEDIUM | FR-123:L34 | Bet schema has no appetite field | Add appetite: integer to schema |
| Error E_APPETITE_EXCEEDED | HIGH | NFR-456:errors | components/schemas missing this code | Add E_APPETITE_EXCEEDED to error enum |
| Performance SLA | LOW | NFR-456:L8 max 100ms | Contract unspecified | Add x-performance-sla: 100 annotation |

## Summary

- HIGH: 2 (blocks implementation)
- MEDIUM: 1 (needs refinement)
- LOW: 1 (nice-to-have)

## Next Action

Fix HIGH gaps before implementation. Gap-check again after contract update.
```

## Rules

- List each gap once
- Severity: HIGH (blocks), MEDIUM (breaks), LOW (improvement)
- Cite line numbers in spec files
- Be specific: quote exact requirement vs actual contract
