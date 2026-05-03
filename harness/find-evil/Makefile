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

WARVIS_BIN   := $(ROOT)/warvis/bin/warvis
CASES_TMP    := /tmp/warvis-ks1-cases
VENV_PY      := $(ROOT)/.venv-find-evil/bin/python
SERVER_CMD   := $(VENV_PY) $(ROOT)/scripts/run_mcp_server.py

.PHONY: kill-switch-check kill-switch-check-1

$(WARVIS_BIN):
	@cd $(ROOT)/warvis && go build -o bin/warvis ./cmd/warvis

kill-switch-check-1: $(WARVIS_BIN)
	@echo "=== Kill Switch Test 1: INITIALIZE → TRACE ==="
	@rm -rf $(CASES_TMP) && mkdir -p $(CASES_TMP)
	@set -e; \
	  OUT=$$(FIND_EVIL_CASES_ROOT=$(CASES_TMP) FIND_EVIL_SERVER_CMD="$(SERVER_CMD)" $(WARVIS_BIN) hunt /evidence/test.img); \
	  echo "warvis output: $$OUT"; \
	  CASE_ID=$$(echo "$$OUT" | jq -r '.case_id'); \
	  test -n "$$CASE_ID" || { echo "FAILED: no case_id in output"; exit 1; }; \
	  STATE=$$(jq -r '.current_state' $(CASES_TMP)/$$CASE_ID/state.json); \
	  test "$$STATE" = "TRACE" || { echo "FAILED: current_state=$$STATE (want TRACE)"; exit 1; }; \
	  LINES=$$(wc -l < $(CASES_TMP)/$$CASE_ID/audit.jsonl | tr -d ' '); \
	  test "$$LINES" -ge 2 || { echo "FAILED: audit.jsonl has $$LINES lines (want >=2)"; exit 1; }; \
	  jq . $(CASES_TMP)/$$CASE_ID/audit.jsonl > /dev/null || { echo "FAILED: audit.jsonl not valid JSONL"; exit 1; }; \
	  echo "PASSED: case_id=$$CASE_ID, state=TRACE, audit_lines=$$LINES"

kill-switch-check: kill-switch-check-1
	@echo "All kill-switch tests PASSED"
