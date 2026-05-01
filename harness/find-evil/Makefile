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
