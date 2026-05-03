# codex advisor artifact

- Provider: codex
- Exit code: 0
- Created at: 2026-05-01T15:43:12.176Z

## Original task

Review the W.A.R.V.I.S Go bridge architecture design at docs/find-evil/warvis-go-architecture.md. This is a Go binary that bridges Gemma 4 26-A4B (via Ollama) to a Python MCP server (find-evil-mcp) for autonomous DFIR investigations. Evaluate: (1) Go module structure soundness, (2) Hunt State Machine FSM design, (3) Gemma 4 tool-calling strategy via JSON parsing, (4) MCP JSON-RPC 2.0 stdio client design, (5) security boundary coverage. Identify any gaps, anti-patterns, or improvements. Be specific and technical.

## Final prompt

Review the W.A.R.V.I.S Go bridge architecture design at docs/find-evil/warvis-go-architecture.md. This is a Go binary that bridges Gemma 4 26-A4B (via Ollama) to a Python MCP server (find-evil-mcp) for autonomous DFIR investigations. Evaluate: (1) Go module structure soundness, (2) Hunt State Machine FSM design, (3) Gemma 4 tool-calling strategy via JSON parsing, (4) MCP JSON-RPC 2.0 stdio client design, (5) security boundary coverage. Identify any gaps, anti-patterns, or improvements. Be specific and technical.

## Raw output

```text
**Recommendation: Request Changes**

No implementation blocker is impossible, but the design should be tightened before coding. The biggest risks are MCP protocol correctness, unbounded LLM control flow, weak path/process isolation details, and some Go package layering ambiguity.

**High Severity Findings**

1. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:640) — MCP lifecycle is incomplete.

The design sends `initialize`, then immediately `tools/list`, but MCP requires protocol version negotiation and an `notifications/initialized` notification after initialization. The request sketch also omits `protocolVersion` and client capabilities. Current MCP lifecycle docs require this sequence, and stdio messages must be newline-delimited valid JSON-RPC only.

Fix: model the full MCP handshake explicitly: `initialize` with protocol version, validate returned protocol version/capabilities, send `notifications/initialized`, then enter operation mode. Also support notifications, server errors, and version mismatch disconnect.

Sources: MCP lifecycle and stdio transport docs: https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle, https://modelcontextprotocol.io/specification/2025-03-26/basic/transports

2. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:730) — The stdio transport design is fragile.

`bufio.Scanner` has a default token limit, so large MCP responses can fail. `Recv(ctx)` starts a goroutine per receive; if the context cancels while `Scan()` blocks, the goroutine leaks and may race future reads. `Close()` closes stdin and waits, but does not drain stderr/stdout, check `Scanner.Err()`, or kill on timeout.

Fix: use one dedicated reader goroutine, `bufio.Reader.ReadBytes('\n')` or a scanner with explicit large buffer, demux responses by JSON-RPC id, handle notifications separately, capture stderr, and implement graceful shutdown with kill-after-timeout.

3. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:441) — Gemma function-calling assumptions are stale.

The doc assumes Gemma 4 does not have native function-calling-like support and uses prompt-only JSON parsing. Current Google docs describe Gemma 4 tool/function calling using tool schemas and tool-call formatting, and Ollama supports a `tools` field for supported models plus structured output via `format`. JSON prompting remains a valid fallback, but it should not be the primary unchecked path.

Fix: make the strategy layered:

1. Prefer Ollama `tools` if the selected Gemma 4 model exposes tool calls through Ollama.
2. Prefer Ollama `format` with a strict JSON Schema for the `Action` object when tool calls are unavailable.
3. Fall back to prompt-only JSON parsing only after probing model behavior at startup.

Sources: Gemma 4 function calling: https://ai.google.dev/gemma/docs/capabilities/text/function-calling-gemma4, Ollama tools: https://ollama.com/blog/tool-support, Ollama structured outputs: https://ollama.com/blog/structured-outputs

4. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:540) — Agent loop has no hard step budget despite retry policy.

Invalid JSON loops forever in the sample. Tool errors can also loop indefinitely because the code appends feedback and continues without enforcing the retry table. A forensic agent needs deterministic stop conditions.

Fix: add per-state budgets: max LLM turns, max invalid JSON attempts, max invalid args per tool, max repeated same tool+args, max state duration, and max total tool calls. Persist these counters in state so resume does not reset abuse loops.

5. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:939) — Path validation is insufficient for the stated security boundary.

`filepath.Abs()` plus `strings.Contains(abs, "..")` is not enough. `Abs` cleans paths, symlinks can escape, and the code does not enforce `/evidence/` or `/cases/<case_id>/`. It also conflicts with the table allowing relative evidence paths.

Fix: use `filepath.Clean`, `EvalSymlinks` where safe, strict allow roots, `filepath.Rel` containment checks, reject relative paths unless explicitly resolved under a configured evidence root, and re-check after file open if symlink races matter.

**Medium Severity Findings**

6. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:340) — Go module/package structure is directionally sound but internally inconsistent.

The tree has `internal/hunt/fsm.go`, `state.go`, `transitions.go`, but the matrix refers to packages like `hunt/fsm`, `hunt/state`, and `hunt/transitions`. In Go, those are separate packages only if they are separate directories. Also, `hunt/fsm -> display` creates a UI dependency inside domain logic.

Fix: either keep one `internal/hunt` package or split actual directories. Prefer domain packages depending on interfaces: `EventSink`, `StateStore`, `ToolExecutor`. Keep terminal display at the edge.

7. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:430) — “All packages ↔ config” is a global dependency smell.

Config should not be imported everywhere. It makes tests harder and couples protocol/domain code to runtime environment.

Fix: load config in `cmd/warvis`, pass typed options into constructors, and keep pure packages unaware of env/flags.

8. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:123) — FSM success criteria are too easy to satisfy.

TRACE can complete with only 10 parsed events, regardless of coverage or suspiciousness. EXPOSE accepts confidence `>=0.5`, and verification timeout records provisional confidence `0.5`, which can satisfy the threshold by default. That risks turning incomplete evidence into “verified enough.”

Fix: separate `complete`, `inconclusive`, `degraded`, and `failed` outcomes. Require explicit evidence coverage metrics and distinguish “no finding” from “unable to analyze.”

9. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:301) — Transition matrix is confusing and partly inconsistent.

The matrix labels “Current / Next” ambiguously, lists “RECURSIVE” for non-recursive phase changes, and includes `SCAN -> TRACE` in text without clearly showing its guard. This will lead to implementation drift.

Fix: specify transitions as typed rules: `From`, `To`, `Guard`, `Action`, `MaxCount`, `TerminalStatus`.

10. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:589) — Tool output is fed back as a user message without untrusted-data isolation.

Forensic logs and tool outputs are attacker-controlled input. Feeding them directly into the model as normal user content invites prompt injection.

Fix: wrap tool output in a dedicated tool-result message if using Ollama tools, or use a strict envelope saying “untrusted forensic data, never instructions,” with truncation, escaping, and schema extraction before summarization.

11. [src/find_evil_mcp/server.py](/home/jang/Workspace/warvis-findEvil/src/find_evil_mcp/server.py:314) — The current Python server returns `str(result)`, not JSON.

The Go architecture expects JSON-ish MCP content, but `str(dict)` produces Python literals with single quotes. That will break robust JSON parsing and audit replay.

Fix either side of the contract: make the Python MCP return `json.dumps(result)` or make the Go client treat tool content as opaque text and only parse when `Content-Type`/schema says JSON.

**Security Boundary Gaps**

The design has good first principles: per-state tool whitelist, MCP schema validation, no shell tools, case sandbox, and audit logging. The missing pieces are operational hardening:

- Hash original evidence at ingest and store immutable evidence metadata before analysis.
- Run the MCP subprocess with constrained env, working directory, resource limits, and preferably container/user isolation.
- Redact or classify audit log fields; raw args/results may contain sensitive host paths, usernames, IPs, or secrets.
- Define append-only audit/report integrity, not just `FindingHash`.
- Add cancellation semantics for long SIFT subprocesses rather than relying on Go context alone.
- Treat Ollama as a local network boundary: bind localhost, configurable allowlist, no remote model endpoint unless explicitly enabled.

**Overall**

The architecture is promising and modular, but the implementation plan should be revised around explicit contracts: a real MCP lifecycle client, strict LLM action schema, bounded FSM execution, hardened path/process isolation, and clean Go dependency boundaries. The highest-value next edit is to replace the pseudo-code client/loop sections with concrete stateful protocols and budgets before writing the binary.


WARNING: failed to clean up stale arg0 temp dirs: Permission denied (os error 13)
Reading additional input from stdin...
OpenAI Codex v0.128.0 (research preview)
--------
workdir: /home/jang/Workspace/warvis-findEvil
model: gpt-5.5
provider: openai
approval: never
sandbox: danger-full-access
reasoning effort: none
reasoning summaries: none
session id: 019de433-e8ea-7751-b0d9-61350e6c6afb
--------
user
Review the W.A.R.V.I.S Go bridge architecture design at docs/find-evil/warvis-go-architecture.md. This is a Go binary that bridges Gemma 4 26-A4B (via Ollama) to a Python MCP server (find-evil-mcp) for autonomous DFIR investigations. Evaluate: (1) Go module structure soundness, (2) Hunt State Machine FSM design, (3) Gemma 4 tool-calling strategy via JSON parsing, (4) MCP JSON-RPC 2.0 stdio client design, (5) security boundary coverage. Identify any gaps, anti-patterns, or improvements. Be specific and technical.
2026-05-01T15:41:41.115841Z ERROR rmcp::transport::worker: worker quit with fatal: Transport channel closed, when Client(HttpRequest(HttpRequest("http/request failed: error sending request for url (http://127.0.0.1:8765/mcp)")))
hook: SessionStart
hook: SessionStart Completed
hook: UserPromptSubmit
hook: UserPromptSubmit Completed
codex
I’ll review the architecture doc as a design review: concrete findings first, scoped to the five areas you named, with file/section references and residual risks.
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "sed -n '1,220p' /home/jang/.agents/skills/code-review/SKILL.md && pwd && rg --files | rg '("'^|/)docs/find-evil/warvis-go-architecture'"\\.md"'$|(''^|/)go'"\\.mod"'$|(''^|/)AGENTS'"\\.md"'$'"'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
---
name: code-review
description: Run a comprehensive code review
---

# Code Review Skill

Conduct a thorough code review for quality, security, and maintainability with severity-rated feedback.

## When to Use

This skill activates when:
- User requests "review this code", "code review"
- Before merging a pull request
- After implementing a major feature
- User wants quality assessment

## What It Does

## GPT-5.4 Guidance Alignment

- Default to concise, evidence-dense progress and completion reporting unless the user or risk level requires more detail.
- Treat newer user task updates as local overrides for the active workflow branch while preserving earlier non-conflicting constraints.
- If correctness depends on additional inspection, retrieval, execution, or verification, keep using the relevant tools until the review is grounded.
- Continue through clear, low-risk, reversible next steps automatically; ask only when the next step is materially branching, destructive, or preference-dependent.

Delegates to the `code-reviewer` agent (THOROUGH tier) for deep analysis:

1. **Identify Changes**
   - Run `git diff` to find changed files
   - Determine scope of review (specific files or entire PR)

2. **Review Categories**
   - **Security** - Hardcoded secrets, injection risks, XSS, CSRF
   - **Code Quality** - Function size, complexity, nesting depth
   - **Performance** - Algorithm efficiency, N+1 queries, caching
   - **Best Practices** - Naming, documentation, error handling
   - **Maintainability** - Duplication, coupling, testability

3. **Severity Rating**
   - **CRITICAL** - Security vulnerability (must fix before merge)
   - **HIGH** - Bug or major code smell (should fix before merge)
   - **MEDIUM** - Minor issue (fix when possible)
   - **LOW** - Style/suggestion (consider fixing)

4. **Specific Recommendations**
   - File:line locations for each issue
   - Concrete fix suggestions
   - Code examples where applicable

## Agent Delegation

```
delegate(
  role="code-reviewer",
  tier="THOROUGH",
  prompt="CODE REVIEW TASK

Review code changes for quality, security, and maintainability.

Scope: [git diff or specific files]

Review Checklist:
- Security vulnerabilities (OWASP Top 10)
- Code quality (complexity, duplication)
- Performance issues (N+1, inefficient algorithms)
- Best practices (naming, documentation, error handling)
- Maintainability (coupling, testability)

Output: Code review report with:
- Files reviewed count
- Issues by severity (CRITICAL, HIGH, MEDIUM, LOW)
- Specific file:line locations
- Fix recommendations
- Approval recommendation (APPROVE / REQUEST CHANGES / COMMENT)"
)
```

## External Model Consultation (Preferred)

The code-reviewer agent SHOULD consult Codex for cross-validation.

### Protocol
1. **Form your OWN review FIRST** - Complete the review independently
2. **Consult for validation** - Cross-check findings with Codex
3. **Critically evaluate** - Never blindly adopt external findings
4. **Graceful fallback** - Never block if tools unavailable

### When to Consult
- Security-sensitive code changes
- Complex architectural patterns
- Unfamiliar codebases or languages
- High-stakes production code

### When to Skip
- Simple refactoring
- Well-understood patterns
- Time-critical reviews
- Small, isolated changes

### Tool Usage
Before first MCP tool use, call `ToolSearch("mcp")` to discover deferred MCP tools.
Use `mcp__x__ask_codex` with `agent_role: "code-reviewer"`.
If ToolSearch finds no MCP tools, fall back to the `code-reviewer` agent.

**Note:** Codex calls can take up to 1 hour. Consider the review timeline before consulting.

## Output Format

```
CODE REVIEW REPORT
==================

Files Reviewed: 8
Total Issues: 15

CRITICAL (0)
-----------
(none)

HIGH (3)
--------
1. src/api/auth.ts:42
   Issue: User input not sanitized before SQL query
   Risk: SQL injection vulnerability
   Fix: Use parameterized queries or ORM

2. src/components/UserProfile.tsx:89
   Issue: Password displayed in plain text in logs
   Risk: Credential exposure
   Fix: Remove password from log statements

3. src/utils/validation.ts:15
   Issue: Email regex allows invalid formats
   Risk: Accepts malformed emails
   Fix: Use proven email validation library

MEDIUM (7)
----------
...

LOW (5)
-------
...

RECOMMENDATION: REQUEST CHANGES

Critical security issues must be addressed before merge.
```

## Review Checklist

The code-reviewer agent checks:

### Security
- [ ] No hardcoded secrets (API keys, passwords, tokens)
- [ ] All user inputs sanitized
- [ ] SQL/NoSQL injection prevention
- [ ] XSS prevention (escaped outputs)
- [ ] CSRF protection on state-changing operations
- [ ] Authentication/authorization properly enforced

### Code Quality
- [ ] Functions < 50 lines (guideline)
- [ ] Cyclomatic complexity < 10
- [ ] No deeply nested code (> 4 levels)
- [ ] No duplicate logic (DRY principle)
- [ ] Clear, descriptive naming

### Performance
- [ ] No N+1 query patterns
- [ ] Appropriate caching where applicable
- [ ] Efficient algorithms (avoid O(n²) when O(n) possible)
- [ ] No unnecessary re-renders (React/Vue)

### Best Practices
- [ ] Error handling present and appropriate
- [ ] Logging at appropriate levels
- [ ] Documentation for public APIs
- [ ] Tests for critical paths
- [ ] No commented-out code

## Approval Criteria

**APPROVE** - No CRITICAL or HIGH issues, minor improvements only
**REQUEST CHANGES** - CRITICAL or HIGH issues present
**COMMENT** - Only LOW/MEDIUM issues, no blocking concerns


## Scenario Examples

**Good:** The user says `continue` after the workflow already has a clear next step. Continue the current branch of work instead of restarting or re-asking the same question.

**Good:** The user changes only the output shape or downstream delivery step (for example `make a PR`). Preserve earlier non-conflicting workflow constraints and apply the update locally.

**Bad:** The user says `continue`, and the workflow restarts discovery or stops before the missing verification/evidence is gathered.

## Use with Other Skills

**With Team:**
```
/team "review recent auth changes and report findings"
```
Includes coordinated review execution across specialized agents.

**With Ralph:**
```
/ralph code-review then fix all issues
```
Review code, get feedback, fix until approved.

**With Ultrawork:**
```
/ultrawork review all files in src/
```
Parallel code review across multiple files.

## Best Practices

- **Review early** - Catch issues before they compound
/home/jang/Workspace/warvis-findEvil
docs/find-evil/warvis-go-architecture.md

hook: PostToolUse
hook: PostToolUse Completed
codex
Using the `code-review` skill because this is a requested architecture review. I’ve confirmed the target doc path and am reading it with line references so the findings can be anchored precisely.
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba docs/find-evil/warvis-go-architecture.md | sed -n '1,260p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
     1	# W.A.R.V.I.S Go Bridge Architecture
     2	
     3	**Evil Has Nowhere to Hide**
     4	
     5	**Version**: 1.0.0-draft  
     6	**Author**: WoopsFactory  
     7	**Status**: Design Phase (Implementation pending)  
     8	**Date**: 2026-05-01
     9	
    10	---
    11	
    12	## 1. Overview
    13	
    14	W.A.R.V.I.S (Woops, A Rather Very Intelligent System) is an agentic incident response orchestrator designed for forensic malware detection. This document specifies the **Go bridge layer** that connects:
    15	
    16	1. **Gemma 4 26-A4B LLM** (via Ollama) — autonomous reasoning and tool selection
    17	2. **Hunt State Machine** (FSM) — orchestrated evidence investigation workflow
    18	3. **MCP Server** (Python) — forensic evidence processing via 9 whitelisted tools
    19	4. **Terminal Display** — real-time hunt progress visualization
    20	
    21	The Go binary (`warvis`) acts as an intelligent bridge that choreographs these components into a cohesive incident response system where **the LLM autonomously hunts within defined state boundaries**, and **the FSM enforces investigation discipline**.
    22	
    23	```
    24	┌─────────────────────────────────────────────────────────────────┐
    25	│                     W.A.R.V.I.S CLI (Go)                        │
    26	│                   Evil Has Nowhere to Hide                      │
    27	└─────────────────────────────────────────────────────────────────┘
    28	                               │
    29	                ┌──────────────┼──────────────┐
    30	                ▼              ▼              ▼
    31	        ┌─────────────┐ ┌─────────────┐ ┌──────────────┐
    32	        │ Hunt Engine │ │ Gemma Agent │ │ MCP Client   │
    33	        │ (FSM)       │ │ (Ollama)    │ │ (stdio)      │
    34	        └─────────────┘ └─────────────┘ └──────────────┘
    35	                │              │              │
    36	        ┌───────┴──────────────┼──────────────┴───────┐
    37	        │                      │                      │
    38	        ▼                      ▼                      ▼
    39	    State Registry      Tool Calling Loop      JSON-RPC 2.0
    40	    (INITIALIZE          (Function Schema       Protocol Handler
    41	     TRACE               Parsing)               (stdin/stdout)
    42	     SCAN
    43	     EXPOSE
    44	     LOCK)
    45	        │
    46	        ▼
    47	    ┌──────────────────────────────────────────┐
    48	    │    MCP Server: find-evil-mcp (Python)    │
    49	    │  9 Tools: case.open, timeline.build,     │
    50	    │  iocs.scan, memory.*, net.*, log.*,      │
    51	    │  report.append, verify.cross_check       │
    52	    └──────────────────────────────────────────┘
    53	        │
    54	        ▼
    55	    ┌──────────────────────────────────────────┐
    56	    │   SIFT Toolchain (Docker Subprocess)     │
    57	    │   Volatility3, YARA, Sigma, Zeek,        │
    58	    │   Suricata, plaso, log2timeline          │
    59	    └──────────────────────────────────────────┘
    60	```
    61	
    62	---
    63	
    64	## 2. Hunt Protocol — State Machine
    65	
    66	The **Hunt State Machine (HSM)** is the backbone of W.A.R.V.I.S. It enforces a disciplined investigation workflow where each state has:
    67	- **Assigned MCP tools** that the LLM can invoke
    68	- **Explicit success criteria** (evidence gathered, patterns detected)
    69	- **Automatic transition rules** based on tool outcomes
    70	- **Failure policies** (retry, escalate, lock down)
    71	
    72	### 2.1 State Definitions
    73	
    74	#### State: INITIALIZE
    75	**Purpose**: Register evidence source and validate case workspace.
    76	
    77	**Assigned Tools**:
    78	- `case.open` — Register evidence, create sandbox
    79	
    80	**LLM Autonomy**: None. This state is driven by CLI input; Gemma acts passively.
    81	
    82	**Success Criteria**:
    83	- `case_id` generated and valid (UUID v4)
    84	- Sandbox root created at `/cases/<case_id>/`
    85	- Evidence source registered (disk image / memory dump / PCAP / logs)
    86	
    87	**Failure Handling**:
    88	- Evidence unreachable → Error, ask user to verify path
    89	- Sandbox creation fails → Error, check disk space
    90	- **Max retries**: 1 (if evidence path is wrong, user must fix)
    91	
    92	**Transition Rules**:
    93	```
    94	INITIALIZE → TRACE (on success)
    95	INITIALIZE → LOCK  (on max retries exceeded)
    96	```
    97	
    98	**State Data**:
    99	```go
   100	type InitializeState struct {
   101	    CaseID       string   // UUID v4
   102	    SandboxRoot  string   // /cases/<case_id>/
   103	    EvidenceKind string   // disk_image|memory_dump|pcap|log_directory
   104	    EvidencePath string   // /evidence/...
   105	}
   106	```
   107	
   108	---
   109	
   110	#### State: TRACE
   111	**Purpose**: Build forensic timeline and identify candidate events.
   112	
   113	**Assigned Tools**:
   114	- `timeline.build` — Merge and de-duplicate events across sources
   115	- `log.query` — Search timeline for anomalies
   116	
   117	**LLM Autonomy**: HIGH. Gemma decides:
   118	- Which timeline sources to prioritize (disk/memory/logs/network)
   119	- Which log patterns to query (process creation, network connect, file access)
   120	- How many events to fetch per query
   121	- When evidence is sufficient to move to SCAN
   122	
   123	**Success Criteria** (any one):
   124	- ≥10 timeline events parsed and deduplicated
   125	- ≥1 suspicious log pattern detected (process names, IPs, file paths)
   126	- Timeline covers >60% of evidence time span
   127	
   128	**Failure Handling**:
   129	- No tools available (SIFT not running) → Warn, retry after 5s (1 retry max)
   130	- Timeout (>120s per call) → Abort state, transition to LOCK with warning
   131	- Empty timeline → Warn, Gemma can retry with different source_filter
   132	
   133	**Transition Rules**:
   134	```
   135	TRACE → SCAN        (on success criteria met + Gemma confidence ≥0.6)
   136	TRACE → EXPOSE      (if Gemma detects high-confidence IOCs in queries)
   137	TRACE → LOCK        (on max timeouts or empty evidence)
   138	```
   139	
   140	**State Data**:
   141	```go
   142	type TraceState struct {
   143	    CaseID         string        // from INITIALIZE
   144	    TimelinePath   string        // /cases/<case_id>/timeline.jsonl
   145	    EventCount     int           // de-duplicated events
   146	    SourcesFound   []string      // disk|memory|logs|network
   147	    LastQuery      string        // human-readable last query
   148	    QueryResults   []interface{} // recent log hits
   149	}
   150	```
   151	
   152	---
   153	
   154	#### State: SCAN
   155	**Purpose**: Run indicator-of-compromise detection and memory analysis.
   156	
   157	**Assigned Tools**:
   158	- `iocs.scan` — Match YARA/Sigma rules
   159	- `memory.process_list` — Extract processes
   160	- `memory.malfind` — Detect memory injection
   161	- `net.flow_summary` — Summarize network flows
   162	
   163	**LLM Autonomy**: VERY HIGH. Gemma orchestrates:
   164	- Which rulesets to scan (yara_default / sigma)
   165	- Which memory modules to analyze (all processes vs. suspicious PIDs)
   166	- Which network flows to rank (top_n, src/dst filtering)
   167	- Which findings are worth escalating to EXPOSE
   168	
   169	**Success Criteria** (all required):
   170	- IOC scan completed (≥1 rule evaluated)
   171	- Memory analysis attempted (if memory dump exists)
   172	- Network flows summarized (if PCAP exists)
   173	- ≥1 finding detected OR clean bill of health documented
   174	
   175	**Failure Handling**:
   176	- IOC scan fails → Warn, continue with memory analysis
   177	- Memory dump missing → Skip memory tools, log reason
   178	- Timeout → Escalate current findings to EXPOSE, don't block
   179	- Rule evaluation fails → Return empty matches, no error
   180	
   181	**Transition Rules**:
   182	```
   183	SCAN → EXPOSE       (on ≥1 finding detected)
   184	SCAN → EXPOSE       (on Gemma requests verification)
   185	SCAN → LOCK         (on max tool failures)
   186	```
   187	
   188	**State Data**:
   189	```go
   190	type ScanState struct {
   191	    CaseID       string
   192	    ScanID       string        // UUID v4, per IOC scan
   193	    Matches      []Finding     // IOC matches
   194	    Processes    []Process     // from memory.process_list
   195	    MalfindsVAD  []VADRegion   // from memory.malfind
   196	    NetFlows     []NetworkFlow // from net.flow_summary
   197	    ToolsRun     []string      // which tools succeeded
   198	}
   199	
   200	type Finding struct {
   201	    Rule       string
   202	    Path       string
   203	    Offset     int
   204	    Severity   string // info|low|medium|high|critical
   205	    Context    string
   206	}
   207	```
   208	
   209	---
   210	
   211	#### State: EXPOSE
   212	**Purpose**: Cross-verify findings and build confidence scores.
   213	
   214	**Assigned Tools**:
   215	- `verify.cross_check` — Re-verify findings via alternative methods
   216	- `report.append` — Append finding to case report
   217	
   218	**LLM Autonomy**: MEDIUM. Gemma:
   219	- Selects which findings to verify (prioritize high-severity)
   220	- Chooses verification method (rerun / alt_tool / counter_evidence / all)
   221	- Writes narrative for each finding (why this matters, chain of evidence)
   222	- Decides when findings are ready for lock-down
   223	
   224	**Success Criteria** (all required):
   225	- All high-severity findings verified
   226	- Confidence scores ≥0.5 for each finding
   227	- Report contains ≥1 finding entry
   228	- MITRE ATT&CK mappings documented (optional but encouraged)
   229	
   230	**Failure Handling**:
   231	- Verification timeout → Record finding as "provisional" (confidence 0.5)
   232	- Confidence <0.5 → Demote severity, flag for manual review
   233	- Report append fails → Error, retry once
   234	- No findings to verify → Mark case as clean, transition to LOCK
   235	
   236	**Transition Rules**:
   237	```
   238	EXPOSE → LOCK       (on findings verified OR case clean)
   239	EXPOSE → SCAN      (if Gemma wants additional scans)
   240	```
   241	
   242	**State Data**:
   243	```go
   244	type ExposeState struct {
   245	    CaseID      string
   246	    Findings    []VerifiedFinding // with confidence, verdict
   247	    ReportPath  string             // /cases/<case_id>/report/findings.jsonl
   248	    FindingHash string             // SHA256 of report for chain-of-custody
   249	}
   250	
   251	type VerifiedFinding struct {
   252	    FindingID      string
   253	    Title          string
   254	    Severity       string
   255	    Narrative      string
   256	    Confidence     float64 // 0.0 - 1.0
   257	    Verdict        string  // confirmed|unconfirmed|false_positive
   258	    Evidence       []string
   259	    MITREAttack    []string
   260	    VerifiedAt     string  // ISO 8601

hook: PostToolUse
hook: PostToolUse Completed
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba docs/find-evil/warvis-go-architecture.md | sed -n '261,620p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
   261	}
   262	```
   263	
   264	---
   265	
   266	#### State: LOCK
   267	**Purpose**: Finalize case, generate report, clean up investigation state.
   268	
   269	**Assigned Tools**: None (read-only final state)
   270	
   271	**LLM Autonomy**: None. HSM drives this state.
   272	
   273	**Success Criteria** (deterministic):
   274	- Report written to `/cases/<case_id>/report/`
   275	- All intermediate state cleared or archived
   276	- Case marked as CLOSED in registry
   277	
   278	**Side Effects**:
   279	- Generate final HTML/JSON report
   280	- Clean up temp files in sandbox
   281	- Log case summary to audit trail
   282	
   283	**Transition Rules**:
   284	```
   285	LOCK → (END)        (terminal state)
   286	```
   287	
   288	**State Data**:
   289	```go
   290	type LockState struct {
   291	    CaseID       string
   292	    Status       string  // CLOSED|ERROR|INCOMPLETE
   293	    ReportPath   string
   294	    Summary      string
   295	    ClosedAt     string  // ISO 8601
   296	}
   297	```
   298	
   299	---
   300	
   301	### 2.2 State Transition Matrix
   302	
   303	```
   304	┌──────────────┬──────────────┬──────────────┬──────────────┬──────────────┬──────────┐
   305	│ Current      │ INITIALIZE   │ TRACE        │ SCAN         │ EXPOSE       │ LOCK     │
   306	│ Next         │              │              │              │              │          │
   307	├──────────────┼──────────────┼──────────────┼──────────────┼──────────────┼──────────┤
   308	│ INITIALIZE   │ (repeat)     │ (invalid)    │ (invalid)    │ (invalid)    │ ERROR    │
   309	│ TRACE        │ (invalid)    │ (repeat)     │ (invalid)    │ (invalid)    │ TIMEOUT  │
   310	│ SCAN         │ (invalid)    │ RECURSIVE    │ (repeat)     │ (invalid)    │ TIMEOUT  │
   311	│ EXPOSE       │ (invalid)    │ (invalid)    │ RECURSIVE    │ (repeat)     │ (always) │
   312	│ LOCK         │ (invalid)    │ (invalid)    │ (invalid)    │ (invalid)    │ (repeat) │
   313	└──────────────┴──────────────┴──────────────┴──────────────┴──────────────┴──────────┘
   314	
   315	RECURSIVE transitions:
   316	- TRACE → SCAN: allowed when Gemma gathers timeline, then triggers IOC scans
   317	- SCAN → EXPOSE: allowed when findings detected, verification requested
   318	- SCAN → TRACE: allowed if Gemma wants deeper timeline query before continuing scan
   319	```
   320	
   321	---
   322	
   323	### 2.3 Automatic Retry & Timeout Policy
   324	
   325	| Trigger | Retry Condition | Max Retries | Backoff | Escalation |
   326	|---------|-----------------|-------------|---------|------------|
   327	| Tool timeout | Tool call exceeds 120s | 1 | 5s exponential | Escalate findings to EXPOSE, mark as provisional |
   328	| Tool error | subprocess fails (e.g., SIFT unavailable) | 2 | 10s linear | WARN, continue with next tool |
   329	| Network error | MCP connection drops | 3 | 5s exponential | Reconnect, preserve state |
   330	| Gemma error | LLM returns invalid tool call (wrong args) | 1 | immediate | Log error, Gemma reattempts |
   331	| Evidence missing | File not found after case.open | 0 | N/A | LOCK with ERROR |
   332	
   333	**Timeout Handling**:
   334	- If tool call exceeds 120s: interrupt subprocess, return timeout error to Gemma
   335	- If entire TRACE state exceeds 10 min: warn Gemma, offer option to move to SCAN
   336	- If entire hunt exceeds 30 min: force transition to LOCK (with warning to user)
   337	
   338	---
   339	
   340	## 3. Go Module Structure
   341	
   342	The Go codebase is organized into layered packages, each with clear responsibility and dependency direction.
   343	
   344	```
   345	warvis/
   346	├── cmd/warvis/
   347	│   ├── main.go              # CLI entry point, flag parsing, dispatch to hunt engine
   348	│   └── cli_flags.go         # --phase, --model, --case-id, --evidence-path definitions
   349	│
   350	├── internal/
   351	│   ├── hunt/                # FSM + state management
   352	│   │   ├── fsm.go           # Hunt State Machine (INITIALIZE→TRACE→SCAN→EXPOSE→LOCK)
   353	│   │   ├── state.go         # State interface + concrete types
   354	│   │   ├── registry.go      # Case registry (JSON-backed state storage)
   355	│   │   └── transitions.go   # State transition logic + validation
   356	│   │
   357	│   ├── agent/               # Gemma 4 tool-calling loop
   358	│   │   ├── client.go        # Ollama HTTP client
   359	│   │   ├── prompt.go        # System prompt construction (Hunt Protocol context)
   360	│   │   ├── tool_call.go     # Function schema parsing + validation
   361	│   │   └── loop.go          # Main agent loop: LLM → tool → response feedback
   362	│   │
   363	│   ├── mcp/                 # MCP stdio client
   364	│   │   ├── client.go        # JSON-RPC 2.0 message handler
   365	│   │   ├── protocol.go      # Request/response types
   366	│   │   ├── transport.go     # stdin/stdout reader/writer
   367	│   │   └── tool_registry.go # Cached list_tools response + schema cache
   368	│   │
   369	│   ├── display/             # Terminal UI (progress, state, tool output)
   370	│   │   ├── spinner.go       # State progress spinner
   371	│   │   ├── formatter.go     # Structured output formatting (JSON/table/markdown)
   372	│   │   └── logger.go        # Audit trail logger
   373	│   │
   374	│   └── config/              # Configuration management
   375	│       ├── config.go        # Config struct + defaults
   376	│       └── env.go           # Environment variable loading
   377	│
   378	├── pkg/ollama/
   379	│   ├── client.go            # HTTP client for Ollama API
   380	│   ├── chat.go              # /api/chat endpoint
   381	│   └── types.go             # Request/response types
   382	│
   383	└── go.mod
   384	```
   385	
   386	### 3.1 Package Responsibility Matrix
   387	
   388	| Package | Responsibility | Exports | Dependencies |
   389	|---------|-----------------|---------|--------------|
   390	| `cmd/warvis` | CLI parsing, entry point, main loop dispatch | `main()` | `hunt`, `agent`, `display`, `config` |
   391	| `hunt/fsm` | State machine implementation, transitions | `FSM`, `State`, `Registry` | `hunt/state`, `hunt/transitions` |
   392	| `hunt/state` | State type definitions (INITIALIZE, TRACE, SCAN, EXPOSE, LOCK) | `State`, `StateData` | none |
   393	| `agent/client` | Gemma 4 HTTP API wrapper (Ollama) | `Client`, `ChatRequest`, `ChatResponse` | `pkg/ollama` |
   394	| `agent/prompt` | Hunt protocol system prompt builder | `BuildSystemPrompt()` | `hunt/fsm` |
   395	| `agent/loop` | Main agentic loop (LLM → tools → feedback) | `Loop`, `ToolCallResult` | `agent/client`, `mcp/client`, `hunt/fsm` |
   396	| `mcp/client` | MCP protocol handler (JSON-RPC 2.0) | `Client`, `CallToolRequest` | `mcp/transport`, `mcp/protocol` |
   397	| `mcp/transport` | MCP stdio communication | `Transport` | none |
   398	| `display/` | Terminal output formatting | `Spinner`, `Format*` | none |
   399	| `config/` | Configuration & environment | `Config`, `Load()` | none |
   400	| `pkg/ollama` | Ollama HTTP API (reusable utility) | `Client`, `ChatMessage` | none |
   401	
   402	### 3.2 Dependency Direction
   403	
   404	```
   405	cmd/warvis
   406	  ├→ hunt/fsm, hunt/state, hunt/transitions
   407	  ├→ agent/loop, agent/client, agent/prompt
   408	  ├→ mcp/client
   409	  ├→ display/
   410	  └→ config/
   411	
   412	hunt/fsm
   413	  ├→ hunt/state
   414	  ├→ hunt/transitions
   415	  └→ display/  (for logging)
   416	
   417	agent/loop
   418	  ├→ agent/client, agent/prompt
   419	  ├→ mcp/client (call tools)
   420	  ├→ hunt/fsm   (transition states)
   421	  └→ display/   (log actions)
   422	
   423	agent/client
   424	  └→ pkg/ollama
   425	
   426	mcp/client
   427	  ├→ mcp/transport
   428	  └→ mcp/protocol
   429	
   430	All packages ←→ config/ (read-only)
   431	```
   432	
   433	**Constraint**: Circular dependencies forbidden. All dependencies point toward **config** (no cycle).
   434	
   435	---
   436	
   437	## 4. Gemma 4 Integration — Tool Calling Strategy
   438	
   439	### 4.1 Gemma 4 Capabilities Assessment
   440	
   441	**Key Question**: Does Gemma 4 26-A4B support native function calling?
   442	
   443	**Assumption** (to be validated before implementation):
   444	- Gemma 4 does **NOT** have native function-calling format support like Claude or GPT-4
   445	- We will use **JSON parsing strategy**: inject tool schemas as JSON in system prompt, Gemma outputs structured JSON for tool calls
   446	
   447	**If native function calling is supported** (discovered during implementation):
   448	- Use native format (if Ollama exposes it)
   449	- Fall back to JSON parsing for compatibility
   450	- This design accommodates both paths
   451	
   452	---
   453	
   454	### 4.2 Tool Schema Injection
   455	
   456	The **Hunt Protocol system prompt** includes:
   457	
   458	1. **Operational context**: Current case_id, evidence type, which state in FSM
   459	2. **Tool schemas**: JSON array of available tools + signatures (for current state)
   460	3. **Tool constraints**: Which tools are valid for current state
   461	4. **Success criteria**: What constitutes "done" for this state
   462	5. **Examples**: Few-shot examples of well-formed tool calls
   463	
   464	**System Prompt Structure**:
   465	
   466	```
   467	# W.A.R.V.I.S Hunt Protocol — {{STATE}} Phase
   468	
   469	You are hunting for malware and evil in forensic evidence.  
   470	Evil has nowhere to hide.
   471	
   472	## Current Investigation
   473	- Case ID: {{CASE_ID}}
   474	- Evidence Type: {{EVIDENCE_KIND}} (disk_image|memory_dump|pcap|log_directory)
   475	- Hunt Phase: {{STATE}}
   476	- Sandbox: /cases/{{CASE_ID}}/
   477	
   478	## Tools Available in {{STATE}} Phase
   479	
   480	{{AVAILABLE_TOOLS_JSON}}
   481	
   482	Each tool is a JSON object with:
   483	- "name": tool identifier
   484	- "description": what it does
   485	- "inputSchema": JSON Schema for arguments
   486	- "examples": sample calls
   487	
   488	## Output Format
   489	
   490	When you need to call a tool, respond ONLY with valid JSON:
   491	
   492	{
   493	  "action": "call_tool",
   494	  "tool_name": "...",
   495	  "arguments": { ... }
   496	}
   497	
   498	When you're done with this phase, respond with:
   499	
   500	{
   501	  "action": "state_complete",
   502	  "reason": "..."
   503	}
   504	
   505	When you need human input or can't proceed:
   506	
   507	{
   508	  "action": "escalate",
   509	  "reason": "..."
   510	}
   511	
   512	## {{STATE}} Success Criteria
   513	
   514	{{SUCCESS_CRITERIA}}
   515	
   516	## Hunt Discipline
   517	
   518	- You CANNOT use tools from other phases
   519	- You CANNOT write to /evidence/ or /home/
   520	- You CANNOT run arbitrary shells or code
   521	- Malware often hides in: process memory, injected code, network exfiltration, log tampering
   522	- Look for: suspicious processes, unsigned binaries, network anomalies, timeline gaps
   523	
   524	Begin hunting. Evil has nowhere to hide.
   525	```
   526	
   527	---
   528	
   529	### 4.3 Tool Call Loop
   530	
   531	```go
   532	type AgentLoop struct {
   533	    ollama     *ollama.Client
   534	    mcp        *mcp.Client
   535	    fsm        *hunt.FSM
   536	    systemPrompt string
   537	    conversationHistory []Message
   538	}
   539	
   540	func (loop *AgentLoop) Run(ctx context.Context) error {
   541	    for {
   542	        // Step 1: Build current system prompt (state-aware)
   543	        systemPrompt := agent.BuildSystemPrompt(loop.fsm.CurrentState())
   544	
   545	        // Step 2: Call Gemma 4
   546	        response, err := loop.ollama.Chat(ctx, &OllamaRequest{
   547	            Model:    "gemma4:26b-a4b-instruct-q4_K_M",
   548	            System:   systemPrompt,
   549	            Messages: loop.conversationHistory,
   550	            Stream:   false,
   551	        })
   552	        if err != nil {
   553	            return fmt.Errorf("gemma4 call failed: %w", err)
   554	        }
   555	
   556	        // Step 3: Parse response for action
   557	        action, err := ParseAction(response.Message.Content)
   558	        if err != nil {
   559	            // Gemma returned invalid JSON; escalate
   560	            loop.conversationHistory = append(loop.conversationHistory,
   561	                Message{Role: "assistant", Content: response.Message.Content})
   562	            loop.conversationHistory = append(loop.conversationHistory,
   563	                Message{Role: "user", Content: "Invalid JSON. Respond with one of: call_tool | state_complete | escalate"})
   564	            continue
   565	        }
   566	
   567	        // Step 4: Dispatch action
   568	        switch action.Type {
   569	        case "call_tool":
   570	            // Validate tool is allowed in current state
   571	            toolName := action.ToolName
   572	            if !loop.fsm.IsToolAllowed(toolName) {
   573	                loop.conversationHistory = append(loop.conversationHistory,
   574	                    Message{Role: "user", Content: fmt.Sprintf("Tool %s not allowed in %s state", toolName, loop.fsm.CurrentState())})
   575	                continue
   576	            }
   577	
   578	            // Call MCP tool
   579	            toolResult, err := loop.mcp.CallTool(ctx, toolName, action.Arguments)
   580	            if err != nil {
   581	                // Tool error; feedback to Gemma
   582	                loop.conversationHistory = append(loop.conversationHistory,
   583	                    Message{Role: "assistant", Content: response.Message.Content})
   584	                loop.conversationHistory = append(loop.conversationHistory,
   585	                    Message{Role: "user", Content: fmt.Sprintf("Tool %s failed: %v", toolName, err)})
   586	                continue
   587	            }
   588	
   589	            // Success; add result to history
   590	            loop.conversationHistory = append(loop.conversationHistory,
   591	                Message{Role: "assistant", Content: response.Message.Content})
   592	            loop.conversationHistory = append(loop.conversationHistory,
   593	                Message{Role: "user", Content: fmt.Sprintf("Tool %s result:\n%s", toolName, toolResult)})
   594	
   595	        case "state_complete":
   596	            // Attempt state transition
   597	            err := loop.fsm.Transition(action.Reason)
   598	            if err != nil {
   599	                // Transition failed; ask Gemma to continue current state
   600	                loop.conversationHistory = append(loop.conversationHistory,
   601	                    Message{Role: "user", Content: fmt.Sprintf("Cannot transition: %v. Continue hunting.", err)})
   602	                continue
   603	            }
   604	            return nil // Move to next state
   605	
   606	        case "escalate":
   607	            // User intervention required; pause FSM
   608	            loop.fsm.Pause(action.Reason)
   609	            return errors.New("hunt paused: " + action.Reason)
   610	        }
   611	    }
   612	}
   613	```
   614	
   615	---
   616	
   617	### 4.4 Conversation History Management
   618	
   619	To avoid token bloat:
   620	

hook: PostToolUse
hook: PostToolUse Completed
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba docs/find-evil/warvis-go-architecture.md | sed -n '617,1040p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
   617	### 4.4 Conversation History Management
   618	
   619	To avoid token bloat:
   620	
   621	1. **Sliding window**: Keep last 20 exchanges (request/response pairs)
   622	2. **Summarization**: If >30 exchanges, summarize older tool results into 1 "historical context" message
   623	3. **Tool output truncation**: Limit each tool response to 2KB in conversation history (full result logged separately)
   624	
   625	```go
   626	func (loop *AgentLoop) TrimHistory() {
   627	    if len(loop.conversationHistory) > 40 {
   628	        // Keep last 20 exchanges (40 messages)
   629	        loop.conversationHistory = loop.conversationHistory[len(loop.conversationHistory)-40:]
   630	    }
   631	}
   632	```
   633	
   634	---
   635	
   636	## 5. MCP Client Design (Go ↔ Python)
   637	
   638	The Go `mcp` package implements a **JSON-RPC 2.0 client** for the Python MCP server.
   639	
   640	### 5.1 Protocol Flow
   641	
   642	```
   643	Go (warvis)                          Python (find-evil-mcp)
   644	│                                            │
   645	├─ initialize() ──────────────────────────→ │
   646	│  {jsonrpc: "2.0",                         │
   647	│   method: "initialize",                   │
   648	│   params: {clientInfo: {...}},            │
   649	│   id: 1}                                  │
   650	│                                    │ parse JSON-RPC
   651	│                        ┌───────────┴──────────────┐
   652	│                        │ Create server instance   │
   653	│                        │ Load tools (case.open,   │
   654	│                        │ timeline.build, ...)     │
   655	│                        └──────────────────────────┘
   656	│ ← initialize response ──────────────────────────┤
   657	│  {result: {                                      │
   658	│    capabilities: {...}}}                        │
   659	│                                            │
   660	├─ list_tools() ─────────────────────────────────→ │
   661	│  {jsonrpc: "2.0",                                │
   662	│   method: "tools/list",                         │
   663	│   id: 2}                                        │
   664	│                                    │ inspect server.tools[]
   665	│ ← list_tools response ─────────────────────────┤
   666	│  {result: {                                     │
   667	│    tools: [{name, inputSchema, ...}, ...]}}    │
   668	│                                            │
   669	├─ call_tool("case.open") ──────────────────────→ │
   670	│  {jsonrpc: "2.0",                              │
   671	│   method: "tools/call",                        │
   672	│   params: {                                    │
   673	│    name: "case.open",                          │
   674	│    arguments: {image_path: "/evidence/..."}},  │
   675	│   id: 3}                                       │
   676	│                        ┌──────────────────────────┐
   677	│                        │ Validate args against    │
   678	│                        │ inputSchema              │
   679	│                        │ Call tool handler        │
   680	│                        │ Validate output against  │
   681	│                        │ outputSchema             │
   682	│                        └──────────────────────────┘
   683	│ ← call_tool response ──────────────────────────┤
   684	│  {result: {                                     │
   685	│    content: [{type: "text",                    │
   686	│      text: "{...json...}"}]}}                  │
   687	```
   688	
   689	---
   690	
   691	### 5.2 Go MCP Client Interface
   692	
   693	```go
   694	package mcp
   695	
   696	// Client is a JSON-RPC 2.0 MCP client.
   697	type Client struct {
   698	    transport Transport
   699	    idCounter int64
   700	    pending   map[int64]chan Response
   701	}
   702	
   703	// Initialize handshake.
   704	func (c *Client) Initialize(ctx context.Context) (ServerInfo, error)
   705	
   706	// List available tools.
   707	func (c *Client) ListTools(ctx context.Context) ([]Tool, error)
   708	
   709	// Call a tool.
   710	func (c *Client) CallTool(ctx context.Context, name string, args map[string]interface{}) (string, error)
   711	
   712	// Response represents a JSON-RPC response.
   713	type Response struct {
   714	    JSONRPC string      `json:"jsonrpc"`
   715	    Result  interface{} `json:"result"`
   716	    Error   *RPCError   `json:"error"`
   717	    ID      int64       `json:"id"`
   718	}
   719	
   720	// RPCError represents a JSON-RPC error.
   721	type RPCError struct {
   722	    Code    int         `json:"code"`
   723	    Message string      `json:"message"`
   724	    Data    interface{} `json:"data"`
   725	}
   726	```
   727	
   728	---
   729	
   730	### 5.3 Transport Layer (stdio)
   731	
   732	```go
   733	package mcp
   734	
   735	// Transport handles JSON-RPC message exchange over stdin/stdout.
   736	type Transport struct {
   737	    cmd    *exec.Cmd
   738	    stdin  io.WriteCloser
   739	    stdout *bufio.Scanner
   740	    mu     sync.Mutex
   741	}
   742	
   743	// New starts the MCP server subprocess.
   744	func New(ctx context.Context, mcpCmd string) (*Transport, error) {
   745	    cmd := exec.CommandContext(ctx, "python", "-m", "find_evil_mcp.server")
   746	    // or: mcpCmd = "find-evil-mcp" if installed via pip
   747	    
   748	    stdin, err := cmd.StdinPipe()
   749	    if err != nil {
   750	        return nil, err
   751	    }
   752	    
   753	    stdout, err := cmd.StdoutPipe()
   754	    if err != nil {
   755	        return nil, err
   756	    }
   757	    
   758	    if err := cmd.Start(); err != nil {
   759	        return nil, err
   760	    }
   761	    
   762	    return &Transport{
   763	        cmd:    cmd,
   764	        stdin:  stdin,
   765	        stdout: bufio.NewScanner(stdout),
   766	    }, nil
   767	}
   768	
   769	// Send writes a JSON-RPC request.
   770	func (t *Transport) Send(req interface{}) error {
   771	    t.mu.Lock()
   772	    defer t.mu.Unlock()
   773	    
   774	    data, err := json.Marshal(req)
   775	    if err != nil {
   776	        return err
   777	    }
   778	    
   779	    _, err = t.stdin.Write(append(data, '\n'))
   780	    return err
   781	}
   782	
   783	// Recv reads a JSON-RPC response.
   784	func (t *Transport) Recv(ctx context.Context) ([]byte, error) {
   785	    done := make(chan []byte, 1)
   786	    go func() {
   787	        if t.stdout.Scan() {
   788	            done <- t.stdout.Bytes()
   789	        } else {
   790	            done <- nil
   791	        }
   792	    }()
   793	    
   794	    select {
   795	    case data := <-done:
   796	        if data == nil {
   797	            return nil, io.EOF
   798	        }
   799	        return data, nil
   800	    case <-ctx.Done():
   801	        return nil, ctx.Err()
   802	    }
   803	}
   804	
   805	// Close terminates the server subprocess.
   806	func (t *Transport) Close() error {
   807	    t.stdin.Close()
   808	    return t.cmd.Wait()
   809	}
   810	```
   811	
   812	---
   813	
   814	### 5.4 Error Handling & Timeouts
   815	
   816	| Error Case | Handling | Retry |
   817	|------------|----------|-------|
   818	| Server process dies | Error, do NOT restart (user must fix) | No |
   819	| JSON parse error | Return parse error, log raw output | No |
   820	| Tool call timeout (>120s) | Interrupt, return "TIMEOUT" error | 1x (exponential backoff 5s) |
   821	| Tool not found | Return "not_found" error | No |
   822	| Tool args invalid | Return validation error (from MCP server) | No (let Gemma retry with fixed args) |
   823	| Server unresponsive (no response for 30s) | Timeout error | 1x reconnect |
   824	
   825	---
   826	
   827	## 6. CLI Design
   828	
   829	### 6.1 Command Syntax
   830	
   831	```bash
   832	# Primary: hunt evidence
   833	warvis hunt <evidence_path> [options]
   834	  --phase <state>            # Start at specific state (default: INITIALIZE)
   835	  --model <model>            # Gemma model name (default: gemma4:26b-a4b-instruct-q4_K_M)
   836	  --case-id <uuid>           # Existing case ID (default: generate new)
   837	  --output <format>          # json|markdown|html (default: json)
   838	  --timeout <seconds>        # Overall hunt timeout (default: 1800s = 30min)
   839	
   840	# Status query
   841	warvis status <case_id> [--output <format>]
   842	
   843	# Report retrieval
   844	warvis report <case_id> [--output json|markdown|html]
   845	
   846	# Case listing
   847	warvis list [--state <state>] [--output <format>]
   848	
   849	# Configuration
   850	warvis config [show|set <key> <value>]
   851	
   852	# Debug
   853	warvis debug <case_id>       # Show detailed state, conversation history, logs
   854	```
   855	
   856	### 6.2 Examples
   857	
   858	```bash
   859	# Analyze disk image, hunt from INITIALIZE (default)
   860	$ warvis hunt /evidence/disk.raw
   861	
   862	# Resume existing case at SCAN phase
   863	$ warvis hunt /evidence/disk.raw --case-id a1b2c3d4-... --phase SCAN
   864	
   865	# Generate markdown report
   866	$ warvis report a1b2c3d4-... --output markdown > report.md
   867	
   868	# Check case status
   869	$ warvis status a1b2c3d4-...
   870	
   871	# Debug: show conversation history and state
   872	$ warvis debug a1b2c3d4-...
   873	```
   874	
   875	### 6.3 Flag Validation
   876	
   877	```go
   878	type Flags struct {
   879	    EvidencePath string        // required (must exist)
   880	    Phase        string        // optional (default: INITIALIZE)
   881	    ModelName    string        // optional (default: gemma4:26b-a4b-instruct-q4_K_M)
   882	    CaseID       string        // optional (UUID format if provided)
   883	    Output       string        // optional (json|markdown|html)
   884	    Timeout      time.Duration // optional (default: 30min, max: 2h)
   885	}
   886	
   887	func (f *Flags) Validate() error {
   888	    // EvidencePath must exist and be readable
   889	    if _, err := os.Stat(f.EvidencePath); err != nil {
   890	        return fmt.Errorf("evidence path not found: %v", err)
   891	    }
   892	    
   893	    // Phase must be valid FSM state
   894	    if f.Phase != "" && !isValidState(f.Phase) {
   895	        return fmt.Errorf("invalid phase: %s", f.Phase)
   896	    }
   897	    
   898	    // CaseID must be valid UUID if provided
   899	    if f.CaseID != "" && !isValidUUID(f.CaseID) {
   900	        return fmt.Errorf("invalid case ID: %s", f.CaseID)
   901	    }
   902	    
   903	    // Timeout must be reasonable
   904	    if f.Timeout > 2*time.Hour {
   905	        return fmt.Errorf("timeout too large: max 2h")
   906	    }
   907	    
   908	    return nil
   909	}
   910	```
   911	
   912	---
   913	
   914	## 7. Security Boundaries (Go Layer)
   915	
   916	The Go bridge adds additional validation on top of the Python MCP server's security model.
   917	
   918	### 7.1 Input Validation
   919	
   920	| Input | Constraint | Enforcement |
   921	|-------|-----------|------------|
   922	| `case_id` | UUID v4 format | Regex validation before any use |
   923	| `evidence_path` | Must be under `/evidence/` or relative | Path normalization + prefix check |
   924	| `phase` | One of 5 FSM states | Enum validation |
   925	| `model_name` | Alphanumeric + colon (Ollama format) | Regex: `^[a-zA-Z0-9_-]+:[a-zA-Z0-9._/-]+$` |
   926	| Ollama URL | Valid HTTP URL | URL parsing + scheme check (http/https only) |
   927	| Tool arguments (from Gemma) | Must pass MCP input schema | Delegated to MCP server (Go validates before forward) |
   928	
   929	**Implementation**:
   930	
   931	```go
   932	func ValidateCaseID(id string) error {
   933	    if !isValidUUID(id) {
   934	        return fmt.Errorf("invalid case ID format: %s", id)
   935	    }
   936	    return nil
   937	}
   938	
   939	func NormalizePath(p string) (string, error) {
   940	    abs, err := filepath.Abs(p)
   941	    if err != nil {
   942	        return "", err
   943	    }
   944	    
   945	    // Prevent directory traversal
   946	    if strings.Contains(abs, "..") {
   947	        return "", fmt.Errorf("path contains ..: %s", p)
   948	    }
   949	    
   950	    return abs, nil
   951	}
   952	```
   953	
   954	---
   955	
   956	### 7.2 Process Isolation
   957	
   958	| Boundary | Mechanism |
   959	|----------|-----------|
   960	| Go binary ↔ Ollama | HTTP only (localhost:11434 default), no direct code execution |
   961	| Go binary ↔ MCP server | stdio only (no shell), JSON-RPC 2.0 messages only |
   962	| Go binary ↔ Filesystem | Case sandbox root enforced at initialization |
   963	| Gemma ↔ Tools | Tool whitelist enforced per FSM state in Go (before forward to MCP) |
   964	
   965	---
   966	
   967	### 7.3 Audit Logging
   968	
   969	All significant actions logged to `/cases/<case_id>/audit.jsonl`:
   970	
   971	```json
   972	{"timestamp": "2026-05-01T12:34:56Z", "event": "case_opened", "case_id": "...", "evidence_path": "..."}
   973	{"timestamp": "2026-05-01T12:35:00Z", "event": "state_transition", "from": "INITIALIZE", "to": "TRACE"}
   974	{"timestamp": "2026-05-01T12:35:05Z", "event": "tool_called", "tool": "timeline.build", "args": {...}}
   975	{"timestamp": "2026-05-01T12:35:30Z", "event": "tool_result", "tool": "timeline.build", "result_summary": "..."}
   976	{"timestamp": "2026-05-01T12:36:00Z", "event": "gemma_response", "action": "call_tool", "tool": "iocs.scan"}
   977	```
   978	
   979	---
   980	
   981	## 8. Open Questions & Implementation Notes
   982	
   983	### 8.1 Gemma 4 Function Calling
   984	
   985	**Q**: Does Gemma 4 support native function-calling format (like Claude's tools API)?
   986	
   987	**A** (to be determined):
   988	- If YES: Use native format, inject tool schemas into Ollama API if supported
   989	- If NO: Use JSON-in-prompt strategy (current design assumption)
   990	- **Implementation path**: Design supports both; validate during POC
   991	
   992	**Action item**: Test with Ollama before finalizing implementation.
   993	
   994	---
   995	
   996	### 8.2 Conversation History Strategy
   997	
   998	**Q**: How do we prevent token bloat with multi-tool hunting sessions?
   999	
  1000	**A** (proposed):
  1001	- Sliding window: keep last 20 exchanges (request/response pairs)
  1002	- Truncate tool outputs to 2KB in conversation (full results logged separately)
  1003	- Summarize if >30 exchanges (create "historical context" message)
  1004	
  1005	**Action item**: Validate with real hunt scenarios; adjust window size based on Gemma 4 context window.
  1006	
  1007	---
  1008	
  1009	### 8.3 Ollama Model Configuration
  1010	
  1011	**Q**: Where does the Gemma 4 model live? How to ensure consistent setup?
  1012	
  1013	**A** (proposed):
  1014	- Assume Ollama running locally at `localhost:11434` (default)
  1015	- Model name: `gemma4:26b-a4b-instruct-q4_K_M` (quantized 4-bit, ~20GB)
  1016	- **Alternative (lighter)**: `gemma4:7b-instruct` (~5GB) for prototyping
  1017	- Configuration via CLI flag `--model` or env var `WARVIS_OLLAMA_MODEL`
  1018	
  1019	**Action item**: Document Ollama setup in main README; provide Docker Compose for reproducibility.
  1020	
  1021	---
  1022	
  1023	### 8.4 State Persistence & Resume
  1024	
  1025	**Q**: Can we pause a hunt and resume later? What if Go/Gemma process crashes?
  1026	
  1027	**A** (proposed):
  1028	- State persisted to `/cases/<case_id>/state.json` after each FSM transition
  1029	- Conversation history persisted to `/cases/<case_id>/conversation.jsonl`
  1030	- On resume with `--case-id`, load prior state and history
  1031	- **Limitation**: If Gemma is mid-response, restart at state transition (no mid-state resume)
  1032	
  1033	**Action item**: Implement state serialization; test crash recovery scenario.
  1034	
  1035	---
  1036	
  1037	### 8.5 Tool Call Validation
  1038	
  1039	**Q**: How do we prevent Gemma from calling tools with invalid arguments?
  1040	

hook: PostToolUse
hook: PostToolUse Completed
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba docs/find-evil/warvis-go-architecture.md | sed -n '1037,1280p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
  1037	### 8.5 Tool Call Validation
  1038	
  1039	**Q**: How do we prevent Gemma from calling tools with invalid arguments?
  1040	
  1041	**A** (proposed):
  1042	- Go validates tool arguments against MCP inputSchema before forwarding
  1043	- If invalid: return validation error to Gemma in conversation history
  1044	- Gemma reattempts with corrected arguments
  1045	- Max 3 retries per tool call; then escalate to human
  1046	
  1047	**Action item**: Implement schema validation in Go; test with adversarial prompts.
  1048	
  1049	---
  1050	
  1051	### 8.6 Memory & Token Limits
  1052	
  1053	**Q**: Gemma 4 26B has ~8K token context. How do we fit case investigation?
  1054	
  1055	**A** (proposed):
  1056	- System prompt: ~1K tokens (Hunt Protocol + tool schemas)
  1057	- Conversation history (sliding window): ~3-4K tokens (20 exchanges)
  1058	- Gemma response: ~2-3K tokens (tool calls + reasoning)
  1059	- **Reserve**: ~1K for unforeseen overhead
  1060	- **Strategy**: If conversation grows beyond 4K, summarize older results
  1061	
  1062	**Action item**: Measure token usage empirically; adjust window size as needed.
  1063	
  1064	---
  1065	
  1066	### 8.7 Failure Modes & Escalation
  1067	
  1068	**Q**: What happens if:
  1069	1. Ollama crashes mid-hunt?
  1070	2. MCP server becomes unresponsive?
  1071	3. Gemma enters infinite loop of invalid tool calls?
  1072	
  1073	**A** (proposed):
  1074	1. **Ollama crash**: Pause hunt, return "Gemma unavailable" error (requires manual restart)
  1075	2. **MCP unresponsive**: After 30s timeout, escalate findings to EXPOSE and lock down
  1076	3. **Invalid tool loops**: After 3 consecutive invalid calls, escalate to human review
  1077	
  1078	**Action item**: Implement timeout and escalation handlers; test each scenario.
  1079	
  1080	---
  1081	
  1082	### 8.8 Output Formats
  1083	
  1084	**Q**: Which output formats to support initially?
  1085	
  1086	**A** (proposed, by priority):
  1087	1. **JSON**: Machine-readable, tool-friendly (Phase 1)
  1088	2. **Markdown**: Human-readable report for Devpost (Phase 1)
  1089	3. **HTML**: Interactive report with timeline visualization (Phase 2, nice-to-have)
  1090	
  1091	**Action item**: Implement JSON + Markdown formatters in `display/` package.
  1092	
  1093	---
  1094	
  1095	## 9. Mermaid Diagram: Full Hunt Flow
  1096	
  1097	```mermaid
  1098	graph TD
  1099	    A["warvis hunt <evidence>"] --> B["Initialize FSM"]
  1100	    B --> C["INITIALIZE state"]
  1101	    C --> D["case.open tool"]
  1102	    D --> E{"Success?"}
  1103	    E -->|Yes| F["TRACE state"]
  1104	    E -->|No| Z["LOCK (error)"]
  1105	    
  1106	    F --> G["Build system prompt<br/>with TRACE tools"]
  1107	    G --> H["Call Gemma 4 (Ollama)"]
  1108	    H --> I["Gemma returns action JSON"]
  1109	    I --> J{"Action type?"}
  1110	    
  1111	    J -->|call_tool| K["Validate tool allowed<br/>in TRACE"]
  1112	    K --> L{"Valid?"}
  1113	    L -->|Yes| M["Forward to MCP"]
  1114	    L -->|No| N["Error to Gemma<br/>restart loop"]
  1115	    N --> H
  1116	    M --> O["MCP calls tool<br/>subprocess"]
  1117	    O --> P["Return result<br/>to Gemma"]
  1118	    P --> H
  1119	    
  1120	    J -->|state_complete| Q["Check success<br/>criteria"]
  1121	    Q --> R{"Criteria met?"}
  1122	    R -->|Yes| S["SCAN state"]
  1123	    R -->|No| N
  1124	    
  1125	    J -->|escalate| T["Pause hunt"]
  1126	    T --> Z
  1127	    
  1128	    S --> U["SCAN phase:<br/>iocs.scan,<br/>memory.*,<br/>net.*"]
  1129	    U --> V["Gemma autonomously<br/>selects tools"]
  1130	    V --> H
  1131	    
  1132	    U --> W{"Findings<br/>detected?"}
  1133	    W -->|Yes| X["EXPOSE state"]
  1134	    W -->|No| Y["LOCK state"]
  1135	    
  1136	    X --> AA["verify.cross_check<br/>report.append"]
  1137	    AA --> AB["LOCK state"]
  1138	    AB --> Z["Generate report<br/>& cleanup"]
  1139	```
  1140	
  1141	---
  1142	
  1143	## 10. References & Appendix
  1144	
  1145	### 10.1 Related Documents
  1146	
  1147	- `docs/find-evil/architecture.md` — Python MCP server architecture
  1148	- `docs/find-evil/dataset.md` — Forensic evidence dataset specifications
  1149	- `plans/ITEM-212-find-evil/spec.md` — Full project specification
  1150	- `harness/find-evil/docker-compose.yml` — SIFT harness definition
  1151	- `src/find_evil_mcp/server.py` — MCP server implementation
  1152	
  1153	### 10.2 External References
  1154	
  1155	- **JSON-RPC 2.0**: https://www.jsonrpc.org/specification
  1156	- **MCP Spec**: https://modelcontextprotocol.io/
  1157	- **Ollama API**: https://github.com/ollama/ollama/blob/main/docs/api.md
  1158	- **Gemma 4 26B**: https://ai.google.dev/gemma (model card)
  1159	- **Go stdlib**: https://golang.org/pkg/
  1160	
  1161	### 10.3 Key Design Principles
  1162	
  1163	1. **Modularity**: Each Go package has single responsibility and clear interfaces
  1164	2. **Security First**: Multi-layer validation (Go → MCP → SIFT), no shell execution
  1165	3. **Transparency**: All state changes logged; conversation history preserved
  1166	4. **Robustness**: Timeouts, retries, graceful degradation on tool failures
  1167	5. **Extensibility**: Easy to add new FSM states, tools, LLM models
  1168	6. **User Control**: Human can pause/resume, review findings before LOCK
  1169	
  1170	### 10.4 Glossary
  1171	
  1172	- **Hunt**: Single forensic investigation (case_id scope)
  1173	- **FSM**: Finite State Machine (INITIALIZE → TRACE → SCAN → EXPOSE → LOCK)
  1174	- **State**: Current phase in hunt (e.g., TRACE)
  1175	- **Tool**: MCP capability (e.g., case.open, timeline.build)
  1176	- **Finding**: IOC match or malware indicator detected
  1177	- **Confidence**: Verification score (0.0 - 1.0) for a finding
  1178	- **Case Registry**: Persistent JSON record of all cases
  1179	- **Audit Trail**: Chronological log of all hunt actions
  1180	
  1181	---
  1182	
  1183	**W.A.R.V.I.S Architecture v1.0**  
  1184	*Evil Has Nowhere to Hide*
  1185	

hook: PostToolUse
hook: PostToolUse Completed
codex
The main document has enough material for a review, but I’m checking the referenced Python MCP/server surfaces briefly to catch contract mismatches rather than reviewing the Go design in isolation.
hook: PreToolUse
hook: PreToolUse
hook: PreToolUse
hook: PreToolUse Completed
hook: PreToolUse Completed
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba docs/find-evil/architecture.md | sed -n '1,220p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
     1	<!-- SPDX-License-Identifier: MIT -->
     2	
     3	# FIND EVIL MCP Server — Architecture & Security Boundaries
     4	
     5	**Version**: 0.1.0 (Phase 1-2 transition)  
     6	**Status**: Phase 2 (9 tools: 3 Phase 1 + 6 Phase 2 implemented)  
     7	**Phase 2 Status**: Schema complete; implementation pending (implementer phase 2)
     8	
     9	## 1. System Architecture
    10	
    11	```mermaid
    12	flowchart TD
    13	    subgraph Orchestrator["Claude Code (Incident Response Orchestrator)"]
    14	        Agent["Agent: autonomous IR loop\n+ self-correction"]
    15	        Loop["Retry logic\n+ find-evil-verifier"]
    16	    end
    17	
    18	    subgraph MCP["MCP Server (find-evil-mcp)"]
    19	        direction TB
    20	        CaseOpen["case.open\n(evidence sandbox)"]
    21	        TimelineBuild["timeline.build\n(normalized events)"]
    22	        IOCsScan["iocs.scan\n(IOC matches)"]
    23	        Whitelist["✓ Whitelisted\nsubprocesses only\n✗ No shell/exec"]
    24	    end
    25	
    26	    subgraph SIFT["SIFT Toolchain (Docker proxy)"]
    27	        direction TB
    28	        Tools["Volatility3\nplaso / log2timeline\nYARA\nSigma\nZeek\nSuricata\n(others: OOS)"]
    29	    end
    30	
    31	    subgraph Evidence["Evidence (SIFT host /evidence/)"]
    32	        direction TB
    33	        Disk["disk images"]
    34	        Memory["memory dumps"]
    35	        Logs["logs"]
    36	        PCAP["network captures"]
    37	    end
    38	
    39	    subgraph Sandbox["Case Sandbox (/cases/<case_id>/)"]
    40	        direction TB
    41	        Registry["case metadata"]
    42	        Timeline["timeline.jsonl"]
    43	        Report["report/"]
    44	    end
    45	
    46	    Orchestrator -->|stdio + HTTP| MCP
    47	    MCP -->|whitelisted subprocess| SIFT
    48	    SIFT -->|ingest from| Evidence
    49	    MCP -->|read| Evidence
    50	    MCP -->|write| Sandbox
    51	    Agent -->|calls via MCP| CaseOpen
    52	    Agent -->|calls via MCP| TimelineBuild
    53	    Agent -->|calls via MCP| IOCsScan
    54	    Loop -->|cross-check| IOCsScan
    55	```
    56	
    57	## 2. Security Boundary Matrix
    58	
    59	| Component | Allow | Deny | Evidence |
    60	|-----------|-------|------|----------|
    61	| **MCP input** | case_id (uuid pattern), path (regex `/evidence/` or `/cases/<id>`), enum params | shell, eval, arbitrary code, `/etc`, `/home`, outside sandbox | JSON Schema with `pattern`, `enum`, `oneOf` constraints |
    62	| **MCP output** | structured JSON only, sandbox-relative paths | raw stdout, binary blobs, absolute paths outside `/cases/` | outputSchema enforces JSON, no `$ref` to external schemas |
    63	| **File read** | `/evidence/*` (SIFT host perspective), `/cases/<case_id>/*` (sandbox) | `/etc`, `/root`, `/tmp` outside sandbox, network egress | case.open enforces `/evidence/` prefix; tool sandbox has read-only mount on evidence |
    64	| **File write** | `/cases/<case_id>/report/` (structured JSON only) | anywhere else; direct write to evidence | report.append (Phase 2) restricted to report/ subdir |
    65	| **Subprocess** | plaso, log2timeline, yara, sigma, volatility3, zeek, suricata (whitelist only) | bash, sh, python -c, perl, ruby -e, nc, curl | wrapper scripts call tools with fixed args + case_id, no user shell pass-through |
    66	| **Network** | none (except SIFT Docker proxy localhost) | external HTTP/DNS/SSH | No socket syscalls in sandbox; proxy-only via Docker bridge |
    67	
    68	## 3. Capability Handshake (Server Info)
    69	
    70	When the orchestrator connects, the MCP server responds with this capability manifest:
    71	
    72	```json
    73	{
    74	  "server_info": {
    75	    "name": "find-evil-mcp",
    76	    "version": "0.1.0",
    77	    "description": "Forensic evidence processing orchestration via SIFT toolchain"
    78	  },
    79	  "capabilities": {
    80	    "tools": [
    81	      {
    82	        "name": "case.open",
    83	        "version": "0.1.0",
    84	        "description": "Register and sandbox forensic evidence source",
    85	        "schema_ref": "harness/find-evil/mcp-schema/case.open.json",
    86	        "phase": "1"
    87	      },
    88	      {
    89	        "name": "timeline.build",
    90	        "version": "0.1.0",
    91	        "description": "Construct normalized forensic timeline from case evidence",
    92	        "schema_ref": "harness/find-evil/mcp-schema/timeline.build.json",
    93	        "phase": "1"
    94	      },
    95	      {
    96	        "name": "iocs.scan",
    97	        "version": "0.1.0",
    98	        "description": "Scan case evidence against YARA/Sigma IOC rules",
    99	        "schema_ref": "harness/find-evil/mcp-schema/iocs.scan.json",
   100	        "phase": "1"
   101	      },
   102	      {
   103	        "name": "memory.process_list",
   104	        "version": "0.1.0",
   105	        "description": "Extract process list from memory dump (Volatility3)",
   106	        "schema_ref": "harness/find-evil/mcp-schema/memory.process_list.json",
   107	        "phase": "2"
   108	      },
   109	      {
   110	        "name": "memory.malfind",
   111	        "version": "0.1.0",
   112	        "description": "Detect injected/suspicious memory regions (Volatility3)",
   113	        "schema_ref": "harness/find-evil/mcp-schema/memory.malfind.json",
   114	        "phase": "2"
   115	      },
   116	      {
   117	        "name": "net.flow_summary",
   118	        "version": "0.1.0",
   119	        "description": "Extract and summarize network flows from PCAP (Zeek/Suricata)",
   120	        "schema_ref": "harness/find-evil/mcp-schema/net.flow_summary.json",
   121	        "phase": "2"
   122	      },
   123	      {
   124	        "name": "log.query",
   125	        "version": "0.1.0",
   126	        "description": "Query forensic timeline and logs with structured filters",
   127	        "schema_ref": "harness/find-evil/mcp-schema/log.query.json",
   128	        "phase": "2"
   129	      },
   130	      {
   131	        "name": "report.append",
   132	        "version": "0.1.0",
   133	        "description": "Append structured finding to case report with hash chain integrity",
   134	        "schema_ref": "harness/find-evil/mcp-schema/report.append.json",
   135	        "phase": "2"
   136	      },
   137	      {
   138	        "name": "verify.cross_check",
   139	        "version": "0.1.0",
   140	        "description": "Re-verify findings against multiple sources for confidence scoring",
   141	        "schema_ref": "harness/find-evil/mcp-schema/verify.cross_check.json",
   142	        "phase": "2"
   143	      }
   144	    ],
   145	    "constraints": {
   146	      "sandbox": "/cases/<case_id>/",
   147	      "evidence_host_path": "/evidence/",
   148	      "no_shell": true,
   149	      "no_eval": true,
   150	      "json_only_output": true,
   151	      "subprocess_whitelist": ["plaso", "log2timeline", "yara", "sigma", "volatility3", "zeek", "suricata"]
   152	    }
   153	  }
   154	}
   155	```
   156	
   157	## 4. Input/Output Validation Flow
   158	
   159	```mermaid
   160	flowchart LR
   161	    A["Agent calls\nMCP tool"]
   162	    B["JSON Schema\ninput validation\ncase.open.json"]
   163	    C["Enforce path\nprefix constraint\n/evidence/ or /cases/"]
   164	    D["Whitelist subprocess\nargs + case_id"]
   165	    E["Tool executes\nwithin sandbox"]
   166	    F["outputSchema\nvalidation"]
   167	    G["Return JSON\nto agent"]
   168	
   169	    A --> B
   170	    B -->|pass| C
   171	    C -->|pass| D
   172	    D --> E
   173	    E --> F
   174	    F -->|pass| G
   175	    B -->|fail| H["Error: invalid input\njson response"]
   176	    C -->|fail| I["Error: path\noutside boundary"]
   177	    H --> G
   178	    I --> G
   179	```
   180	
   181	## 5. Phase 1 Tool Specifications
   182	
   183	### 5.1 case.open (evidence sandbox)
   184	
   185	**Purpose**: Register a forensic evidence source and create an isolated case workspace.
   186	
   187	**Input constraints**:
   188	- Exactly one of: `image_path`, `mem_path`, `pcap_path`, `logdir_path`
   189	- All paths match regex `^/evidence/[a-zA-Z0-9._/-]+$` (SIFT host absolute)
   190	- No symlink traversal or `..` sequences allowed
   191	
   192	**Output**:
   193	- `case_id`: UUID v4 (immutable case namespace)
   194	- `sandbox_root`: `/cases/<case_id>/` (case working directory)
   195	- `accepted_inputs`: array of {kind, path} for audit trail
   196	
   197	**Side effects**:
   198	- Creates `/cases/<case_id>/` directory tree (owned by orchestrator)
   199	- Creates `/cases/<case_id>/report/` subdirectory for findings
   200	- Registers case metadata in local registry (JSON file)
   201	
   202	**Error handling**:
   203	- If evidence source unreachable: return 404-equivalent JSON error
   204	- If sandbox creation fails: return 500-equivalent JSON error
   205	- All errors are structured JSON; no exception stack traces
   206	
   207	### 5.2 timeline.build (normalized forensic timeline)
   208	
   209	**Purpose**: De-duplicate and merge filesystem, memory, and log events into a canonical timeline.
   210	
   211	**Input constraints**:
   212	- `case_id`: must be valid UUID v4 from prior `case.open` call
   213	- `source_filter`: enum {disk, memory, logs, network, all} (default: all)
   214	- `since`, `until`: ISO 8601 timestamps (optional; filter events by time range)
   215	
   216	**Output**:
   217	- `timeline_path`: `/cases/<case_id>/timeline.jsonl` (JSON-Lines, one event per line)
   218	- `event_count`: total events after de-duplication
   219	- `tool_used`: enum {plaso, log2timeline}
   220	- `generated_at`: ISO 8601 completion timestamp

hook: PostToolUse
exec
/bin/bash -lc "nl -ba src/find_evil_mcp/server.py | sed -n '1,260p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
     1	# SPDX-License-Identifier: MIT
     2	
     3	import sys
     4	import asyncio
     5	from mcp.server import Server
     6	from mcp.types import Tool, TextContent
     7	
     8	from find_evil_mcp.tools.case_open import handle_case_open
     9	from find_evil_mcp.tools.timeline_build import handle_timeline_build
    10	from find_evil_mcp.tools.iocs_scan import handle_iocs_scan
    11	from find_evil_mcp.tools.memory_process_list import handle_memory_process_list
    12	from find_evil_mcp.tools.memory_malfind import handle_memory_malfind
    13	from find_evil_mcp.tools.net_flow_summary import handle_net_flow_summary
    14	from find_evil_mcp.tools.log_query import handle_log_query
    15	from find_evil_mcp.tools.report_append import handle_report_append
    16	from find_evil_mcp.tools.verify_cross_check import handle_verify_cross_check
    17	from find_evil_mcp.schema_loader import get_output_schema
    18	
    19	
    20	def create_server() -> Server:
    21	    """Create and configure the MCP server."""
    22	    server = Server("find-evil-mcp")
    23	
    24	    @server.list_tools()
    25	    async def list_tools():
    26	        """List available tools."""
    27	        return [
    28	            Tool(
    29	                name="case.open",
    30	                description="Register and sandbox forensic evidence source",
    31	                inputSchema={
    32	                    "type": "object",
    33	                    "oneOf": [
    34	                        {
    35	                            "type": "object",
    36	                            "properties": {
    37	                                "image_path": {
    38	                                    "type": "string",
    39	                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
    40	                                }
    41	                            },
    42	                            "required": ["image_path"],
    43	                            "additionalProperties": False,
    44	                        },
    45	                        {
    46	                            "type": "object",
    47	                            "properties": {
    48	                                "mem_path": {
    49	                                    "type": "string",
    50	                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
    51	                                }
    52	                            },
    53	                            "required": ["mem_path"],
    54	                            "additionalProperties": False,
    55	                        },
    56	                        {
    57	                            "type": "object",
    58	                            "properties": {
    59	                                "pcap_path": {
    60	                                    "type": "string",
    61	                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
    62	                                }
    63	                            },
    64	                            "required": ["pcap_path"],
    65	                            "additionalProperties": False,
    66	                        },
    67	                        {
    68	                            "type": "object",
    69	                            "properties": {
    70	                                "logdir_path": {
    71	                                    "type": "string",
    72	                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
    73	                                }
    74	                            },
    75	                            "required": ["logdir_path"],
    76	                            "additionalProperties": False,
    77	                        },
    78	                    ],
    79	                },
    80	                outputSchema=get_output_schema("case.open"),
    81	            ),
    82	            Tool(
    83	                name="timeline.build",
    84	                description="Construct normalized forensic timeline from case evidence",
    85	                inputSchema={
    86	                    "type": "object",
    87	                    "properties": {
    88	                        "case_id": {
    89	                            "type": "string",
    90	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
    91	                        },
    92	                        "source_filter": {
    93	                            "type": "string",
    94	                            "enum": ["disk", "memory", "logs", "network", "all"],
    95	                            "default": "all",
    96	                        },
    97	                        "since": {"type": "string", "format": "date-time"},
    98	                        "until": {"type": "string", "format": "date-time"},
    99	                    },
   100	                    "required": ["case_id"],
   101	                    "additionalProperties": False,
   102	                },
   103	                outputSchema=get_output_schema("timeline.build"),
   104	            ),
   105	            Tool(
   106	                name="iocs.scan",
   107	                description="Scan case evidence against YARA/Sigma IOC rules",
   108	                inputSchema={
   109	                    "type": "object",
   110	                    "properties": {
   111	                        "case_id": {
   112	                            "type": "string",
   113	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   114	                        },
   115	                        "ruleset": {
   116	                            "type": "string",
   117	                            "enum": ["yara_default", "yara_custom", "sigma"],
   118	                        },
   119	                        "target_glob": {
   120	                            "type": "string",
   121	                            "pattern": "^[a-zA-Z0-9._/*?-]+$",
   122	                        },
   123	                    },
   124	                    "required": ["case_id", "ruleset"],
   125	                    "additionalProperties": False,
   126	                },
   127	                outputSchema=get_output_schema("iocs.scan"),
   128	            ),
   129	            Tool(
   130	                name="memory.process_list",
   131	                description="Extract process list from memory dump (Volatility3)",
   132	                inputSchema={
   133	                    "type": "object",
   134	                    "properties": {
   135	                        "case_id": {
   136	                            "type": "string",
   137	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   138	                        },
   139	                        "kdbg_offset": {
   140	                            "type": "string",
   141	                            "pattern": "^0x[0-9a-fA-F]+$",
   142	                        },
   143	                    },
   144	                    "required": ["case_id"],
   145	                    "additionalProperties": False,
   146	                },
   147	                outputSchema=get_output_schema("memory.process_list"),
   148	            ),
   149	            Tool(
   150	                name="memory.malfind",
   151	                description="Detect injected/suspicious memory regions (Volatility3)",
   152	                inputSchema={
   153	                    "type": "object",
   154	                    "properties": {
   155	                        "case_id": {
   156	                            "type": "string",
   157	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   158	                        },
   159	                        "pid": {
   160	                            "type": "integer",
   161	                            "minimum": 0,
   162	                        },
   163	                    },
   164	                    "required": ["case_id"],
   165	                    "additionalProperties": False,
   166	                },
   167	                outputSchema=get_output_schema("memory.malfind"),
   168	            ),
   169	            Tool(
   170	                name="net.flow_summary",
   171	                description="Extract and summarize network flows from PCAP (Zeek/Suricata)",
   172	                inputSchema={
   173	                    "type": "object",
   174	                    "properties": {
   175	                        "case_id": {
   176	                            "type": "string",
   177	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   178	                        },
   179	                        "pcap_id": {
   180	                            "type": "string",
   181	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   182	                        },
   183	                        "top_n": {
   184	                            "type": "integer",
   185	                            "minimum": 1,
   186	                            "maximum": 1000,
   187	                            "default": 50,
   188	                        },
   189	                    },
   190	                    "required": ["case_id", "pcap_id"],
   191	                    "additionalProperties": False,
   192	                },
   193	                outputSchema=get_output_schema("net.flow_summary"),
   194	            ),
   195	            Tool(
   196	                name="log.query",
   197	                description="Query forensic timeline and logs with structured filters",
   198	                inputSchema={
   199	                    "type": "object",
   200	                    "properties": {
   201	                        "case_id": {
   202	                            "type": "string",
   203	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   204	                        },
   205	                        "q": {
   206	                            "type": "string",
   207	                            "maxLength": 512,
   208	                            "pattern": "^[a-zA-Z0-9_.:= /*\"'-]+$",
   209	                        },
   210	                        "source": {
   211	                            "type": "string",
   212	                            "enum": ["evtx", "syslog", "audit", "all"],
   213	                            "default": "all",
   214	                        },
   215	                        "since": {"type": "string", "format": "date-time"},
   216	                        "until": {"type": "string", "format": "date-time"},
   217	                        "limit": {
   218	                            "type": "integer",
   219	                            "minimum": 1,
   220	                            "maximum": 10000,
   221	                            "default": 1000,
   222	                        },
   223	                    },
   224	                    "required": ["case_id", "q"],
   225	                    "additionalProperties": False,
   226	                },
   227	                outputSchema=get_output_schema("log.query"),
   228	            ),
   229	            Tool(
   230	                name="report.append",
   231	                description="Append structured finding to case report with hash chain integrity",
   232	                inputSchema={
   233	                    "type": "object",
   234	                    "properties": {
   235	                        "case_id": {
   236	                            "type": "string",
   237	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   238	                        },
   239	                        "finding": {
   240	                            "type": "object",
   241	                            "properties": {
   242	                                "title": {"type": "string"},
   243	                                "severity": {
   244	                                    "type": "string",
   245	                                    "enum": ["info", "low", "medium", "high", "critical"],
   246	                                },
   247	                                "narrative": {"type": "string"},
   248	                                "evidence": {"type": "array"},
   249	                                "mitre_attack": {"type": "array"},
   250	                            },
   251	                            "required": ["title", "severity", "narrative", "evidence"],
   252	                        },
   253	                    },
   254	                    "required": ["case_id", "finding"],
   255	                    "additionalProperties": False,
   256	                },
   257	                outputSchema=get_output_schema("report.append"),
   258	            ),
   259	            Tool(
   260	                name="verify.cross_check",

hook: PostToolUse
exec
/bin/bash -lc "rg -n \"case\\.open|timeline\\.build|iocs\\.scan|memory\\.|net\\.|log\\.|report\\.append|verify\\.cross_check|tools/list|initialize|stdio|FastMCP|MCP\" src docs/find-evil -S" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
docs/find-evil/architecture.md:3:# FIND EVIL MCP Server — Architecture & Security Boundaries
docs/find-evil/architecture.md:18:    subgraph MCP["MCP Server (find-evil-mcp)"]
docs/find-evil/architecture.md:20:        CaseOpen["case.open\n(evidence sandbox)"]
docs/find-evil/architecture.md:21:        TimelineBuild["timeline.build\n(normalized events)"]
docs/find-evil/architecture.md:22:        IOCsScan["iocs.scan\n(IOC matches)"]
docs/find-evil/architecture.md:46:    Orchestrator -->|stdio + HTTP| MCP
docs/find-evil/architecture.md:47:    MCP -->|whitelisted subprocess| SIFT
docs/find-evil/architecture.md:49:    MCP -->|read| Evidence
docs/find-evil/architecture.md:50:    MCP -->|write| Sandbox
docs/find-evil/architecture.md:51:    Agent -->|calls via MCP| CaseOpen
docs/find-evil/architecture.md:52:    Agent -->|calls via MCP| TimelineBuild
docs/find-evil/architecture.md:53:    Agent -->|calls via MCP| IOCsScan
docs/find-evil/architecture.md:61:| **MCP input** | case_id (uuid pattern), path (regex `/evidence/` or `/cases/<id>`), enum params | shell, eval, arbitrary code, `/etc`, `/home`, outside sandbox | JSON Schema with `pattern`, `enum`, `oneOf` constraints |
docs/find-evil/architecture.md:62:| **MCP output** | structured JSON only, sandbox-relative paths | raw stdout, binary blobs, absolute paths outside `/cases/` | outputSchema enforces JSON, no `$ref` to external schemas |
docs/find-evil/architecture.md:63:| **File read** | `/evidence/*` (SIFT host perspective), `/cases/<case_id>/*` (sandbox) | `/etc`, `/root`, `/tmp` outside sandbox, network egress | case.open enforces `/evidence/` prefix; tool sandbox has read-only mount on evidence |
docs/find-evil/architecture.md:64:| **File write** | `/cases/<case_id>/report/` (structured JSON only) | anywhere else; direct write to evidence | report.append (Phase 2) restricted to report/ subdir |
docs/find-evil/architecture.md:70:When the orchestrator connects, the MCP server responds with this capability manifest:
docs/find-evil/architecture.md:82:        "name": "case.open",
docs/find-evil/architecture.md:85:        "schema_ref": "harness/find-evil/mcp-schema/case.open.json",
docs/find-evil/architecture.md:89:        "name": "timeline.build",
docs/find-evil/architecture.md:92:        "schema_ref": "harness/find-evil/mcp-schema/timeline.build.json",
docs/find-evil/architecture.md:96:        "name": "iocs.scan",
docs/find-evil/architecture.md:99:        "schema_ref": "harness/find-evil/mcp-schema/iocs.scan.json",
docs/find-evil/architecture.md:103:        "name": "memory.process_list",
docs/find-evil/architecture.md:106:        "schema_ref": "harness/find-evil/mcp-schema/memory.process_list.json",
docs/find-evil/architecture.md:110:        "name": "memory.malfind",
docs/find-evil/architecture.md:113:        "schema_ref": "harness/find-evil/mcp-schema/memory.malfind.json",
docs/find-evil/architecture.md:117:        "name": "net.flow_summary",
docs/find-evil/architecture.md:120:        "schema_ref": "harness/find-evil/mcp-schema/net.flow_summary.json",
docs/find-evil/architecture.md:124:        "name": "log.query",
docs/find-evil/architecture.md:127:        "schema_ref": "harness/find-evil/mcp-schema/log.query.json",
docs/find-evil/architecture.md:131:        "name": "report.append",
docs/find-evil/architecture.md:134:        "schema_ref": "harness/find-evil/mcp-schema/report.append.json",
docs/find-evil/architecture.md:138:        "name": "verify.cross_check",
docs/find-evil/architecture.md:141:        "schema_ref": "harness/find-evil/mcp-schema/verify.cross_check.json",
docs/find-evil/architecture.md:161:    A["Agent calls\nMCP tool"]
docs/find-evil/architecture.md:162:    B["JSON Schema\ninput validation\ncase.open.json"]
docs/find-evil/architecture.md:183:### 5.1 case.open (evidence sandbox)
docs/find-evil/architecture.md:207:### 5.2 timeline.build (normalized forensic timeline)
docs/find-evil/architecture.md:212:- `case_id`: must be valid UUID v4 from prior `case.open` call
docs/find-evil/architecture.md:240:### 5.3 iocs.scan (indicator-of-compromise detection)
docs/find-evil/architecture.md:245:- `case_id`: valid UUID v4 from prior `case.open`
docs/find-evil/architecture.md:278:| `memory.process_list` | Extract process list from memory dump | case_id, kdbg_offset? | array of {pid, ppid, name, create_time, exit_time?, suspicious_flags[]} | volatility3 |
docs/find-evil/architecture.md:279:| `memory.malfind` | Detect injected/suspicious memory regions | case_id, pid? | array of {pid, vad_start, vad_end, protection, yara_hits[], severity} | volatility3 |
docs/find-evil/architecture.md:280:| `net.flow_summary` | Summarize network flows from PCAP | case_id, pcap_id, top_n? | array of {src_ip, dst_ip, src_port, dst_port, proto, bytes, packets, suspicious_flags[]} | zeek, suricata |
docs/find-evil/architecture.md:281:| `log.query` | Query timeline/logs with structured filters | case_id, q, source?, since?, until?, limit? | array of {timestamp, source, host?, user?, level?, message, raw_offset} | log2timeline, plaso-query, sigma |
docs/find-evil/architecture.md:282:| `report.append` | Append finding with hash chain integrity | case_id, finding{title, severity, narrative, evidence[], mitre_attack[]?} | {finding_id, appended_at, report_path, prev_hash, this_hash} | none (internal write only) |
docs/find-evil/architecture.md:283:| `verify.cross_check` | Re-verify finding via cross-check | case_id, finding_id, method? | {agreements[], disagreements[], confidence, verdict, tool_used[], generated_at} | none (internal read-only) |
docs/find-evil/architecture.md:289:- ❌ Shell or exec MCP tools (architectural constraint, not prompt guard)
docs/find-evil/architecture.md:308:- `case.open(image_path)` → case_id + sandbox_root
docs/find-evil/architecture.md:309:- `timeline.build(case_id)` → ≥10 events in timeline.jsonl
docs/find-evil/architecture.md:310:- `iocs.scan(case_id, ruleset="yara_default")` → ≥1 match from bundled YARA rules
docs/find-evil/architecture.md:314:- `memory.process_list(case_id)` → process array with timestamps and suspicion flags
docs/find-evil/architecture.md:315:- `memory.malfind(case_id)` → findings with VAD regions and YARA hits
docs/find-evil/architecture.md:316:- `net.flow_summary(case_id, pcap_id)` → flow array sorted by bytes, top_n limit enforced
docs/find-evil/architecture.md:317:- `log.query(case_id, q="malware")` → log hits with truncation flag, no shell injection
docs/find-evil/architecture.md:318:- `report.append(case_id, finding)` → finding_id + hash chain (prev_hash, this_hash)
docs/find-evil/architecture.md:319:- `verify.cross_check(case_id, finding_id)` → confidence score, verdict enum validation
docs/find-evil/dataset.md:34:#### 1. syslog.log (5.5 KB)
docs/find-evil/dataset.md:61:#### 4. memory.dmp (1.0 KB)
docs/find-evil/dataset.md:71:- **Purpose**: Unit testing for `memory.process_list()` and `memory.malfind()` without requiring real volatility images
docs/find-evil/dataset.md:109:### iocs.scan (2 matches)
docs/find-evil/dataset.md:113:| rootkit_marker | syslog.log | critical | 1 match |
docs/find-evil/dataset.md:115:### log.query (4 queries with hits)
docs/find-evil/dataset.md:118:| "Failed password" | syslog.log | ≥4 | medium | SSH brute-force |
docs/find-evil/dataset.md:119:| "rootkit.ko" | syslog.log | ≥1 | critical | Kernel module loading |
docs/find-evil/dataset.md:120:| "/bin/curl" | syslog.log | ≥1 | high | Suspicious download command |
docs/find-evil/dataset.md:121:| "malware.example" | syslog.log | ≥1 | high | External malware domain |
docs/find-evil/dataset.md:123:### memory.process_list (3 processes)
docs/find-evil/dataset.md:130:### memory.malfind (1 injected region)
docs/find-evil/dataset.md:135:### net.flow_summary (3 flows)
docs/find-evil/dataset.md:158:1. Load fixture via `case.open(mem_path="/evidence/case-001/evidence/memory.dmp")`
docs/find-evil/dataset.md:167:### MCP Tool Usage Example
docs/find-evil/dataset.md:171:case_resp = client.call_tool("case.open", {
docs/find-evil/dataset.md:172:    "mem_path": "/evidence/case-001/evidence/memory.dmp"
docs/find-evil/dataset.md:177:ioc_resp = client.call_tool("iocs.scan", {
docs/find-evil/dataset.md:184:expected_iocs = manifest["expected_findings"]["iocs.scan"]
docs/find-evil/dataset.md:245:| cases/case-001/evidence/syslog.log | 5.5 KB | 6a2630a3... | Log evidence |
docs/find-evil/dataset.md:248:| cases/case-001/evidence/memory.dmp | 1.0 KB | 4a8bbdb1... | Memory placeholder |
docs/find-evil/dataset.md:255:- [x] syslog.log: 50 lines, 6+ evil markers
docs/find-evil/dataset.md:257:- [x] memory.dmp: JSON placeholder with expected_pids/malfind_pids
docs/find-evil/dataset.md:268:- **MCP Schemas**: `harness/find-evil/mcp-schema/` (9 tools)
docs/find-evil/warvis-go-architecture.md:18:3. **MCP Server** (Python) — forensic evidence processing via 9 whitelisted tools
docs/find-evil/warvis-go-architecture.md:32:        │ Hunt Engine │ │ Gemma Agent │ │ MCP Client   │
docs/find-evil/warvis-go-architecture.md:33:        │ (FSM)       │ │ (Ollama)    │ │ (stdio)      │
docs/find-evil/warvis-go-architecture.md:48:    │    MCP Server: find-evil-mcp (Python)    │
docs/find-evil/warvis-go-architecture.md:49:    │  9 Tools: case.open, timeline.build,     │
docs/find-evil/warvis-go-architecture.md:50:    │  iocs.scan, memory.*, net.*, log.*,      │
docs/find-evil/warvis-go-architecture.md:51:    │  report.append, verify.cross_check       │
docs/find-evil/warvis-go-architecture.md:67:- **Assigned MCP tools** that the LLM can invoke
docs/find-evil/warvis-go-architecture.md:78:- `case.open` — Register evidence, create sandbox
docs/find-evil/warvis-go-architecture.md:114:- `timeline.build` — Merge and de-duplicate events across sources
docs/find-evil/warvis-go-architecture.md:115:- `log.query` — Search timeline for anomalies
docs/find-evil/warvis-go-architecture.md:158:- `iocs.scan` — Match YARA/Sigma rules
docs/find-evil/warvis-go-architecture.md:159:- `memory.process_list` — Extract processes
docs/find-evil/warvis-go-architecture.md:160:- `memory.malfind` — Detect memory injection
docs/find-evil/warvis-go-architecture.md:161:- `net.flow_summary` — Summarize network flows
docs/find-evil/warvis-go-architecture.md:194:    Processes    []Process     // from memory.process_list
docs/find-evil/warvis-go-architecture.md:195:    MalfindsVAD  []VADRegion   // from memory.malfind
docs/find-evil/warvis-go-architecture.md:196:    NetFlows     []NetworkFlow // from net.flow_summary
docs/find-evil/warvis-go-architecture.md:215:- `verify.cross_check` — Re-verify findings via alternative methods
docs/find-evil/warvis-go-architecture.md:216:- `report.append` — Append finding to case report
docs/find-evil/warvis-go-architecture.md:329:| Network error | MCP connection drops | 3 | 5s exponential | Reconnect, preserve state |
docs/find-evil/warvis-go-architecture.md:331:| Evidence missing | File not found after case.open | 0 | N/A | LOCK with ERROR |
docs/find-evil/warvis-go-architecture.md:363:│   ├── mcp/                 # MCP stdio client
docs/find-evil/warvis-go-architecture.md:396:| `mcp/client` | MCP protocol handler (JSON-RPC 2.0) | `Client`, `CallToolRequest` | `mcp/transport`, `mcp/protocol` |
docs/find-evil/warvis-go-architecture.md:397:| `mcp/transport` | MCP stdio communication | `Transport` | none |
docs/find-evil/warvis-go-architecture.md:578:            // Call MCP tool
docs/find-evil/warvis-go-architecture.md:636:## 5. MCP Client Design (Go ↔ Python)
docs/find-evil/warvis-go-architecture.md:638:The Go `mcp` package implements a **JSON-RPC 2.0 client** for the Python MCP server.
docs/find-evil/warvis-go-architecture.md:645:├─ initialize() ──────────────────────────→ │
docs/find-evil/warvis-go-architecture.md:647:│   method: "initialize",                   │
docs/find-evil/warvis-go-architecture.md:653:│                        │ Load tools (case.open,   │
docs/find-evil/warvis-go-architecture.md:654:│                        │ timeline.build, ...)     │
docs/find-evil/warvis-go-architecture.md:656:│ ← initialize response ──────────────────────────┤
docs/find-evil/warvis-go-architecture.md:662:│   method: "tools/list",                         │
docs/find-evil/warvis-go-architecture.md:669:├─ call_tool("case.open") ──────────────────────→ │
docs/find-evil/warvis-go-architecture.md:673:│    name: "case.open",                          │
docs/find-evil/warvis-go-architecture.md:691:### 5.2 Go MCP Client Interface
docs/find-evil/warvis-go-architecture.md:696:// Client is a JSON-RPC 2.0 MCP client.
docs/find-evil/warvis-go-architecture.md:730:### 5.3 Transport Layer (stdio)
docs/find-evil/warvis-go-architecture.md:743:// New starts the MCP server subprocess.
docs/find-evil/warvis-go-architecture.md:822:| Tool args invalid | Return validation error (from MCP server) | No (let Gemma retry with fixed args) |
docs/find-evil/warvis-go-architecture.md:916:The Go bridge adds additional validation on top of the Python MCP server's security model.
docs/find-evil/warvis-go-architecture.md:927:| Tool arguments (from Gemma) | Must pass MCP input schema | Delegated to MCP server (Go validates before forward) |
docs/find-evil/warvis-go-architecture.md:961:| Go binary ↔ MCP server | stdio only (no shell), JSON-RPC 2.0 messages only |
docs/find-evil/warvis-go-architecture.md:963:| Gemma ↔ Tools | Tool whitelist enforced per FSM state in Go (before forward to MCP) |
docs/find-evil/warvis-go-architecture.md:974:{"timestamp": "2026-05-01T12:35:05Z", "event": "tool_called", "tool": "timeline.build", "args": {...}}
docs/find-evil/warvis-go-architecture.md:975:{"timestamp": "2026-05-01T12:35:30Z", "event": "tool_result", "tool": "timeline.build", "result_summary": "..."}
docs/find-evil/warvis-go-architecture.md:976:{"timestamp": "2026-05-01T12:36:00Z", "event": "gemma_response", "action": "call_tool", "tool": "iocs.scan"}
docs/find-evil/warvis-go-architecture.md:1042:- Go validates tool arguments against MCP inputSchema before forwarding
docs/find-evil/warvis-go-architecture.md:1070:2. MCP server becomes unresponsive?
docs/find-evil/warvis-go-architecture.md:1075:2. **MCP unresponsive**: After 30s timeout, escalate findings to EXPOSE and lock down
docs/find-evil/warvis-go-architecture.md:1101:    C --> D["case.open tool"]
docs/find-evil/warvis-go-architecture.md:1113:    L -->|Yes| M["Forward to MCP"]
docs/find-evil/warvis-go-architecture.md:1116:    M --> O["MCP calls tool<br/>subprocess"]
docs/find-evil/warvis-go-architecture.md:1128:    S --> U["SCAN phase:<br/>iocs.scan,<br/>memory.*,<br/>net.*"]
docs/find-evil/warvis-go-architecture.md:1136:    X --> AA["verify.cross_check<br/>report.append"]
docs/find-evil/warvis-go-architecture.md:1147:- `docs/find-evil/architecture.md` — Python MCP server architecture
docs/find-evil/warvis-go-architecture.md:1151:- `src/find_evil_mcp/server.py` — MCP server implementation
docs/find-evil/warvis-go-architecture.md:1156:- **MCP Spec**: https://modelcontextprotocol.io/
docs/find-evil/warvis-go-architecture.md:1164:2. **Security First**: Multi-layer validation (Go → MCP → SIFT), no shell execution
docs/find-evil/warvis-go-architecture.md:1175:- **Tool**: MCP capability (e.g., case.open, timeline.build)
src/find_evil_mcp/server.py:21:    """Create and configure the MCP server."""
src/find_evil_mcp/server.py:29:                name="case.open",
src/find_evil_mcp/server.py:80:                outputSchema=get_output_schema("case.open"),
src/find_evil_mcp/server.py:83:                name="timeline.build",
src/find_evil_mcp/server.py:103:                outputSchema=get_output_schema("timeline.build"),
src/find_evil_mcp/server.py:106:                name="iocs.scan",
src/find_evil_mcp/server.py:127:                outputSchema=get_output_schema("iocs.scan"),
src/find_evil_mcp/server.py:130:                name="memory.process_list",
src/find_evil_mcp/server.py:147:                outputSchema=get_output_schema("memory.process_list"),
src/find_evil_mcp/server.py:150:                name="memory.malfind",
src/find_evil_mcp/server.py:167:                outputSchema=get_output_schema("memory.malfind"),
src/find_evil_mcp/server.py:170:                name="net.flow_summary",
src/find_evil_mcp/server.py:193:                outputSchema=get_output_schema("net.flow_summary"),
src/find_evil_mcp/server.py:196:                name="log.query",
src/find_evil_mcp/server.py:227:                outputSchema=get_output_schema("log.query"),
src/find_evil_mcp/server.py:230:                name="report.append",
src/find_evil_mcp/server.py:257:                outputSchema=get_output_schema("report.append"),
src/find_evil_mcp/server.py:260:                name="verify.cross_check",
src/find_evil_mcp/server.py:282:                outputSchema=get_output_schema("verify.cross_check"),
src/find_evil_mcp/server.py:290:            if name == "case.open":
src/find_evil_mcp/server.py:292:            elif name == "timeline.build":
src/find_evil_mcp/server.py:294:            elif name == "iocs.scan":
src/find_evil_mcp/server.py:296:            elif name == "memory.process_list":
src/find_evil_mcp/server.py:298:            elif name == "memory.malfind":
src/find_evil_mcp/server.py:300:            elif name == "net.flow_summary":
src/find_evil_mcp/server.py:302:            elif name == "log.query":
src/find_evil_mcp/server.py:304:            elif name == "report.append":
src/find_evil_mcp/server.py:306:            elif name == "verify.cross_check":
src/find_evil_mcp/server.py:329:    """Main entry point for stdio transport."""
src/find_evil_mcp/server.py:331:    async with server.stdio():
src/find_evil_mcp/tools/net_flow_summary.py:75:    Handle net.flow_summary tool invocation.
src/find_evil_mcp/tools/net_flow_summary.py:84:    validate_input("net.flow_summary", payload)
src/find_evil_mcp/tools/net_flow_summary.py:146:    validate_output("net.flow_summary", output)
src/find_evil_mcp/tools/log_query.py:78:    Handle log.query tool invocation.
src/find_evil_mcp/tools/log_query.py:87:    validate_input("log.query", payload)
src/find_evil_mcp/tools/log_query.py:114:    validate_output("log.query", output)
src/find_evil_mcp/tools/timeline_build.py:41:    Handle timeline.build tool invocation.
src/find_evil_mcp/tools/timeline_build.py:50:    validate_input("timeline.build", payload)
src/find_evil_mcp/tools/timeline_build.py:108:    validate_output("timeline.build", output)
src/find_evil_mcp/tools/iocs_scan.py:87:    Handle iocs.scan tool invocation.
src/find_evil_mcp/tools/iocs_scan.py:96:    validate_input("iocs.scan", payload)
src/find_evil_mcp/tools/iocs_scan.py:152:    validate_output("iocs.scan", output)
src/find_evil_mcp/tools/case_open.py:19:    Handle case.open tool invocation.
src/find_evil_mcp/tools/case_open.py:29:    validate_input("case.open", payload)
src/find_evil_mcp/tools/case_open.py:74:    validate_output("case.open", output)
src/find_evil_mcp/tools/memory_process_list.py:109:    Handle memory.process_list tool invocation.
src/find_evil_mcp/tools/memory_process_list.py:118:    validate_input("memory.process_list", payload)
src/find_evil_mcp/tools/memory_process_list.py:125:    mem_image_path = cases_root / case_id / "memory.dmp"
src/find_evil_mcp/tools/memory_process_list.py:157:        "mem_image_path": f"/cases/{case_id}/memory.dmp",
src/find_evil_mcp/tools/memory_process_list.py:163:    validate_output("memory.process_list", output)
src/find_evil_mcp/tools/report_append.py:27:    Handle report.append tool invocation.
src/find_evil_mcp/tools/report_append.py:39:    validate_input("report.append", payload)
src/find_evil_mcp/tools/report_append.py:127:    validate_output("report.append", output)
src/find_evil_mcp/tools/verify_cross_check.py:46:    Handle verify.cross_check tool invocation.
src/find_evil_mcp/tools/verify_cross_check.py:51:    Cross-check logic: re-verify findings via alternate tools (e.g., rerun timeline.build).
src/find_evil_mcp/tools/verify_cross_check.py:54:    validate_input("verify.cross_check", payload)
src/find_evil_mcp/tools/verify_cross_check.py:76:        # Rerun timeline.build to verify timeline consistency
src/find_evil_mcp/tools/verify_cross_check.py:81:                    agreements.append(f"timeline.build confirmed {timeline_result['event_count']} events")
src/find_evil_mcp/tools/verify_cross_check.py:109:    validate_output("verify.cross_check", output)
src/find_evil_mcp/tools/memory_malfind.py:110:    Handle memory.malfind tool invocation.
src/find_evil_mcp/tools/memory_malfind.py:119:    validate_input("memory.malfind", payload)
src/find_evil_mcp/tools/memory_malfind.py:125:    mem_image_path = cases_root / case_id / "memory.dmp"
src/find_evil_mcp/tools/memory_malfind.py:164:    validate_output("memory.malfind", output)

hook: PostToolUse
hook: PostToolUse Completed
hook: PostToolUse Completed
hook: PostToolUse Completed
hook: PreToolUse
hook: PreToolUse Completed
exec
/bin/bash -lc "nl -ba src/find_evil_mcp/server.py | sed -n '260,350p'" in /home/jang/Workspace/warvis-findEvil
 succeeded in 0ms:
   260	                name="verify.cross_check",
   261	                description="Re-verify findings against multiple sources for confidence scoring",
   262	                inputSchema={
   263	                    "type": "object",
   264	                    "properties": {
   265	                        "case_id": {
   266	                            "type": "string",
   267	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   268	                        },
   269	                        "finding_id": {
   270	                            "type": "string",
   271	                            "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$",
   272	                        },
   273	                        "method": {
   274	                            "type": "string",
   275	                            "enum": ["rerun", "alt_tool", "counter_evidence", "all"],
   276	                            "default": "all",
   277	                        },
   278	                    },
   279	                    "required": ["case_id", "finding_id"],
   280	                    "additionalProperties": False,
   281	                },
   282	                outputSchema=get_output_schema("verify.cross_check"),
   283	            ),
   284	        ]
   285	
   286	    @server.call_tool()
   287	    async def call_tool(name: str, arguments: dict):
   288	        """Call a tool by name."""
   289	        try:
   290	            if name == "case.open":
   291	                result = await handle_case_open(arguments)
   292	            elif name == "timeline.build":
   293	                result = await handle_timeline_build(arguments)
   294	            elif name == "iocs.scan":
   295	                result = await handle_iocs_scan(arguments)
   296	            elif name == "memory.process_list":
   297	                result = await handle_memory_process_list(arguments)
   298	            elif name == "memory.malfind":
   299	                result = await handle_memory_malfind(arguments)
   300	            elif name == "net.flow_summary":
   301	                result = await handle_net_flow_summary(arguments)
   302	            elif name == "log.query":
   303	                result = await handle_log_query(arguments)
   304	            elif name == "report.append":
   305	                result = await handle_report_append(arguments)
   306	            elif name == "verify.cross_check":
   307	                result = await handle_verify_cross_check(arguments)
   308	            else:
   309	                return TextContent(
   310	                    type="text",
   311	                    text=f"Unknown tool: {name}",
   312	                )
   313	
   314	            return TextContent(
   315	                type="text",
   316	                text=str(result),
   317	            )
   318	
   319	        except Exception as e:
   320	            return TextContent(
   321	                type="text",
   322	                text=f"Error calling tool {name}: {str(e)}",
   323	            )
   324	
   325	    return server
   326	
   327	
   328	async def main():
   329	    """Main entry point for stdio transport."""
   330	    server = create_server()
   331	    async with server.stdio():
   332	        await server.wait_closed()
   333	
   334	
   335	if __name__ == "__main__":
   336	    asyncio.run(main())

hook: PostToolUse
hook: PostToolUse Completed
web search: 
web search: site:github.com/ollama/ollama docs api tools function calling Ollama chat API tools
web search: 
web search: site:modelcontextprotocol.io specification initialize initialized notification stdio JSON-RPC MCP lifecycle
web search: 
web search: 'newline' in https://modelcontextprotocol.io/specification/2025-03-26/basic/transports
web search: 
web search: https://ollama.com/blog/structured-outputs?__from__=talkingdev
codex
**Recommendation: Request Changes**

No implementation blocker is impossible, but the design should be tightened before coding. The biggest risks are MCP protocol correctness, unbounded LLM control flow, weak path/process isolation details, and some Go package layering ambiguity.

**High Severity Findings**

1. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:640) — MCP lifecycle is incomplete.

The design sends `initialize`, then immediately `tools/list`, but MCP requires protocol version negotiation and an `notifications/initialized` notification after initialization. The request sketch also omits `protocolVersion` and client capabilities. Current MCP lifecycle docs require this sequence, and stdio messages must be newline-delimited valid JSON-RPC only.

Fix: model the full MCP handshake explicitly: `initialize` with protocol version, validate returned protocol version/capabilities, send `notifications/initialized`, then enter operation mode. Also support notifications, server errors, and version mismatch disconnect.

Sources: MCP lifecycle and stdio transport docs: https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle, https://modelcontextprotocol.io/specification/2025-03-26/basic/transports

2. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:730) — The stdio transport design is fragile.

`bufio.Scanner` has a default token limit, so large MCP responses can fail. `Recv(ctx)` starts a goroutine per receive; if the context cancels while `Scan()` blocks, the goroutine leaks and may race future reads. `Close()` closes stdin and waits, but does not drain stderr/stdout, check `Scanner.Err()`, or kill on timeout.

Fix: use one dedicated reader goroutine, `bufio.Reader.ReadBytes('\n')` or a scanner with explicit large buffer, demux responses by JSON-RPC id, handle notifications separately, capture stderr, and implement graceful shutdown with kill-after-timeout.

3. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:441) — Gemma function-calling assumptions are stale.

The doc assumes Gemma 4 does not have native function-calling-like support and uses prompt-only JSON parsing. Current Google docs describe Gemma 4 tool/function calling using tool schemas and tool-call formatting, and Ollama supports a `tools` field for supported models plus structured output via `format`. JSON prompting remains a valid fallback, but it should not be the primary unchecked path.

Fix: make the strategy layered:

1. Prefer Ollama `tools` if the selected Gemma 4 model exposes tool calls through Ollama.
2. Prefer Ollama `format` with a strict JSON Schema for the `Action` object when tool calls are unavailable.
3. Fall back to prompt-only JSON parsing only after probing model behavior at startup.

Sources: Gemma 4 function calling: https://ai.google.dev/gemma/docs/capabilities/text/function-calling-gemma4, Ollama tools: https://ollama.com/blog/tool-support, Ollama structured outputs: https://ollama.com/blog/structured-outputs

4. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:540) — Agent loop has no hard step budget despite retry policy.

Invalid JSON loops forever in the sample. Tool errors can also loop indefinitely because the code appends feedback and continues without enforcing the retry table. A forensic agent needs deterministic stop conditions.

Fix: add per-state budgets: max LLM turns, max invalid JSON attempts, max invalid args per tool, max repeated same tool+args, max state duration, and max total tool calls. Persist these counters in state so resume does not reset abuse loops.

5. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:939) — Path validation is insufficient for the stated security boundary.

`filepath.Abs()` plus `strings.Contains(abs, "..")` is not enough. `Abs` cleans paths, symlinks can escape, and the code does not enforce `/evidence/` or `/cases/<case_id>/`. It also conflicts with the table allowing relative evidence paths.

Fix: use `filepath.Clean`, `EvalSymlinks` where safe, strict allow roots, `filepath.Rel` containment checks, reject relative paths unless explicitly resolved under a configured evidence root, and re-check after file open if symlink races matter.

**Medium Severity Findings**

6. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:340) — Go module/package structure is directionally sound but internally inconsistent.

The tree has `internal/hunt/fsm.go`, `state.go`, `transitions.go`, but the matrix refers to packages like `hunt/fsm`, `hunt/state`, and `hunt/transitions`. In Go, those are separate packages only if they are separate directories. Also, `hunt/fsm -> display` creates a UI dependency inside domain logic.

Fix: either keep one `internal/hunt` package or split actual directories. Prefer domain packages depending on interfaces: `EventSink`, `StateStore`, `ToolExecutor`. Keep terminal display at the edge.

7. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:430) — “All packages ↔ config” is a global dependency smell.

Config should not be imported everywhere. It makes tests harder and couples protocol/domain code to runtime environment.

Fix: load config in `cmd/warvis`, pass typed options into constructors, and keep pure packages unaware of env/flags.

8. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:123) — FSM success criteria are too easy to satisfy.

TRACE can complete with only 10 parsed events, regardless of coverage or suspiciousness. EXPOSE accepts confidence `>=0.5`, and verification timeout records provisional confidence `0.5`, which can satisfy the threshold by default. That risks turning incomplete evidence into “verified enough.”

Fix: separate `complete`, `inconclusive`, `degraded`, and `failed` outcomes. Require explicit evidence coverage metrics and distinguish “no finding” from “unable to analyze.”

9. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:301) — Transition matrix is confusing and partly inconsistent.

The matrix labels “Current / Next” ambiguously, lists “RECURSIVE” for non-recursive phase changes, and includes `SCAN -> TRACE` in text without clearly showing its guard. This will lead to implementation drift.

Fix: specify transitions as typed rules: `From`, `To`, `Guard`, `Action`, `MaxCount`, `TerminalStatus`.

10. [docs/find-evil/warvis-go-architecture.md](/home/jang/Workspace/warvis-findEvil/docs/find-evil/warvis-go-architecture.md:589) — Tool output is fed back as a user message without untrusted-data isolation.

Forensic logs and tool outputs are attacker-controlled input. Feeding them directly into the model as normal user content invites prompt injection.

Fix: wrap tool output in a dedicated tool-result message if using Ollama tools, or use a strict envelope saying “untrusted forensic data, never instructions,” with truncation, escaping, and schema extraction before summarization.

11. [src/find_evil_mcp/server.py](/home/jang/Workspace/warvis-findEvil/src/find_evil_mcp/server.py:314) — The current Python server returns `str(result)`, not JSON.

The Go architecture expects JSON-ish MCP content, but `str(dict)` produces Python literals with single quotes. That will break robust JSON parsing and audit replay.

Fix either side of the contract: make the Python MCP return `json.dumps(result)` or make the Go client treat tool content as opaque text and only parse when `Content-Type`/schema says JSON.

**Security Boundary Gaps**

The design has good first principles: per-state tool whitelist, MCP schema validation, no shell tools, case sandbox, and audit logging. The missing pieces are operational hardening:

- Hash original evidence at ingest and store immutable evidence metadata before analysis.
- Run the MCP subprocess with constrained env, working directory, resource limits, and preferably container/user isolation.
- Redact or classify audit log fields; raw args/results may contain sensitive host paths, usernames, IPs, or secrets.
- Define append-only audit/report integrity, not just `FindingHash`.
- Add cancellation semantics for long SIFT subprocesses rather than relying on Go context alone.
- Treat Ollama as a local network boundary: bind localhost, configurable allowlist, no remote model endpoint unless explicitly enabled.

**Overall**

The architecture is promising and modular, but the implementation plan should be revised around explicit contracts: a real MCP lifecycle client, strict LLM action schema, bounded FSM execution, hardened path/process isolation, and clean Go dependency boundaries. The highest-value next edit is to replace the pseudo-code client/loop sections with concrete stateful protocols and budgets before writing the binary.
hook: Stop
hook: Stop Completed
tokens used
70,018

```

## Concise summary

Provider completed successfully. Review the raw output for details.

## Action items

- Review the response and extract decisions you want to apply.
- Capture follow-up implementation tasks if needed.
