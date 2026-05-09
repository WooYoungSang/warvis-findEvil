#!/usr/bin/env python3
"""Mock Ollama server for kill-switch testing.

Returns deterministic JSON actions so KS-2/3 (tool_called, gemma_response events)
pass without needing a real GPU or network.
"""
import json
import sys
from http.server import HTTPServer, BaseHTTPRequestHandler

_call_count = 0

# Cycle of actions returned per chat request.
# index 0 and 1 → call_tool in TRACE (satisfies KS-3: ≥2 tool_called events)
# index 2 → state_complete (advances TRACE → SCAN with the new agent loop)
# index 3 and 4 → call_tool in SCAN (memory.* — exercises vol3 via sift_runner;
#                supports the comprehensive-mock-hunt L3/L4 lock-in for the
#                prompt-context-threading Bet without disturbing kill-switch).
# index 5 → state_complete (SCAN → EXPOSE)
# index 5+ → state_complete repeats (further transitions through EXPOSE → LOCK)
# Kill-switch tests use WARVIS_MAX_TURNS=3 and never reach index 3+.
ACTIONS = [
    {"action": "call_tool", "tool_name": "timeline.build", "arguments": {"limit": 100},
     "reason": "Building forensic timeline to identify suspicious activities"},
    {"action": "call_tool", "tool_name": "log.query", "arguments": {"limit": 50},
     "reason": "Querying event logs for anomalous entries"},
    {"action": "state_complete",
     "reason": "Timeline and log analysis complete; advancing to SCAN for memory analysis"},
    {"action": "call_tool", "tool_name": "memory.process_list",
     "arguments": {"limit": 200},
     "reason": "Enumerating processes via Volatility3 to spot anomalies"},
    {"action": "call_tool", "tool_name": "memory.malfind", "arguments": {},
     "reason": "Hunting for code-injected memory regions via Volatility3 malfind"},
    {"action": "state_complete",
     "reason": "Memory enumeration complete; advancing to EXPOSE for verification"},
]


class MockOllamaHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        global _call_count
        if self.path == "/api/chat":
            content_len = int(self.headers.get("Content-Length", 0))
            _ = self.rfile.read(content_len)

            idx = min(_call_count, len(ACTIONS) - 1)
            action_content = json.dumps(ACTIONS[idx])
            _call_count += 1

            body = json.dumps({
                "model": "gemma4:26b-a4b-it-q4_K_M",
                "created_at": "2026-05-05T00:00:00Z",
                "message": {"role": "assistant", "content": action_content},
                "done": True,
                "done_reason": "stop",
            }).encode()

            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.send_header("Content-Length", str(len(body)))
            self.end_headers()
            self.wfile.write(body)
        else:
            self.send_response(404)
            self.end_headers()

    def log_message(self, fmt, *args):  # suppress access log noise
        pass


def main():
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 19134
    server = HTTPServer(("127.0.0.1", port), MockOllamaHandler)
    server.serve_forever()


if __name__ == "__main__":
    main()
