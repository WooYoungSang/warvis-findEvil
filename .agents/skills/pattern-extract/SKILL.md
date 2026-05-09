---
name: pattern-extract
description: Extract learnings from completed Bet and recommend patterns for promotion (agent, skill, hook, or core).
argument-hint: "<bet_id>"
level: 3
---

Extract patterns from completed **{{PROMPT}}** for reuse.

## Trigger

`/forge:pattern-extract <bet_id>` at Reflect phase (after ship)

## Steps

1. **Read Journal + Lesson entries** for bet
   - Lessons written during dev (insights, blockers, workarounds)
   - Retrospective notes (what went well, what was hard)
   - Failed approaches and why they didn't work

2. **Read Hill Chart change history**
   - How did completion % evolve day-by-day?
   - Which phases took longest?
   - Were there unexpected jumps or plateaus?

3. **Read failed tests / retries / manual corrections**
   - Which tests failed repeatedly?
   - Which implementations required multiple attempts?
   - Were there edge cases discovered in testing?

4. **Classify each pattern**

   - **promote-to-core**: Repeatable utility needed by multiple bets
     - Example: Hill Chart computation, acceptance criteria parser
     - Action: Add to libs/ or src/context_devos/

   - **promote-to-skill**: Workflow that other bets will repeat
     - Example: "check payment vendor readiness" (only payment bets need this)
     - Action: Create new skill under .claude/skills/

   - **promote-to-agent**: Specialized role that can be parameterized
     - Example: "payment-vendor-integrator" agent (other projects may need variant)
     - Action: Create new agent under .claude/agents/

   - **promote-to-hook**: Automatic gate or check that should run proactively
     - Example: "warn if Hill Chart executing > 7 days without progress"
     - Action: Add to .claude/hooks/

   - **keep-local**: Specific to this bet, not reusable
     - Example: Workaround for payment vendor's API quirk
     - Action: Document in Lesson as reference only

   - **discard**: Anti-pattern, doesn't apply elsewhere
     - Example: Temporary debugging code, vendor-specific hack
     - Action: Note why and move on

5. **Output: retrospective block + promotion list**

```markdown
# Pattern Extraction: {bet_id}

**Bet**: Shape Up Bet for Payment Integration
**Completed**: 2026-04-21
**Duration**: 2 weeks (14 days)

## Key Learnings

### Hill Chart Evolution

Phase | Days spent | Note
------|-----------|------
Research | 2 | Vendor API docs incomplete
Executing | 8 | Implementation straightforward, test flakes caused delays
Waiting | 3 | Vendor approval gate (expected)
Landed | 1 | Quick polish + final integration test

→ **Learning**: Vendor integration requires 2-day buffer for approvals. Plan accordingly for future.

### Failed Tests + Iterations

| Test | Failures | Root cause | Resolution |
|------|----------|-----------|-----------|
| test_payment_webhook_signature | 3 | Vendor changed signing algo | Updated test + added vendor version check |
| test_refund_idempotency | 2 | Missing idempotent key in request | Added generate_idempotent_key() utility |

→ **Learning**: Vendor APIs are fragile. Build abstraction layer (payment-gateway adapter) to isolate vendor changes.

### Blockers

- Day 5: Vendor API doc bug (vendor fixed on day 6)
- Day 9: Test infrastructure didn't support concurrent payment tests (added mutex)

## Patterns for Promotion

| Pattern | Type | Priority | Description | Target |
|---------|------|----------|-------------|--------|
| Hill Chart idle detect | hook | HIGH | Alert if Hill Chart executing > 7 days with <5% progress | .claude/hooks/ |
| Payment gateway adapter | core | HIGH | Abstraction layer for vendor API changes | src/context_devos/payment/ |
| Idempotent key generation | core | MEDIUM | Utility for vendor integrations requiring idempotency | libs/canonical-domain |
| Vendor readiness check | skill | MEDIUM | Workflow for validating vendor prerequisites | .claude/skills/vendor-readiness-check/ |
| Concurrent test mutex | core | MEDIUM | Fixture for tests that can't run in parallel | tests/fixtures/ |
| Payment-vendor-integrator | agent | LOW | Variant agent for payment vendor work (future) | not-created-yet (proposal) |

## Next Actions

1. Create hook: Hill Chart idle detection (3 hours)
2. Refactor: Extract payment gateway adapter to libs (4 hours)
3. Create skill: Vendor readiness check (2 hours)
4. Proposal: Payment-vendor-integrator agent (link to this report)

---

Alternatively, if no patterns worth promoting:

## Patterns for Promotion

None. This bet was highly specific to vendor requirements and doesn't generalize.

All learnings documented in Lesson files for reference in future payment-related work.
```

## Rules

- Classify patterns objectively: would another bet reuse this?
- Prioritize promotions: HIGH (blocks other bets), MEDIUM (nice-to-have), LOW (future consideration)
- Be specific: what would the promoted artifact be called and where would it live?
- If many patterns to promote, prioritize top 3-5 for immediate action
