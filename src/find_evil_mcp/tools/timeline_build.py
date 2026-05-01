# SPDX-License-Identifier: MIT

import json
from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.sift_runner import SiftRunner

# Module-level singleton for SiftRunner (tests can monkeypatch)
runner = SiftRunner()


def _parse_plaso_jsonl(stdout: bytes) -> int:
    """
    Parse plaso JSONL output and count events.

    Args:
        stdout: Raw stdout from plaso (JSON-Lines format)

    Returns:
        Count of valid JSON lines parsed
    """
    count = 0
    for line in stdout.decode("utf-8", errors="ignore").split("\n"):
        line = line.strip()
        if not line:
            continue
        try:
            json.loads(line)
            count += 1
        except json.JSONDecodeError:
            # Skip malformed lines
            pass
    return count


async def handle_timeline_build(payload: dict) -> dict:
    """
    Handle timeline.build tool invocation.

    Input: case_id, source_filter (optional), since/until (optional)
    Output: case_id, timeline_path, event_count, tool_used, generated_at, sources_processed

    Wires plaso/log2timeline subprocess call with whitelisted args.
    Parses JSONL output; counts events.
    """
    # Validate input against schema
    validate_input("timeline.build", payload)

    case_id = payload["case_id"]
    source_filter = payload.get("source_filter", "all")
    since = payload.get("since")
    until = payload.get("until")

    # Create timeline file path
    cases_root = get_cases_root()
    timeline_path = cases_root / case_id / "timeline.jsonl"

    # Ensure case directory exists
    timeline_path.parent.mkdir(parents=True, exist_ok=True)

    event_count = 0
    sources_processed = [] if source_filter == "all" else [source_filter]

    # Check if plaso is available
    if runner.is_available("plaso"):
        # Build plaso args
        plaso_args = [
            "--status_view=none",
            "-q",
            "-o=jsonl",
            f"--output-file={timeline_path}",
            str(timeline_path.parent / "evidence"),  # Evidence dir
        ]

        try:
            result = runner.run("plaso", plaso_args, cwd=timeline_path.parent)

            if result.returncode == 0:
                # Parse output to count events
                event_count = _parse_plaso_jsonl(result.stdout)
                # If no stdout, try reading the generated file
                if event_count == 0 and timeline_path.exists():
                    with open(timeline_path, "rb") as f:
                        event_count = _parse_plaso_jsonl(f.read())
            # If subprocess failed, event_count remains 0

        except (ValueError, Exception):
            # If args validation fails or subprocess errors, fall back to empty
            timeline_path.touch()
    else:
        # Tool not available: create empty file
        timeline_path.touch()

    # Build output
    output = {
        "case_id": case_id,
        "timeline_path": f"/cases/{case_id}/timeline.jsonl",
        "event_count": event_count,
        "tool_used": "plaso",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "sources_processed": sources_processed,
    }

    # Validate output against schema
    validate_output("timeline.build", output)

    return output
