# Open Questions — ITEM-212 FIND EVIL

## Phase 3 (Go Bridge) — 2026-05-02

### Pre-Execution Consensus (needed before Day 1)
- [ ] **Ollama Endpoint** — Is `localhost:11434` the correct endpoint, or should we support remote via flag?
  - Why: Affects network security model + test setup
- [ ] **Gemma 4 Model Name** — Confirm exact model name for Ollama (gemma4:26b-a4b-instruct-q4_K_M or different)?
  - Why: Affects system prompt injection + tool-calling tier selection (native vs Tier 2/3)
- [ ] **Test Fixture Path** — Use Phase 2's `/evidence/test.img`, or create fresh minimal fixture?
  - Why: Affects Day 3 kill-switch test setup + reproducibility
- [ ] **MCP Spawn Command** — Exact Python entry point (`python -m find_evil_mcp.server` or installed CLI)?
  - Why: Affects subprocess spawning + error handling in mcp/transport.go
- [ ] **Go Module Path** — Root-level `warvis/` or subdirectory?
  - Why: Affects git history + import paths + CI/CD setup

### Architecture Clarifications
- [ ] **Tier 1 Tool Calling Support** — Should we probe Ollama at startup for native tools support, or assume Tier 3 fallback?
  - Why: Tier 1 (if available) is cleaner, but Tier 3 always works. Startup probe adds complexity but automates fallback.
- [ ] **MCP Server Lifecycle** — Should Go bridge spawn Python server once per hunt, or maintain long-lived connection?
  - Why: Affects resource cleanup + error recovery + test isolation

### Phase 4 Prerequisites (decision points at kill-switch evaluation)
- [ ] **If kill-switch FAILS**: Accept Honorable Mention placement + Python-only submission, or pivot to alternate Go-based approach?
  - Why: Determines fallback strategy + timeline for Phase 4 artifacts
- [ ] **Demo Video Scope** — Should demo show full hunt (INITIALIZE→LOCK) or just INITIALIZE→TRACE?
  - Why: Full hunt may exceed 5-min time limit; TRACE demo may not demonstrate autonomy sufficiently

---

## Phase 2 (Completed) — 2026-04-30

### Resolved (PASS)
- ✓ MCP server skeleton + 9 tools + test harness
- ✓ Detector integration (7 handlers wired)
- ✓ Fixtures + recall harness (100% recall on synthetic)

### Deferred to Phase 3 (depends on Go bridge)
- [ ] Real-world fixture testing (beyond synthetic evil)
- [ ] End-to-end demo with actual SIFT toolchain
- [ ] Accuracy report on forensic evidence corpus

