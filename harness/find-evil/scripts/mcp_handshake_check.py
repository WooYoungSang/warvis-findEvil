# SPDX-License-Identifier: MIT
"""ITEM-212 Phase 1-2 — capability handshake conformance check.

Verifies the find-evil-mcp server exposes:
  - server_info.name == "find-evil-mcp"
  - server_info.version == "0.1.0"
  - Exactly 9 tools (3 Phase-1 + 6 Phase-2)
  - No shell/exec/eval/read_file/write_file tools (architectural constraint).

Exit 0 = PASS; non-zero = BLOCK.
"""
from __future__ import annotations

import asyncio
import sys

EXPECTED_TOOLS = {
    "case.open",
    "timeline.build",
    "iocs.scan",
    "memory.process_list",
    "memory.malfind",
    "net.flow_summary",
    "log.query",
    "report.append",
    "verify.cross_check",
}
FORBIDDEN_TOOLS = {"shell", "exec", "eval", "read_file", "write_file"}
EXPECTED_NAME = "find-evil-mcp"
EXPECTED_VERSION = "0.1.0"


async def _check() -> int:
    from find_evil_mcp import __version__
    from find_evil_mcp.server import create_server
    from mcp.types import ListToolsRequest

    server = create_server()
    name = server.name
    version = __version__

    handler = server.request_handlers.get(ListToolsRequest)
    if handler is None:
        print("FAIL: server has no ListToolsRequest handler", file=sys.stderr)
        return 3
    result = await handler(ListToolsRequest(method="tools/list"))
    tools = result.root.tools if hasattr(result, "root") else result.tools

    tool_names = {t.name for t in tools}

    errors: list[str] = []
    if name != EXPECTED_NAME:
        errors.append(f"server.name={name!r}, expected {EXPECTED_NAME!r}")
    if str(version) != EXPECTED_VERSION:
        errors.append(f"server.version={version!r}, expected {EXPECTED_VERSION!r}")
    missing = EXPECTED_TOOLS - tool_names
    extra = tool_names - EXPECTED_TOOLS
    if missing:
        errors.append(f"missing tools: {sorted(missing)}")
    if extra:
        errors.append(f"unexpected tools (Phase 2 leakage?): {sorted(extra)}")
    forbidden = FORBIDDEN_TOOLS & tool_names
    if forbidden:
        errors.append(
            f"forbidden tools exposed (architectural constraint violation): {sorted(forbidden)}"
        )

    if errors:
        for e in errors:
            print(f"FAIL: {e}", file=sys.stderr)
        return 1

    tools_str = str(sorted(tool_names))
    print(f"mcp-conformance: PASS — {EXPECTED_NAME} v{EXPECTED_VERSION} with tools {tools_str}")
    return 0


if __name__ == "__main__":
    sys.exit(asyncio.run(_check()))
