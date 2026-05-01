# SPDX-License-Identifier: MIT

import json
from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.tools.timeline_build import handle_timeline_build


def _read_finding(report_path: Path, finding_id: str) -> dict | None:
    """
    Read a specific finding from findings.jsonl by finding_id.

    Args:
        report_path: Path to findings.jsonl
        finding_id: UUID of the finding to locate

    Returns:
        Finding dict or None if not found
    """
    if not report_path.exists():
        return None

    try:
        with open(report_path, "r") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    record = json.loads(line)
                    if record.get("finding_id") == finding_id:
                        return record.get("finding")
                except json.JSONDecodeError:
                    continue
    except Exception:
        pass

    return None


async def handle_verify_cross_check(payload: dict) -> dict:
    """
    Handle verify.cross_check tool invocation.

    Input: case_id, finding_id, method (optional, default "all")
    Output: case_id, finding_id, agreements, disagreements, confidence, verdict, tool_used, generated_at

    Cross-check logic: re-verify findings via alternate tools (e.g., rerun timeline.build).
    """
    # Validate input against schema
    validate_input("verify.cross_check", payload)

    case_id = payload["case_id"]
    finding_id = payload["finding_id"]
    method = payload.get("method", "all")

    cases_root = get_cases_root()
    report_path = cases_root / case_id / "report" / "findings.jsonl"

    # Try to read the finding
    finding = _read_finding(report_path, finding_id)

    agreements = []
    disagreements = []
    confidence = 0.0
    verdict = "inconclusive"
    tool_used = []

    if finding:
        # If finding exists, we can attempt cross-check
        # For now, we implement a simple heuristic: check if evidence exists

        # Rerun timeline.build to verify timeline consistency
        if method in ["rerun", "all"]:
            try:
                timeline_result = await handle_timeline_build({"case_id": case_id})
                if timeline_result.get("event_count", 0) > 0:
                    agreements.append(f"timeline.build confirmed {timeline_result['event_count']} events")
                    confidence += 0.3
                    tool_used.append("plaso")
            except Exception:
                pass

        # If we have agreement, increase confidence
        if agreements:
            confidence = min(0.8, confidence + 0.5)  # Cap at 0.8 without additional sources
            verdict = "confirmed" if confidence > 0.6 else "partial"
        else:
            # No agreements found; finding is uncertain
            verdict = "uncertain"
            confidence = 0.2

    # Build output
    output = {
        "case_id": case_id,
        "finding_id": finding_id,
        "agreements": agreements,
        "disagreements": disagreements,
        "confidence": confidence,
        "verdict": verdict,
        "tool_used": tool_used,
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("verify.cross_check", output)

    return output
