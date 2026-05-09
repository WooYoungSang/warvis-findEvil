# ITEM-212 FIND EVIL — gate automation
# Run from repo root: make -C harness/find-evil <target>

SHELL := /usr/bin/env bash
ROOT  := $(shell git rev-parse --show-toplevel)
ENV   := $(ROOT)/harness/find-evil/.env
GATES := $(ROOT)/harness/find-evil/gates.yaml

.PHONY: help spec-check build test mcp-conformance accuracy demo-record tail-logs secrets-scan up down clean

help:
	@echo "Targets:"
	@echo "  spec-check       — assert spec.md + 7 agents + harness files exist"
	@echo "  up / down        — docker-compose lifecycle"
	@echo "  build            — install MCP server deps (no network in CI mode)"
	@echo "  test             — pytest suite + schema validation"
	@echo "  mcp-conformance  — capability handshake against running server"
	@echo "  accuracy         — run detector on fixtures, emit recall/precision"
	@echo "  demo-record      — placeholder; actual recording is human-driven"
	@echo "  secrets-scan     — pre-commit secret scan over staged files"
	@echo "  tail-logs        — tail JSONL traces in logs/"

spec-check:
	@test -f $(ROOT)/plans/ITEM-212-find-evil/spec.md         || { echo "MISSING: spec.md"; exit 1; }
	@test -f $(ROOT)/plans/ITEM-212-find-evil/handover-r1.md  || { echo "MISSING: handover-r1.md"; exit 1; }
	@for a in orchestrator mcp-architect mcp-implementer detector-engineer evidence-curator verifier pitch-writer; do \
	  test -f $(ROOT)/.claude/agents/find-evil-$$a.md || { echo "MISSING agent: $$a"; exit 1; }; \
	done
	@test -f $(GATES) || { echo "MISSING: gates.yaml"; exit 1; }
	@echo "spec-check: PASS"

up:
	docker compose -f $(ROOT)/harness/find-evil/docker-compose.yml --env-file $(ENV) up -d

down:
	docker compose -f $(ROOT)/harness/find-evil/docker-compose.yml down

VENV  := $(ROOT)/.venv-find-evil
PY    := $(VENV)/bin/python
PIP   := $(VENV)/bin/pip
PYTEST:= $(VENV)/bin/pytest

$(PY):
	python3 -m venv $(VENV)
	$(PIP) install -U pip wheel
	$(PIP) install -e $(ROOT)[dev]

build: $(PY)
	@echo "build: PASS"

test: $(PY)
	@cd $(ROOT) && $(PYTEST) tests/test_schema.py tests/test_server.py tests/test_detector.py -v

mcp-conformance: $(PY)
	@cd $(ROOT) && $(PY) $(ROOT)/harness/find-evil/scripts/mcp_handshake_check.py

accuracy: $(PY)
	@cd $(ROOT) && $(PY) $(ROOT)/harness/find-evil/scripts/accuracy.py

demo-record:
	@echo "Human-driven. Output expected at docs/find-evil/demo.mp4 (gitignored)."

secrets-scan:
	@$(ROOT)/harness/find-evil/scripts/secrets-scan.sh

tail-logs:
	@tail -F $(ROOT)/harness/find-evil/logs/*.jsonl

# ── Kill Switch Gates ──────────────────────────────────────────────────────────

WARVIS_BIN        := $(ROOT)/warvis/bin/warvis
CASES_TMP         := /tmp/warvis-ks1-cases
VENV_PY           := $(ROOT)/.venv-find-evil/bin/python
SERVER_CMD        := python3 $(ROOT)/harness/find-evil/mock_mcp_server.py
MOCK_OLLAMA_PORT  := 19134
MOCK_OLLAMA_CMD   := python3 $(ROOT)/harness/find-evil/mock_ollama_server.py $(MOCK_OLLAMA_PORT)
MOCK_OLLAMA_URL   := http://localhost:$(MOCK_OLLAMA_PORT)

.PHONY: kill-switch-check kill-switch-check-1 kill-switch-check-2 kill-switch-check-3 kill-switch-check-4 kill-switch-check-5

$(WARVIS_BIN):
	@cd $(ROOT)/warvis && go build -o bin/warvis ./cmd/warvis

kill-switch-check-1: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 1: INITIALIZE → TRACE ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  fuser -k $(MOCK_OLLAMA_PORT)/tcp 2>/dev/null || true; \
	  $(MOCK_OLLAMA_CMD) >/dev/null 2>&1 & MOCK_PID=$$!; \
	  sleep 0.3; \
	  TEST_IMG="$(ROOT)/evidence/test.img"; \
	  test -f "$$TEST_IMG" || { kill $$MOCK_PID 2>/dev/null; echo "FAILED: test evidence not found at $$TEST_IMG"; exit 1; }; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" OLLAMA_URL=$(MOCK_OLLAMA_URL) WARVIS_MAX_TURNS=3 $(WARVIS_BIN) hunt "$$TEST_IMG"); \
	  kill $$MOCK_PID 2>/dev/null || true; \
	  echo "warvis output: $$OUT"; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
	  test -n "$$CASE_ID" || { echo "FAILED: no case_id in output"; exit 1; }; \
	  STATE=$$(jq -r '.current_state' $(CASES_TMP)/$$CASE_ID/state.json); \
	  test "$$STATE" != "INITIALIZE" || { echo "FAILED: current_state=$$STATE (must advance from INITIALIZE)"; exit 1; }; \
	  test "$$STATE" != "null" || { echo "FAILED: current_state=null (state.json malformed)"; exit 1; }; \
	  LINES=$$(wc -l < $(CASES_TMP)/$$CASE_ID/audit.jsonl | tr -d ' '); \
	  test "$$LINES" -ge 2 || { echo "FAILED: audit.jsonl has $$LINES lines (want >=2)"; exit 1; }; \
	  jq . $(CASES_TMP)/$$CASE_ID/audit.jsonl > /dev/null || { echo "FAILED: audit.jsonl not valid JSONL"; exit 1; }; \
	  echo "PASSED: case_id=$$CASE_ID, state=$$STATE (advanced past INITIALIZE), audit_lines=$$LINES"

kill-switch-check-2: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 2: Tool calls in audit.jsonl ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  fuser -k $(MOCK_OLLAMA_PORT)/tcp 2>/dev/null || true; \
	  $(MOCK_OLLAMA_CMD) >/dev/null 2>&1 & MOCK_PID=$$!; \
	  sleep 0.3; \
	  TEST_IMG="$(ROOT)/evidence/test.img"; \
	  test -f "$$TEST_IMG" || { kill $$MOCK_PID 2>/dev/null; echo "FAILED: test evidence not found"; exit 1; }; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" OLLAMA_URL=$(MOCK_OLLAMA_URL) WARVIS_MAX_TURNS=3 $(WARVIS_BIN) hunt "$$TEST_IMG" 2>&1); \
	  kill $$MOCK_PID 2>/dev/null || true; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id' 2>/dev/null || echo ""); \
	  if [ -z "$$CASE_ID" ]; then \
	    CASE_ID=$$(ls -t $(CASES_TMP) | head -1); \
	    test -n "$$CASE_ID" || { echo "FAILED: no case directory found"; exit 1; }; \
	  fi; \
	  test -f $(CASES_TMP)/$$CASE_ID/audit.jsonl || { echo "FAILED: audit.jsonl not found for case $$CASE_ID"; exit 1; }; \
	  TOOL_CALLS=$$(grep -c '"event":"tool_called"' $(CASES_TMP)/$$CASE_ID/audit.jsonl || true); \
	  TOOL_RESULTS=$$(grep -c '"event":"tool_result"' $(CASES_TMP)/$$CASE_ID/audit.jsonl || true); \
	  test "$${TOOL_CALLS:-0}" -gt 0 || { echo "FAILED: no tool_called events (got $${TOOL_CALLS:-0})"; exit 1; }; \
	  test "$${TOOL_RESULTS:-0}" -gt 0 || { echo "FAILED: no tool_result events (got $${TOOL_RESULTS:-0})"; exit 1; }; \
	  echo "PASSED: tool_calls=$$TOOL_CALLS, tool_results=$$TOOL_RESULTS"

kill-switch-check-3: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 3: Gemma JSON autonomy ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  fuser -k $(MOCK_OLLAMA_PORT)/tcp 2>/dev/null || true; \
	  $(MOCK_OLLAMA_CMD) >/dev/null 2>&1 & MOCK_PID=$$!; \
	  sleep 0.3; \
	  TEST_IMG="$(ROOT)/evidence/test.img"; \
	  test -f "$$TEST_IMG" || { kill $$MOCK_PID 2>/dev/null; echo "FAILED: test evidence not found"; exit 1; }; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" OLLAMA_URL=$(MOCK_OLLAMA_URL) WARVIS_MAX_TURNS=3 $(WARVIS_BIN) hunt "$$TEST_IMG" 2>&1); \
	  kill $$MOCK_PID 2>/dev/null || true; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id' 2>/dev/null || echo ""); \
	  if [ -z "$$CASE_ID" ]; then \
	    CASE_ID=$$(ls -t $(CASES_TMP) | head -1); \
	    test -n "$$CASE_ID" || { echo "FAILED: no case directory"; exit 1; }; \
	  fi; \
	  TOOL_CALLS=$$(grep -c '"event":"tool_called"' $(CASES_TMP)/$$CASE_ID/audit.jsonl 2>/dev/null || true); \
	  test "$${TOOL_CALLS:-0}" -ge 2 || { echo "FAILED: only $${TOOL_CALLS:-0} tool calls (want >=2)"; exit 1; }; \
	  GEMMA_RESPONSES=$$(grep -c '"event":"gemma_response"' $(CASES_TMP)/$$CASE_ID/audit.jsonl 2>/dev/null || true); \
	  test "$${GEMMA_RESPONSES:-0}" -gt 0 || { echo "FAILED: no gemma_response events"; exit 1; }; \
	  echo "PASSED: tool_calls=$$TOOL_CALLS, gemma_responses=$$GEMMA_RESPONSES"

kill-switch-check-4: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 4: audit.jsonl JSONL integrity ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  fuser -k $(MOCK_OLLAMA_PORT)/tcp 2>/dev/null || true; \
	  $(MOCK_OLLAMA_CMD) >/dev/null 2>&1 & MOCK_PID=$$!; \
	  sleep 0.3; \
	  TEST_IMG="$(ROOT)/evidence/test.img"; \
	  test -f "$$TEST_IMG" || { kill $$MOCK_PID 2>/dev/null; echo "FAILED: test evidence not found"; exit 1; }; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" OLLAMA_URL=$(MOCK_OLLAMA_URL) WARVIS_MAX_TURNS=3 $(WARVIS_BIN) hunt "$$TEST_IMG" 2>&1); \
	  kill $$MOCK_PID 2>/dev/null || true; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id' 2>/dev/null || echo ""); \
	  if [ -z "$$CASE_ID" ]; then \
	    CASE_ID=$$(ls -t $(CASES_TMP) 2>/dev/null | head -1); \
	    test -n "$$CASE_ID" || { echo "FAILED: no case directory"; exit 1; }; \
	  fi; \
	  test -f $(CASES_TMP)/$$CASE_ID/audit.jsonl || { echo "FAILED: audit.jsonl not found"; exit 1; }; \
	  while IFS= read -r line; do \
	    echo "$$line" | jq . > /dev/null 2>&1 || { echo "FAILED: malformed JSON line in audit.jsonl"; exit 1; }; \
	  done < $(CASES_TMP)/$$CASE_ID/audit.jsonl; \
	  echo "PASSED: all audit.jsonl lines are valid JSON"

kill-switch-check-5: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 5: warvis status JSON output ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  fuser -k $(MOCK_OLLAMA_PORT)/tcp 2>/dev/null || true; \
	  $(MOCK_OLLAMA_CMD) >/dev/null 2>&1 & MOCK_PID=$$!; \
	  sleep 0.3; \
	  TEST_IMG="$(ROOT)/evidence/test.img"; \
	  test -f "$$TEST_IMG" || { kill $$MOCK_PID 2>/dev/null; echo "FAILED: test evidence not found"; exit 1; }; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" OLLAMA_URL=$(MOCK_OLLAMA_URL) WARVIS_MAX_TURNS=3 $(WARVIS_BIN) hunt "$$TEST_IMG" 2>&1); \
	  kill $$MOCK_PID 2>/dev/null || true; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id' 2>/dev/null || echo ""); \
	  if [ -z "$$CASE_ID" ]; then \
	    CASE_ID=$$(ls -t $(CASES_TMP) 2>/dev/null | head -1); \
	    test -n "$$CASE_ID" || { echo "FAILED: no case directory"; exit 1; }; \
	  fi; \
	  STATUS=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) $(WARVIS_BIN) status $$CASE_ID 2>&1); \
	  echo "$$STATUS" | jq -e '.current_state' > /dev/null 2>&1 || { echo "FAILED: missing or invalid current_state in status output"; exit 1; }; \
	  echo "PASSED: warvis status outputs valid JSON with current_state"

kill-switch-check: kill-switch-check-1 kill-switch-check-2 kill-switch-check-3 kill-switch-check-4 kill-switch-check-5
	@echo "✓ All 5 kill-switch tests PASSED"
