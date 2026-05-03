#!/usr/bin/env bash
# PreCompact hook: snapshot .omc/state/ before Claude Code compacts context.
set -euo pipefail

STATE_DIR="${PWD}/.omc/state"
if [[ ! -d "$STATE_DIR" ]]; then exit 0; fi

TIMESTAMP=$(date -u +"%Y%m%dT%H%M%SZ")
python3 - <<PYEOF
import json, os
from pathlib import Path
state_dir = Path("${STATE_DIR}")
timestamp = "${TIMESTAMP}"
out = {"timestamp": timestamp, "state": {}}
for f in sorted(state_dir.glob("*.json")):
    if f.name.startswith("precompact-"):
        continue
    try:
        out["state"][f.name] = json.loads(f.read_text())
    except Exception:
        pass
(state_dir / f"precompact-{timestamp}.json").write_text(json.dumps(out, indent=2))
PYEOF
