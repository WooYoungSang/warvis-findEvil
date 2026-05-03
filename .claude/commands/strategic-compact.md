# strategic-compact

Audit the active conversation for context rot and decide whether to compact now.

## Protocol

1. Score urgency (0-10): stale facts, repeated reads, completed tasks still in context
   - 0-3: healthy, continue
   - 4-6: compact if starting a new workstream
   - 7-10: compact now

2. If score >= 7, write a snapshot to .omc/state/compact-<timestamp>.json:

```
## Context Snapshot
### Active task
### Key decisions (non-obvious only)
### Files modified
### Next step
### Do NOT re-derive
```

3. Recommend /clear or continue.
