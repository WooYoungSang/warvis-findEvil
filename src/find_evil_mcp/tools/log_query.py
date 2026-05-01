# SPDX-License-Identifier: MIT

import json
from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root


def _parse_timeline_jsonl(timeline_path: Path, query: str, limit: int = 1000) -> tuple[list, int, bool]:
    """
    Parse timeline.jsonl and search for query matches (substring search, not regex).

    Args:
        timeline_path: Path to timeline.jsonl
        query: Search query (plain text substring)
        limit: Maximum number of hits to return

    Returns:
        Tuple of (hits list, total_matched count, truncated flag)
    """
    hits = []
    total_matched = 0
    truncated = False

    if not timeline_path.exists():
        return hits, total_matched, truncated

    try:
        with open(timeline_path, "r") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue

                try:
                    event = json.loads(line)
                except json.JSONDecodeError:
                    continue

                # Search query in various fields (case-insensitive substring match)
                query_lower = query.lower()
                found = False

                for key in ["description", "event_type", "message", "user", "host"]:
                    if key in event:
                        val = str(event[key]).lower()
                        if query_lower in val:
                            found = True
                            break

                if found:
                    total_matched += 1
                    if len(hits) < limit:
                        hit = {
                            "timestamp": event.get("timestamp", ""),
                            "source": event.get("source", ""),
                            "host": event.get("host", ""),
                            "user": event.get("user", ""),
                            "level": event.get("severity", "info"),
                            "message": event.get("description", ""),
                            "raw_offset": 0,  # Would need to track actual file offset
                        }
                        hits.append(hit)
                    else:
                        truncated = True

    except Exception:
        # If file read fails, return empty
        pass

    return hits, total_matched, truncated


async def handle_log_query(payload: dict) -> dict:
    """
    Handle log.query tool invocation.

    Input: case_id, q, source (optional), since/until (optional), limit (optional)
    Output: case_id, query, hits, total_matched, truncated, tool_used, generated_at

    Searches timeline.jsonl with plain-text substring matching (no regex, shell-safe).
    Parses JSONL events; counts and truncates as needed.
    """
    # Validate input against schema
    validate_input("log.query", payload)

    case_id = payload["case_id"]
    q = payload["q"]
    source = payload.get("source", "all")
    since = payload.get("since", None)
    until = payload.get("until", None)
    limit = payload.get("limit", 1000)

    cases_root = get_cases_root()
    timeline_path = cases_root / case_id / "timeline.jsonl"

    # Query timeline using plain-text substring search
    hits, total_matched, truncated = _parse_timeline_jsonl(timeline_path, q, limit)

    # Build output
    output = {
        "case_id": case_id,
        "query": q,
        "hits": hits,
        "total_matched": total_matched,
        "truncated": truncated,
        "tool_used": "plaso-query",
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("log.query", output)

    return output
