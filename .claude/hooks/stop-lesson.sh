#!/usr/bin/env bash
# Stop hook: auto-call devos_prepare_lesson when Claude Code session ends.
set -euo pipefail

SESSION_FILE="${PWD}/.omc/state/current-session.json"
SESSION_ID=""
if [[ -f "$SESSION_FILE" ]]; then
  SESSION_ID=$(python3 -c "import json; d=json.load(open('$SESSION_FILE')); print(d.get('session_id',''))" 2>/dev/null || true)
fi
if [[ -z "$SESSION_ID" ]]; then exit 0; fi

claude -p "Call the devos_prepare_lesson MCP tool with session_id='${SESSION_ID}'. Do not do anything else." \
  --output-format text --max-turns 2 2>/dev/null || true
