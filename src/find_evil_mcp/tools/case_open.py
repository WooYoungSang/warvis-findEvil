# SPDX-License-Identifier: MIT

import os
import uuid
from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output


def get_cases_root() -> Path:
    """Get cases root directory from env or default."""
    cases_root = os.environ.get("FIND_EVIL_CASES_ROOT", "./.cases")
    return Path(cases_root)


async def handle_case_open(payload: dict) -> dict:
    """
    Handle case.open tool invocation.

    Input: exactly one of {image_path, mem_path, pcap_path, logdir_path}
    Output: case_id (uuid4), sandbox_root, accepted_inputs, created_at

    Side effects:
    - Creates /cases/<case_id>/ directory tree
    - Creates /cases/<case_id>/report/ subdirectory
    """
    # Validate input against schema
    validate_input("case.open", payload)

    # Determine input kind and path
    input_kind = None
    input_path = None

    if "image_path" in payload:
        input_kind = "disk_image"
        input_path = payload["image_path"]
    elif "mem_path" in payload:
        input_kind = "memory_dump"
        input_path = payload["mem_path"]
    elif "pcap_path" in payload:
        input_kind = "pcap"
        input_path = payload["pcap_path"]
    elif "logdir_path" in payload:
        input_kind = "log_directory"
        input_path = payload["logdir_path"]

    # Generate case_id (UUID v4)
    case_id = str(uuid.uuid4())

    # Create sandbox root
    cases_root = get_cases_root()
    sandbox_root = cases_root / case_id
    sandbox_root.mkdir(parents=True, exist_ok=True)

    # Create report subdirectory
    report_dir = sandbox_root / "report"
    report_dir.mkdir(parents=True, exist_ok=True)

    # Build output
    output = {
        "case_id": case_id,
        "sandbox_root": f"/cases/{case_id}/",
        "accepted_inputs": [
            {
                "kind": input_kind,
                "path": input_path,
            }
        ],
        "created_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("case.open", output)

    return output
