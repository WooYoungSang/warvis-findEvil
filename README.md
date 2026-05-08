# W.A.R.V.I.S — Find Evil

> *The smallest IR agent whose architecture — not its prompt — guarantees it cannot escape its forensic role*

**W.A.R.V.I.S** (Woops, A Rather Very Intelligent System) is a single Go binary that orchestrates automated digital forensics using a deterministic 5-state FSM (Finite State Machine) and budget-preserving resume capability. Built for airgapped incident response environments where reliability and reproducibility matter more than breadth.

**Submission**: SANS FIND EVIL Hackathon | **Deadline**: 2026-06-15 | **Author**: WoopsFactory

## Why warvis (vs Valhuntir)

We acknowledge **Valhuntir** as the comprehensive DFIR reference platform (90+ tools, 22K-record RAG, Sigma rules, multi-backend orchestration). We do not claim to match its breadth. Instead, warvis focuses on three narrow differentiators:

1. **Single Go binary** — No Python runtime, no Docker, no gateway. Deployable offline on a fresh SIFT VM in seconds.
2. **5-state Hunt FSM with compile-time tool whitelist** — Architectural guarantee that the agent cannot escape its forensic role, enforced at build time (not runtime).
3. **Budget-preserving resume + kill-switch harness** — Every hunt is reproducible and verifiable in CI; interrupted hunts resume without re-burning LLM-turn quota.

For forensic teams in airgapped or regulated environments, this narrow approach is a deployable answer. For open settings where breadth dominates, Valhuntir is the reference. See [docs/find-evil/valhuntir-comparison.md](./docs/find-evil/valhuntir-comparison.md) for the full honest positioning matrix.

---

## Overview

W.A.R.V.I.S combines a Python MCP server (9 forensic tools) with a Go orchestrator bridge to provide **end-to-end automated incident response**. The system:

- **INITIALIZE**: Opens a forensic case and ingests evidence
- **TRACE**: Builds timeline and queries structured logs
- **SCAN**: Scans for indicators of compromise (IoCs) and memory artifacts
- **EXPOSE**: Cross-checks findings and validates hypotheses
- **LOCK**: Finalizes audit trail and generates forensic report

The Gemma 4 LLM (via Ollama) autonomously interprets tool outputs and selects the next investigative step.

---

## Prerequisites

### Required

- **Go** 1.21+ (for warvis bridge)
- **Python** 3.10+ (for MCP server)
- **Ollama** (running at `localhost:29134`)
  - Model: `gemma4:26b-a4b-it-q4_K_M`
  - Setup: `ollama pull gemma4:26b-a4b-it-q4_K_M && ollama serve`

### Verified Platforms

- Linux (Ubuntu 22.04+, Debian 12+)
- macOS (M1/M2, Intel; Ollama may require higher RAM)
- Windows (WSL2 with Linux kernel)

---

## Installation

### 1. Clone the Repository

```bash
git clone https://github.com/WoopsFactory/warvis-findEvil.git
cd warvis-findEvil
```

### 2. Set Up Python MCP Server

```bash
# Create virtual environment
python3.10 -m venv .venv
source .venv/bin/activate

# Install dependencies
pip install -e .[dev]

# Verify MCP server is importable
python -c "from find_evil_mcp import MCP"
```

### 3. Build Go Bridge

```bash
cd warvis
go mod download
go build -o bin/warvis ./cmd/warvis
cd ..
```

### 4. Verify Installation

```bash
# Test Go binary
./warvis/bin/warvis --version

# Test Python environment
python -m pytest --co -q
```

---

## Quick Start

### Prepare Evidence

Place forensic evidence (disk images, memory dumps, log archives) in the `evidence/` directory:

```bash
mkdir -p evidence
# Add your evidence files (e.g., evidence/test.img, evidence/memory.bin)
```

### Run Hunt

```bash
./warvis/bin/warvis hunt evidence/test.img
```

**Output**: Case directory `/cases/<case_id>/` with audit trail:

```
/cases/<case_id>/
  audit.jsonl           # FSM state transitions + tool calls
  report.json           # Forensic findings (EXPOSE phase)
  timeline.jsonl        # Event timeline (TRACE phase)
```

### Inspect Results

```bash
# View FSM progression
jq .event_type /cases/<case_id>/audit.jsonl | sort | uniq -c

# Extract findings
jq '.findings' /cases/<case_id>/report.json | head -20
```

### Resume Interrupted Hunt

If a hunt is interrupted (network, Ollama timeout), resume with:

```bash
./warvis/bin/warvis hunt evidence/test.img --case-id <case_id> --resume
```

---

## Commands Reference

### warvis hunt

Start a forensic investigation on evidence file(s).

```bash
warvis hunt <evidence_path> [--case-id <id>] [--resume]
```

**Options**:
- `--case-id`: Use existing case ID (resume mode)
- `--resume`: Continue from last checkpoint

**Example**:
```bash
warvis hunt evidence/disk.img
# Output: Case opened with case_id=abc123def456
```

### warvis status

Check FSM state of an active or completed hunt.

```bash
warvis status <case_id>
```

**Output**:
```
Case: abc123def456
State: SCAN (3/5)
Last Tool: iocs.scan
Progress: 60%
Timestamp: 2026-05-07T10:23:45Z
```

### warvis report

Generate final forensic report.

```bash
warvis report <case_id> [--format json|html]
```

---

## Architecture

### FSM States & Autonomy

| State | LLM Autonomy | Tools | Purpose |
|-------|---|---|---|
| INITIALIZE | 0% | `case.open` | Ingest evidence, compute hashes |
| TRACE | 70% | `timeline.build`, `log.query` | Build event timeline, identify anomalies |
| SCAN | 90% | `iocs.scan`, `memory.*`, `net.*` | Hunt indicators, scan memory for artifacts |
| EXPOSE | 60% | `verify.cross_check` | Correlate findings, validate hypotheses |
| LOCK | 0% | `report.append` | Finalize audit trail, seal case |

### Component Stack

- **MCP Server** (Python): 9 forensic tools exposing `case`, `timeline`, `log`, `iocs`, `memory`, `network`, `verify`, `report` namespaces
- **Go Bridge** (Hunt Orchestrator): Manages FSM state, runs Ollama tool-call loop, maintains audit trail
- **Gemma 4 LLM** (Ollama): Interprets tool outputs, selects next investigative action
- **Audit Trail** (JSONL): Immutable record of all FSM transitions, tool calls, and LLM reasoning

---

## Limitations

### Known Constraints

1. **Lite Scanners**: YARA and Plaso tools are emulated with synthetic patterns (not production-grade)
2. **Synthetic Fixtures**: Test evidence is auto-generated; real DFIR samples pending
3. **Agent Autonomy**: LLM decision gates (Tests 2 & 3) still in progress; some tool invocations require manual gates
4. **Ollama Dependency**: Requires local Ollama instance; no cloud LLM support yet
5. **Recall**: ~60% on synthetic test suite; accuracy >80% pending real dataset tuning

### Unsupported Features

- ⛔ Distributed forensics (multi-node investigation)
- ⛔ Real-time network traffic analysis (PCAP files only)
- ⛔ Encrypted partition analysis (LUKS, BitLocker)
- ⛔ Cloud forensics (S3, Azure Blob, GCS integrations)

---

## Contributing

W.A.R.V.I.S welcomes contributions! See [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

**Contributors**: WoopsFactory (author)  
**Hackathon**: SANS FIND EVIL 2026  
**License**: MIT (see [LICENSE](./LICENSE))

---

## Testing

### Python Tests

```bash
python -m pytest tests/ -v
python -m pytest --cov=src/find_evil_mcp/
```

### Go Tests

```bash
cd warvis
go test ./...
```

### Integration Test

```bash
make -C harness/find-evil test-hunt
```

---

## Troubleshooting

### "Ollama connection refused"

Ensure Ollama is running:

```bash
ollama serve
# In another terminal:
ollama pull gemma4:26b-a4b-it-q4_K_M
```

### "MCP server import failed"

Verify Python environment:

```bash
source .venv/bin/activate
pip install -e .[dev]
python -c "from find_evil_mcp import MCP; print(MCP.__version__)"
```

### "Go build fails: module not found"

Ensure you're in the `warvis/` directory:

```bash
cd warvis
go mod tidy
go build -o bin/warvis ./cmd/warvis
```

---

## References

- **Architecture**: [docs/find-evil/warvis-go-architecture.md](./docs/find-evil/warvis-go-architecture.md)
- **MCP Spec**: https://modelcontextprotocol.io
- **Ollama**: https://ollama.ai
- **SANS FIND EVIL**: https://findevil.devpost.com/

---

**W.A.R.V.I.S — Turning digital chaos into forensic clarity.**
