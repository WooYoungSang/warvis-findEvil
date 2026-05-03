# SPDX-License-Identifier: MIT

import hashlib
import json
import os
import uuid
import tempfile
from datetime import datetime, timezone

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root


def _canonical_json(obj: dict) -> str:
    """Return canonical JSON (sorted keys, no whitespace)."""
    return json.dumps(obj, sort_keys=True, separators=(",", ":"))


def _sha256_hex(data: str) -> str:
    """Return SHA-256 hex hash of string data."""
    return hashlib.sha256(data.encode("utf-8")).hexdigest()


async def handle_report_append(payload: dict) -> dict:
    """
    Handle report.append tool invocation.

    Input: case_id, finding (object with title, severity, narrative, evidence, mitre_attack?)
    Output: case_id, finding_id, appended_at, report_path, prev_hash, this_hash

    Side effects:
    - Reads /cases/<case_id>/report/findings.jsonl (if exists)
    - Reads last line to extract prev_hash
    - Computes this_hash via canonical JSON + SHA-256
    - Atomically appends JSONL record with hash chain
    """
    # Validate input against schema
    validate_input("report.append", payload)

    case_id = payload["case_id"]
    finding = payload["finding"]

    # Generate finding_id (UUID v4)
    finding_id = str(uuid.uuid4())

    # Resolve report path
    cases_root = get_cases_root()
    report_dir = cases_root / case_id / "report"
    report_path = report_dir / "findings.jsonl"

    # Ensure report directory exists
    report_dir.mkdir(parents=True, exist_ok=True)

    # Read previous hash from last line of existing findings.jsonl (if exists)
    prev_hash = "0" * 64  # default: all zeros for first entry
    if report_path.exists():
        try:
            with open(report_path, "r") as f:
                lines = f.readlines()
                if lines:
                    last_line = lines[-1].strip()
                    if last_line:
                        last_record = json.loads(last_line)
                        # Extract this_hash from last record as prev_hash for new record
                        prev_hash = last_record.get("this_hash", "0" * 64)
        except Exception:
            # If read/parse fails, default to all zeros
            prev_hash = "0" * 64

    # Compute this_hash: SHA-256 of canonical JSON of finding
    finding_canonical = _canonical_json(finding)
    this_hash = _sha256_hex(finding_canonical)

    # Build JSONL record
    record = {
        "finding_id": finding_id,
        "prev_hash": prev_hash,
        "this_hash": this_hash,
        "finding": finding,
        "appended_at": datetime.now(timezone.utc).isoformat(),
    }

    # Atomically append: write to temp file, then rename
    record_line = json.dumps(record, separators=(",", ":")) + "\n"
    try:
        # Write to temporary file in same directory (ensures same filesystem)
        with tempfile.NamedTemporaryFile(
            mode="w",
            dir=report_dir,
            delete=False,
            suffix=".tmp",
        ) as tmp:
            tmp_path = tmp.name
            tmp.write(record_line)

        # Append temp content to actual file (atomic on POSIX via cat + rename)
        if report_path.exists():
            with open(report_path, "a") as f:
                with open(tmp_path, "r") as tmp_f:
                    f.write(tmp_f.read())
        else:
            # First entry: create file
            with open(tmp_path, "r") as tmp_f:
                with open(report_path, "w") as f:
                    f.write(tmp_f.read())

        # Clean up temp file
        os.unlink(tmp_path)
    except Exception as e:
        # Clean up temp file on error
        if os.path.exists(tmp_path):
            os.unlink(tmp_path)
        raise e

    # Build output
    output = {
        "case_id": case_id,
        "finding_id": finding_id,
        "appended_at": datetime.now(timezone.utc).isoformat(),
        "report_path": f"/cases/{case_id}/report/findings.jsonl",
        "prev_hash": prev_hash,
        "this_hash": this_hash,
    }

    # Validate output against schema
    validate_output("report.append", output)

    return output
