# SPDX-License-Identifier: MIT

import json
import asyncio
from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

from find_evil_mcp.tools.case_open import handle_case_open
from find_evil_mcp.tools.timeline_build import handle_timeline_build
from find_evil_mcp.tools.iocs_scan import handle_iocs_scan
from find_evil_mcp.tools.memory_process_list import handle_memory_process_list
from find_evil_mcp.tools.memory_malfind import handle_memory_malfind
from find_evil_mcp.tools.net_flow_summary import handle_net_flow_summary
from find_evil_mcp.tools.log_query import handle_log_query
from find_evil_mcp.tools.report_append import handle_report_append
from find_evil_mcp.tools.verify_cross_check import handle_verify_cross_check
from find_evil_mcp.schema_loader import get_output_schema


_UUID_RE = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"


def create_server() -> Server:
    """Create and configure the MCP server."""
    server = Server("find-evil-mcp")

    @server.list_tools()
    async def list_tools():
        """List available tools."""
        return [
            Tool(
                name="case.open",
                description="Register and sandbox forensic evidence source",
                inputSchema={
                    "type": "object",
                    "oneOf": [
                        {
                            "type": "object",
                            "properties": {
                                "image_path": {
                                    "type": "string",
                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
                                }
                            },
                            "required": ["image_path"],
                            "additionalProperties": False,
                        },
                        {
                            "type": "object",
                            "properties": {
                                "mem_path": {
                                    "type": "string",
                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
                                }
                            },
                            "required": ["mem_path"],
                            "additionalProperties": False,
                        },
                        {
                            "type": "object",
                            "properties": {
                                "pcap_path": {
                                    "type": "string",
                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
                                }
                            },
                            "required": ["pcap_path"],
                            "additionalProperties": False,
                        },
                        {
                            "type": "object",
                            "properties": {
                                "logdir_path": {
                                    "type": "string",
                                    "pattern": "^/evidence/[a-zA-Z0-9._/-]+$",
                                }
                            },
                            "required": ["logdir_path"],
                            "additionalProperties": False,
                        },
                    ],
                },
                outputSchema=get_output_schema("case.open"),
            ),
            Tool(
                name="timeline.build",
                description="Construct normalized forensic timeline from case evidence",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "source_filter": {
                            "type": "string",
                            "enum": ["disk", "memory", "logs", "network", "all"],
                            "default": "all",
                        },
                        "since": {"type": "string", "format": "date-time"},
                        "until": {"type": "string", "format": "date-time"},
                    },
                    "required": ["case_id"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("timeline.build"),
            ),
            Tool(
                name="iocs.scan",
                description="Scan case evidence against YARA/Sigma IOC rules",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "ruleset": {
                            "type": "string",
                            "enum": ["yara_default", "yara_custom", "sigma"],
                        },
                        "target_glob": {
                            "type": "string",
                            "pattern": "^[a-zA-Z0-9._/*?-]+$",
                        },
                    },
                    "required": ["case_id", "ruleset"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("iocs.scan"),
            ),
            Tool(
                name="memory.process_list",
                description="Extract process list from memory dump (Volatility3)",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "kdbg_offset": {
                            "type": "string",
                            "pattern": "^0x[0-9a-fA-F]+$",
                        },
                    },
                    "required": ["case_id"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("memory.process_list"),
            ),
            Tool(
                name="memory.malfind",
                description="Detect injected/suspicious memory regions (Volatility3)",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "pid": {
                            "type": "integer",
                            "minimum": 0,
                        },
                    },
                    "required": ["case_id"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("memory.malfind"),
            ),
            Tool(
                name="net.flow_summary",
                description="Extract and summarize network flows from PCAP (Zeek/Suricata)",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "pcap_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "top_n": {
                            "type": "integer",
                            "minimum": 1,
                            "maximum": 1000,
                            "default": 50,
                        },
                    },
                    "required": ["case_id", "pcap_id"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("net.flow_summary"),
            ),
            Tool(
                name="log.query",
                description="Query forensic timeline and logs with structured filters",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "q": {
                            "type": "string",
                            "maxLength": 512,
                            "pattern": "^[a-zA-Z0-9_.:= /*\"'-]+$",
                        },
                        "source": {
                            "type": "string",
                            "enum": ["evtx", "syslog", "audit", "all"],
                            "default": "all",
                        },
                        "since": {"type": "string", "format": "date-time"},
                        "until": {"type": "string", "format": "date-time"},
                        "limit": {
                            "type": "integer",
                            "minimum": 1,
                            "maximum": 10000,
                            "default": 1000,
                        },
                    },
                    "required": ["case_id", "q"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("log.query"),
            ),
            Tool(
                name="report.append",
                description="Append structured finding to case report with hash chain integrity",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "finding": {
                            "type": "object",
                            "properties": {
                                "title": {"type": "string"},
                                "severity": {
                                    "type": "string",
                                    "enum": ["info", "low", "medium", "high", "critical"],
                                },
                                "narrative": {"type": "string"},
                                "evidence": {"type": "array"},
                                "mitre_attack": {"type": "array"},
                            },
                            "required": ["title", "severity", "narrative", "evidence"],
                        },
                    },
                    "required": ["case_id", "finding"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("report.append"),
            ),
            Tool(
                name="verify.cross_check",
                description="Re-verify findings against multiple sources for confidence scoring",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "case_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "finding_id": {
                            "type": "string",
                            "pattern": _UUID_RE,
                        },
                        "method": {
                            "type": "string",
                            "enum": ["rerun", "alt_tool", "counter_evidence", "all"],
                            "default": "all",
                        },
                    },
                    "required": ["case_id", "finding_id"],
                    "additionalProperties": False,
                },
                outputSchema=get_output_schema("verify.cross_check"),
            ),
        ]

    @server.call_tool()
    async def call_tool(name: str, arguments: dict):
        """Call a tool by name.

        mcp >= 1.0 with outputSchema requires returning a (content, structured)
        tuple so the framework can validate the structured output against the
        per-tool outputSchema declared in list_tools(). Returning only
        TextContent fails with "Output validation error: outputSchema defined
        but no structured output returned".
        """
        try:
            if name == "case.open":
                result = await handle_case_open(arguments)
            elif name == "timeline.build":
                result = await handle_timeline_build(arguments)
            elif name == "iocs.scan":
                result = await handle_iocs_scan(arguments)
            elif name == "memory.process_list":
                result = await handle_memory_process_list(arguments)
            elif name == "memory.malfind":
                result = await handle_memory_malfind(arguments)
            elif name == "net.flow_summary":
                result = await handle_net_flow_summary(arguments)
            elif name == "log.query":
                result = await handle_log_query(arguments)
            elif name == "report.append":
                result = await handle_report_append(arguments)
            elif name == "verify.cross_check":
                result = await handle_verify_cross_check(arguments)
            else:
                return [TextContent(type="text", text=f"Unknown tool: {name}")]

            content = [TextContent(type="text", text=json.dumps(result))]
            return content, result

        except Exception as e:
            return [TextContent(
                type="text",
                text=f"Error calling tool {name}: {str(e)}",
            )]

    return server


async def main():
    """Main entry point for stdio transport (mcp >= 1.0 API)."""
    server = create_server()
    async with stdio_server() as (read_stream, write_stream):
        await server.run(
            read_stream,
            write_stream,
            server.create_initialization_options(),
        )


if __name__ == "__main__":
    asyncio.run(main())
