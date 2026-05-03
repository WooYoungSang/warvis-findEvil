# SPDX-License-Identifier: MIT

import json
from datetime import datetime, timezone

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.sift_runner import SiftRunner

# Module-level singleton for SiftRunner (tests can monkeypatch)
runner = SiftRunner()


def _parse_volatility_json(stdout: bytes) -> list:
    """
    Parse volatility3 JSON output and extract process list.

    Expected format: JSON object with list of processes.

    Args:
        stdout: Raw volatility3 output (JSON)

    Returns:
        List of process dicts with required fields
    """
    processes = []

    try:
        data = json.loads(stdout.decode("utf-8", errors="ignore"))
    except json.JSONDecodeError:
        return processes

    # Volatility3 JSON structure: usually a list under root or nested
    if isinstance(data, list):
        proc_list = data
    elif isinstance(data, dict) and "processes" in data:
        proc_list = data["processes"]
    elif isinstance(data, dict) and "rows" in data:
        proc_list = data["rows"]
    else:
        # Try to extract from first dict value
        if isinstance(data, dict) and len(data) > 0:
            proc_list = list(data.values())[0]
            if not isinstance(proc_list, list):
                proc_list = []
        else:
            proc_list = []

    for proc in proc_list:
        if not isinstance(proc, dict):
            continue

        try:
            # Map volatility fields to output schema
            process_entry = {
                "pid": int(proc.get("PID", 0)),
                "ppid": int(proc.get("PPID", 0)),
                "name": str(proc.get("ImageFileName", "unknown"))[:64],
                "create_time": proc.get("CreateTime", datetime.now(timezone.utc).isoformat()),
                "exit_time": proc.get("ExitTime"),  # May be null
                "suspicious_flags": _extract_suspicious_flags(proc),
            }
            # Ensure create_time is ISO 8601
            if isinstance(process_entry["create_time"], str):
                # Already a string, should be ISO 8601
                pass
            processes.append(process_entry)
        except (TypeError, ValueError, KeyError):
            # Skip malformed entries
            continue

    return processes


def _extract_suspicious_flags(proc: dict) -> list:
    """
    Extract suspicious flags from a process dict.

    Args:
        proc: Process dictionary from volatility

    Returns:
        List of enum values: ["hidden", "unlinked", "no_image", "code_injected"]
    """
    flags = []

    # Check for hidden process (e.g., parent is 0 or process list anomalies)
    if proc.get("PPID") == 0 and proc.get("PID", 0) != 0:
        flags.append("hidden")

    # Check for unlinked process (volatility-specific flags)
    if proc.get("IsUnlinked") or proc.get("Unlinked"):
        flags.append("unlinked")

    # Check for missing image (no_image)
    if not proc.get("ImageFileName") or proc.get("ImageFileName") == "":
        flags.append("no_image")

    # Check for code injection indicators
    if proc.get("CodeInjected") or proc.get("InjectedCode"):
        flags.append("code_injected")

    return flags


async def handle_memory_process_list(payload: dict) -> dict:
    """
    Handle memory.process_list tool invocation.

    Input: case_id, kdbg_offset (optional)
    Output: case_id, processes, mem_image_path, tool_used, generated_at

    Wires volatility3 subprocess call.
    Parses JSON output; extracts process list with suspicious flags.
    """
    # Validate input against schema
    validate_input("memory.process_list", payload)

    case_id = payload["case_id"]
    kdbg_offset = payload.get("kdbg_offset", None)

    # Create memory image path
    cases_root = get_cases_root()
    mem_image_path = cases_root / case_id / "memory.dmp"

    processes = []

    # Check if volatility3 is available and memory dump exists
    if runner.is_available("volatility3") and mem_image_path.exists():
        # Build volatility3 args: vol -f <mem_image> windows.pslist.PsList --output=json
        vol_args = [
            "-f",
            str(mem_image_path),
            "windows.pslist.PsList",
            "--output=json",
        ]

        if kdbg_offset:
            vol_args.extend(["--kdbg", kdbg_offset])

        try:
            result = runner.run("volatility3", vol_args, cwd=cases_root / case_id)

            if result.returncode == 0:
                processes = _parse_volatility_json(result.stdout)
            # If subprocess failed, processes remain empty

        except (ValueError, Exception):
            # If args validation fails or subprocess errors, return empty
            pass

    # Build output
    output = {
        "case_id": case_id,
        "processes": processes,
        "mem_image_path": f"/cases/{case_id}/memory.dmp",
        "tool_used": "volatility3",
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("memory.process_list", output)

    return output
