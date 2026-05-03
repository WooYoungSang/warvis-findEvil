# SPDX-License-Identifier: MIT

import json
from datetime import datetime, timezone

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.sift_runner import SiftRunner

# Module-level singleton for SiftRunner (tests can monkeypatch)
runner = SiftRunner()


def _parse_malfind_json(stdout: bytes) -> tuple[list, int]:
    """
    Parse volatility3 malfind JSON output.

    Expected format: JSON list of findings with pid, protection, vad info.

    Args:
        stdout: Raw volatility3 malfind JSON output

    Returns:
        Tuple of (findings list, unique_pids count)
    """
    findings = []
    unique_pids = set()

    try:
        data = json.loads(stdout.decode("utf-8", errors="ignore"))
    except json.JSONDecodeError:
        return findings, 0

    # Volatility3 malfind JSON structure: usually a list
    if isinstance(data, list):
        finding_list = data
    elif isinstance(data, dict) and "findings" in data:
        finding_list = data["findings"]
    elif isinstance(data, dict) and "rows" in data:
        finding_list = data["rows"]
    else:
        finding_list = []

    for finding in finding_list:
        if not isinstance(finding, dict):
            continue

        try:
            pid = int(finding.get("PID", 0))
            unique_pids.add(pid)

            # Parse Start/End as hex strings if necessary
            start = finding.get("Start", 0)
            end = finding.get("End", 0)
            if isinstance(start, str):
                start = int(start, 16)
            else:
                start = int(start)
            if isinstance(end, str):
                end = int(end, 16)
            else:
                end = int(end)

            finding_entry = {
                "pid": pid,
                "vad_start": start,
                "vad_end": end,
                "protection": str(finding.get("Protect", "UNKNOWN"))[:32],
                "yara_hits": finding.get("HitList", []),
                "severity": _estimate_severity(finding),
            }
            findings.append(finding_entry)
        except (TypeError, ValueError, KeyError):
            # Skip malformed entries
            continue

    return findings, len(unique_pids)


def _estimate_severity(finding: dict) -> str:
    """
    Estimate severity of a malfind finding.

    Args:
        finding: Malfind record from volatility3

    Returns:
        Enum string: "info", "low", "medium", "high", "critical"
    """
    # Check for YARA hits (higher severity if hits)
    if finding.get("HitList") and len(finding.get("HitList", [])) > 0:
        return "high"

    # Check protection flags for suspicious patterns
    protect = str(finding.get("Protect", "")).upper()
    if "EXECUTE" in protect and "WRITE" in protect:
        # RWX is always suspicious
        return "high"
    elif "EXECUTE" in protect:
        # Executable but not image-backed is suspicious
        if "MEM_IMAGE" not in protect:
            return "medium"

    return "low"


async def handle_memory_malfind(payload: dict) -> dict:
    """
    Handle memory.malfind tool invocation.

    Input: case_id, pid (optional)
    Output: case_id, findings, scanned_pids, tool_used, generated_at

    Wires volatility3 malfind subprocess call.
    Parses JSON output; extracts suspicious memory regions.
    """
    # Validate input against schema
    validate_input("memory.malfind", payload)

    case_id = payload["case_id"]
    pid = payload.get("pid", None)

    cases_root = get_cases_root()
    mem_image_path = cases_root / case_id / "memory.dmp"

    findings = []
    scanned_pids = 0

    # Check if volatility3 is available and memory dump exists
    if runner.is_available("volatility3") and mem_image_path.exists():
        # Build volatility3 args: vol -f <mem_image> windows.malfind.Malfind --output=json
        vol_args = [
            "-f",
            str(mem_image_path),
            "windows.malfind.Malfind",
            "--output=json",
        ]

        if pid is not None:
            vol_args.extend(["--pid", str(pid)])

        try:
            result = runner.run("volatility3", vol_args, cwd=cases_root / case_id)

            if result.returncode == 0:
                findings, scanned_pids = _parse_malfind_json(result.stdout)
            # If subprocess failed, findings/scanned_pids remain empty

        except (ValueError, Exception):
            # If args validation fails or subprocess errors, return empty
            pass

    # Build output
    output = {
        "case_id": case_id,
        "findings": findings,
        "scanned_pids": scanned_pids,
        "tool_used": "volatility3",
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("memory.malfind", output)

    return output
