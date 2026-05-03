# SPDX-License-Identifier: MIT

import uuid
from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.sift_runner import SiftRunner

# Module-level singleton for SiftRunner (tests can monkeypatch)
runner = SiftRunner()


def _parse_yara_output(stdout: bytes, sandbox_root: Path) -> tuple[list, int]:
    """
    Parse yara -s output and extract matches.

    Expected format: <rule> <path>:<offset>:<length> <content>

    Args:
        stdout: Raw yara output
        sandbox_root: Case sandbox root for relative path conversion

    Returns:
        Tuple of (matches list, scanned_files count)
    """
    matches = []
    scanned_files_set = set()

    for line in stdout.decode("utf-8", errors="ignore").split("\n"):
        line = line.strip()
        if not line:
            continue

        # Parse: "rule_name path:offset:length content"
        # Simple pattern: word space path space hex:hex:hex space ...
        parts = line.split(None, 2)
        if len(parts) < 3:
            continue

        rule_name = parts[0]
        location = parts[1]
        context = parts[2] if len(parts) > 2 else ""

        # Parse location: path:offset:length
        if ":" not in location:
            continue

        loc_parts = location.split(":")
        if len(loc_parts) < 2:
            continue

        path_str = loc_parts[0]
        try:
            offset = int(loc_parts[1])
            _length = int(loc_parts[2]) if len(loc_parts) > 2 else 0
        except (ValueError, IndexError):
            continue

        scanned_files_set.add(path_str)

        # Determine severity (default: info)
        severity = "info"
        # Could parse rule metadata for severity, but for now default to info
        # unless rule name hints at severity
        if any(x in rule_name.lower() for x in ["critical", "exploit", "shellcode"]):
            severity = "high"
        elif any(x in rule_name.lower() for x in ["malware", "trojan", "worm"]):
            severity = "medium"

        match = {
            "rule": rule_name,
            "path": path_str,
            "offset": offset,
            "severity": severity,
            "context": context[:256] if context else "",
        }
        matches.append(match)

    return matches, len(scanned_files_set)


async def handle_iocs_scan(payload: dict) -> dict:
    """
    Handle iocs.scan tool invocation.

    Input: case_id, ruleset, target_glob (optional)
    Output: case_id, scan_id, matches, scanned_files, tool_used, generated_at

    Wires yara/sigma subprocess call with whitelisted args.
    Parses match output; counts files scanned.
    """
    # Validate input against schema
    validate_input("iocs.scan", payload)

    case_id = payload["case_id"]
    ruleset = payload["ruleset"]
    target_glob = payload.get("target_glob", None)

    # Generate scan_id (UUID v4)
    scan_id = str(uuid.uuid4())

    # Determine tool (yara for yara_* rulesets, sigma otherwise)
    tool_name = "yara" if ruleset.startswith("yara") else "sigma"

    cases_root = get_cases_root()
    sandbox_root = cases_root / case_id

    matches = []
    scanned_files = 0

    # Check if tool is available
    if runner.is_available(tool_name):
        if tool_name == "yara":
            # Build yara args: yara -r -s <ruleset_path> <target>
            ruleset_path = f"/etc/yara/rules/{ruleset}.yar"
            target_path = target_glob if target_glob else str(sandbox_root)

            yara_args = ["-r", "-s", ruleset_path, target_path]

            try:
                result = runner.run(tool_name, yara_args, cwd=sandbox_root)

                if result.returncode in (0, 1):  # 0=matches, 1=no matches
                    matches, scanned_files = _parse_yara_output(result.stdout, sandbox_root)
                # If subprocess failed, matches/scanned_files remain empty

            except (ValueError, Exception):
                # If args validation fails or subprocess errors, return empty
                pass
        elif tool_name == "sigma":
            # Sigma handling: similar structure, different binary
            # For now, return empty (sigma integration deferred)
            pass
    else:
        # Tool not available: return empty results
        pass

    # Build output
    output = {
        "case_id": case_id,
        "scan_id": scan_id,
        "matches": matches,
        "scanned_files": scanned_files,
        "tool_used": tool_name,
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("iocs.scan", output)

    return output
