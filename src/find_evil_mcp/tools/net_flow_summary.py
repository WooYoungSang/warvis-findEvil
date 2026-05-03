# SPDX-License-Identifier: MIT

from datetime import datetime, timezone
from pathlib import Path

from find_evil_mcp.schema_loader import validate_input, validate_output
from find_evil_mcp.tools.case_open import get_cases_root
from find_evil_mcp.sift_runner import SiftRunner

# Module-level singleton for SiftRunner (tests can monkeypatch)
runner = SiftRunner()


def _parse_zeek_conn_log(conn_log_path: Path, top_n: int = 50) -> tuple[list, int]:
    """
    Parse Zeek conn.log (TSV format) and extract flow summary.

    Expected format: TSV with columns id.orig_h, id.orig_p, id.resp_h, id.resp_p,
    proto, orig_bytes, resp_bytes, orig_pkts, resp_pkts

    Args:
        conn_log_path: Path to conn.log file
        top_n: Limit to top N flows by bytes

    Returns:
        Tuple of (flows list sorted by bytes desc, total_flows count)
    """
    flows = []

    if not conn_log_path.exists():
        return flows, 0

    try:
        with open(conn_log_path, "r") as f:
            lines = f.readlines()
    except Exception:
        return flows, 0

    # Parse TSV: skip comments and header
    for line in lines:
        line = line.strip()
        if not line or line.startswith("#"):
            continue

        try:
            fields = line.split("\t")
            if len(fields) < 9:
                continue

            # Map TSV columns to output schema (field indices may vary)
            # Standard: id.orig_h, id.orig_p, id.resp_h, id.resp_p, proto,
            # duration, orig_bytes, resp_bytes, ...
            flow = {
                "src_ip": fields[0],
                "src_port": int(fields[1]),
                "dst_ip": fields[2],
                "dst_port": int(fields[3]),
                "proto": fields[4],
                "bytes": int(fields[6]) + int(fields[7]),  # orig_bytes + resp_bytes
                "packets": int(fields[8]),  # orig_pkts (simplified)
                "suspicious_flags": [],
            }
            flows.append(flow)
        except (IndexError, ValueError):
            # Skip malformed lines
            continue

    # Sort by bytes descending, take top_n
    flows.sort(key=lambda f: f["bytes"], reverse=True)
    flows = flows[:top_n]

    return flows, len(flows)


async def handle_net_flow_summary(payload: dict) -> dict:
    """
    Handle net.flow_summary tool invocation.

    Input: case_id, pcap_id, top_n (optional, default 50)
    Output: case_id, pcap_id, flows, total_flows, tool_used, generated_at

    Wires zeek subprocess call to process PCAP.
    Parses conn.log TSV output; extracts top N flows by bytes.
    """
    # Validate input against schema
    validate_input("net.flow_summary", payload)

    case_id = payload["case_id"]
    pcap_id = payload["pcap_id"]
    top_n = payload.get("top_n", 50)

    cases_root = get_cases_root()
    sandbox_root = cases_root / case_id

    # Infer PCAP file path from pcap_id or look in evidence dir
    pcap_path = sandbox_root / f"{pcap_id}.pcap"
    if not pcap_path.exists():
        # Try evidence subdir
        evidence_dir = sandbox_root / "evidence"
        if evidence_dir.exists():
            for f in evidence_dir.glob("*.pcap"):
                pcap_path = f
                break

    flows = []
    total_flows = 0

    # Check if zeek is available and PCAP exists
    if runner.is_available("zeek") and pcap_path.exists():
        # Create zeek output directory
        zeek_outdir = sandbox_root / "zeek_output"
        zeek_outdir.mkdir(exist_ok=True)

        # Build zeek args: zeek -r <pcap> -C -d -l zeek_output
        zeek_args = [
            "-r",
            str(pcap_path),
            "-C",  # Ignore checksums
            "-d",  # Don't run default scripts
            "-l",
            str(zeek_outdir),
        ]

        try:
            result = runner.run("zeek", zeek_args, cwd=sandbox_root)

            if result.returncode == 0:
                # Parse conn.log output
                conn_log = zeek_outdir / "conn.log"
                flows, total_flows = _parse_zeek_conn_log(conn_log, top_n)
            # If subprocess failed, flows remain empty

        except (ValueError, Exception):
            # If args validation fails or subprocess errors, return empty
            pass

    # Build output
    output = {
        "case_id": case_id,
        "pcap_id": pcap_id,
        "flows": flows,
        "total_flows": total_flows,
        "tool_used": "zeek",
        "generated_at": datetime.now(timezone.utc).isoformat(),
    }

    # Validate output against schema
    validate_output("net.flow_summary", output)

    return output
