---
name: ship-or-cut
description: At 75% appetite, evaluate completion vs Hill Chart and decide SHIP / CUT / SCOPE_HAMMER.
argument-hint: "<bet_id>"
level: 3
---

Evaluate ship readiness for **{{PROMPT}}** at 75% appetite.

## Trigger

`/forge:ship-or-cut <bet_id>` or manual gate at 75% elapsed days

## Steps

1. **Read Bet appetite and elapsed time**
   - Bet appetite: {n} weeks
   - Elapsed: {n} weeks
   - Days remaining: {n}

2. **Read Hill Chart positions** for all scopes in bet
   - Which scopes are landed? (no more work needed)
   - Which scopes are executing? (work ongoing)
   - Which scopes are stuck? (blocked or no-go)

3. **Evaluate acceptance criteria completion %**
   - For each must-have criterion: is it satisfied?
   - Calculate: {satisfied} / {total} × 100%
   - Extract evidence: test pass, code review, demo

4. **Check kill conditions**
   - Any kill conditions triggered? (e.g. "if vendor unblocks after day 8, cut scope")
   - If triggered: is this a SHIP-without or a BLOCK?

5. **Output verdict with evidence**

```markdown
# Ship or Cut Decision: {bet_id}

**Appetite**: {n} weeks
**Elapsed**: {n} weeks (75%)
**Days remaining**: {n}
**Decision**: SHIP / CUT / SCOPE_HAMMER

## Acceptance Criteria Status

| Criterion | Must-have | Status | Evidence |
|-----------|-----------|--------|----------|
| Auth login | YES | ✓ DONE | 23 tests pass |
| Payment processing | YES | ✓ DONE | 8 e2e tests pass |
| Admin dashboard | NO | ~ PARTIAL | 15/20 features complete |

**Completion**: 2/2 must-haves = 100% ✓

## Hill Chart Status

| Scope | Days spent | Phase | Risk |
|-------|-----------|-------|------|
| Auth | 4 | landed | CLEAR |
| Payment | 8 | landed | CLEAR |
| Admin UI | 6 | executing | CAUTION (40% appetite left, 60% work remains) |

## Kill Conditions

- "If vendor unblocks before day 8" → NOT triggered ✓
- "If auth tests drop below 90%" → NOT triggered ✓

## Verdict

### SHIP ✓

All must-haves landed. 75% appetite elapsed. Ready to ship.

Optional: Admin UI can move to next bet if timeline is tight.

---

### OR CUT ✗

1/2 must-haves still executing. Cannot land in time.

**Recommendation**: Cut lower-priority scope (Admin UI) and ship core (Auth + Payment).

---

### OR SCOPE_HAMMER ⚠

All must-haves done, but Admin UI incomplete. Appetite running out.

**Options**:
1. Ship without Admin UI (reduce scope to Auth + Payment)
2. Extend appetite (adds 1+ week, delay next bet)
3. Defer Admin UI to next bet

**Recommended**: Option 1 (ship core, defer polish)
```

## Rules

- SHIP: 100% must-haves, Hill Charts landed/near-landed, no kill blocks
- CUT: Critical must-haves missing, cannot finish in time, recommend reduced scope
- SCOPE_HAMMER: Must-haves done but nice-to-haves incomplete; recommend cutting optional scope to ship on time
