"""Compat runner: patches mcp>=1.0.0 API gaps, then runs find_evil_mcp.server.

Patches applied:
  1. Server.stdio() — removed in mcp 1.x; restored via stdio_server() context manager.
  2. call_tool single-TextContent — older versions iterate pydantic model as tuples;
     newer versions enforce structuredContent when outputSchema is set.
     Fix: wrap single ContentBlock returns in CallToolResult directly, which skips
     both issues (CallToolResult branch exits before outputSchema validation).
"""
import asyncio
from contextlib import asynccontextmanager


def _patch():
    from mcp.server.lowlevel import server as _lowlevel
    from mcp import types

    if getattr(_lowlevel.Server, '_warvis_patched', False):
        return

    try:
        from mcp.server import stdio_server
    except ImportError:
        from mcp.server.stdio import stdio_server

    # Patch 1: Server.stdio() context manager + wait_closed()
    @asynccontextmanager
    async def _stdio(self):
        async with stdio_server() as (read_stream, write_stream):
            self._stdio_streams = (read_stream, write_stream)
            yield

    async def _wait_closed(self):
        r, w = self._stdio_streams
        await self.run(r, w, self.create_initialization_options())

    _lowlevel.Server.stdio = _stdio
    _lowlevel.Server.wait_closed = _wait_closed

    # Patch 2: wrap call_tool so single TextContent/ImageContent bypasses
    # outputSchema validation (CallToolResult branch exits before that check).
    _orig_call_tool = _lowlevel.Server.call_tool

    def _patched_call_tool(self, **kwargs):
        orig_decorator = _orig_call_tool(self, **kwargs)

        def new_decorator(func):
            async def wrapped(name, arguments):
                result = await func(name, arguments)
                if isinstance(result, (types.TextContent, types.ImageContent)):
                    return types.CallToolResult(content=[result], isError=False)
                return result
            return orig_decorator(wrapped)

        return new_decorator

    _lowlevel.Server.call_tool = _patched_call_tool
    _lowlevel.Server._warvis_patched = True


_patch()

from find_evil_mcp.server import main  # noqa: E402
asyncio.run(main())
