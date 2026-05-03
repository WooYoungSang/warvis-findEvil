# Plan: ITEM-212-day4-ollama-agent

## Harness
- config: `.omc/plans/ITEM-212-day4-harness.md`
- test_cmd: `cd warvis && go test ./internal/agent/... -v`
- lint_cmd: `cd warvis && go vet ./... && golangci-lint run ./internal/agent/...`
- reused_skills: None (novel Gemma 4 integration)
- strategy: Direct TDD (red→green→refactor), no subagent delegation

## Scope
- Create: pkg/ollama/client.go, pkg/ollama/types.go, internal/agent/{prompt,tool_call,loop,types}.go
- Modify: warvis/go.mod (add context if needed), warvis/cmd/warvis/main.go (agent loop dispatch)
- Frozen: src/find_evil_mcp/ (Python, Days 1-2 complete)

## Milestones

### M4a: Ollama HTTP Client
- Tasks:
  - Write `pkg/ollama/client.go`: Chat endpoint POST to localhost:11434/api/chat
  - Implement `Chat(ctx, model, messages)` → response with message.content
  - Add retry logic (3 attempts on network timeout)
  - Unit test with mocked HTTP server
- Validation: `go test ./pkg/ollama -v`
- Risk: MEDIUM (HTTP client, network dependency on Ollama availability check)

### M4b: Tool Schema Injection
- Tasks:
  - Write `internal/agent/prompt.go`: BuildSystemPrompt(state, tools)
  - System prompt template: Hunt protocol, current state, available tools, JSON format spec
  - Tool schema: name, description, required_fields for each allowed tool in state
  - JSON format: `{ "action": "call_tool|state_complete|escalate", "tool_name": "...", "arguments": {...}, "reason": "..." }`
- Validation: `go test ./internal/agent -run TestPrompt`
- Risk: LOW (pure string building, no I/O)

### M4c: Tool-Call Parsing (Tier 3)
- Tasks:
  - Write `internal/agent/tool_call.go`: ParseAction(response string) → (*Action, error)
  - Extract JSON from Gemma response (may be wrapped in markdown, etc.)
  - Parse into Action struct: Type, ToolName, Arguments, Reason
  - 3-retry loop: malformed JSON → retry with error context feedback
  - After 3 failures: return escalate action (not error, not crash)
- Validation: `go test ./internal/agent -run TestParse` (3 retry scenarios, markdown wrapping)
- Risk: MEDIUM (JSON parsing, error recovery, retry logic)

### M4d–M4f: Agent Loop Skeleton
- Tasks:
  - Write `internal/agent/loop.go`: Run(ctx context.Context) error
  - Pseudo-loop:
    1. Get current FSM state + budget + allowed tools
    2. Build system prompt (M4b) + history window
    3. Call Ollama Chat (M4a)
    4. Parse action (M4c, 3 retries)
    5. Switch on action.Type:
       - call_tool: validate state-allowed, dispatch to MCP (M4e), sanitize output, append to history
       - state_complete: check budget, transition FSM (return success)
       - escalate: log escalation, transition to expose state (safety fallback)
    6. Repeat until state_complete or escalate
  - Conversation history: sliding window 20 exchanges, truncate tool outputs to 500 chars
  - Set timeout: 600s per state (configurable)
- Validation: `go test ./internal/agent/... -v -cover` (60%+ coverage)
- Risk: HIGH (complex state machine, tool validation, history management, budget enforcement)

## Done when
- [x] Harness config created
- [ ] M4a: Ollama client tests pass (mock HTTP)
- [ ] M4b: System prompt tests pass (schema validation)
- [ ] M4c: Tool-call parsing tests pass (3 retry scenarios)
- [ ] M4d–M4f: Agent loop tests pass (60%+ coverage, mock FSM + MCP)
- [ ] `golangci-lint run ./pkg/ollama ./internal/agent` passes (0 issues)
- [ ] `go build -o bin/warvis ./cmd/warvis` succeeds
- [ ] Ollama handshake verified (gemma4:26b model available)
- [ ] No crash on malformed JSON (3 retries, escalate action triggered)

## Status
Started: 2026-05-05 (Sunday, Day 4)
Target completion: 2026-05-05 (same day, 8–12 hours)
