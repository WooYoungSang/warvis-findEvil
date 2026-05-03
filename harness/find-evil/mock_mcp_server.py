#!/usr/bin/env python3
"""Mock MCP server for kill-switch testing."""
import json
import sys
import uuid

def send_json(obj):
    sys.stdout.write(json.dumps(obj) + '\n')
    sys.stdout.flush()

def read_json():
    try:
        line = sys.stdin.readline()
        if not line:
            return None
        return json.loads(line)
    except:
        return None

# Main loop
while True:
    msg = read_json()
    if not msg:
        break
    
    msg_id = msg.get('id')
    method = msg.get('method')
    
    if method == 'initialize':
        send_json({
            'jsonrpc': '2.0',
            'id': msg_id,
            'result': {
                'protocolVersion': '2024-11-05',
                'capabilities': {},
                'serverInfo': {
                    'name': 'find-evil-mcp',
                    'version': '0.0.1'
                }
            }
        })
    elif method == 'tools/list':
        send_json({
            'jsonrpc': '2.0',
            'id': msg_id,
            'result': {
                'tools': [
                    {'name': 'case.open', 'description': 'Open case', 'inputSchema': {}},
                    {'name': 'timeline.build', 'description': 'Build timeline', 'inputSchema': {}},
                    {'name': 'log.query', 'description': 'Query logs', 'inputSchema': {}},
                    {'name': 'iocs.scan', 'description': 'Scan IOCs', 'inputSchema': {}},
                    {'name': 'memory.process_list', 'description': 'List processes', 'inputSchema': {}},
                    {'name': 'memory.malfind', 'description': 'Find malware', 'inputSchema': {}},
                    {'name': 'net.flow_summary', 'description': 'Network flows', 'inputSchema': {}},
                    {'name': 'verify.cross_check', 'description': 'Verify', 'inputSchema': {}},
                    {'name': 'report.append', 'description': 'Append report', 'inputSchema': {}},
                ]
            }
        })
    elif method == 'tools/call':
        tool_name = msg.get('params', {}).get('name')
        if tool_name == 'case.open':
            case_id = str(uuid.uuid4())
            send_json({
                'jsonrpc': '2.0',
                'id': msg_id,
                'result': {
                    'content': [
                        {
                            'type': 'text',
                            'text': json.dumps({
                                'case_id': case_id,
                                'sandbox_root': f'/tmp/sandbox/{case_id}',
                                'evidence_kind': 'disk_image'
                            })
                        }
                    ],
                    'isError': False
                }
            })
        else:
            send_json({
                'jsonrpc': '2.0',
                'id': msg_id,
                'result': {
                    'content': [{'type': 'text', 'text': '{}'}],
                    'isError': False
                }
            })
    elif method == 'notifications/initialized':
        # Notification, no response needed
        pass
    else:
        # Unknown method
        if msg_id:
            send_json({
                'jsonrpc': '2.0',
                'id': msg_id,
                'error': {
                    'code': -32601,
                    'message': f'Method not found: {method}'
                }
            })
