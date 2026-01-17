#!/usr/bin/env python3
"""
Echo Agent - A simple A2A agent that echoes back messages.
Demonstrates basic A2A protocol implementation.
"""

import json
import uuid
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

PORT = 8001

AGENT_CARD = {
    "name": "Echo Agent",
    "description": "A simple agent that echoes back any message it receives",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "echo",
            "name": "Echo Message",
            "description": "Echoes back any message sent to it",
            "examples": ["Echo this message back to me"]
        }
    ]
}

# In-memory task storage
tasks = {}


class A2AHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[{datetime.now().strftime('%H:%M:%S')}] {args[0]}")

    def send_json(self, data, status=200):
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode())

    def do_GET(self):
        if self.path == "/.well-known/agent.json":
            self.send_json(AGENT_CARD)
        elif self.path == "/health":
            self.send_json({"status": "ok"})
        else:
            self.send_json({"error": "Not found"}, 404)

    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length).decode()

        try:
            request = json.loads(body)
        except json.JSONDecodeError:
            self.send_json({
                "jsonrpc": "2.0",
                "error": {"code": -32700, "message": "Parse error"},
                "id": None
            }, 400)
            return

        method = request.get("method", "")
        request_id = request.get("id")
        params = request.get("params", {})

        print(f"  Method: {method}")
        print(f"  Params: {json.dumps(params)[:100]}...")

        if method == "tasks/create":
            self.handle_create_task(request_id, params)
        elif method == "tasks/get":
            self.handle_get_task(request_id, params)
        elif method == "tasks/cancel":
            self.handle_cancel_task(request_id, params)
        else:
            self.send_json({
                "jsonrpc": "2.0",
                "error": {"code": -32601, "message": f"Method not found: {method}"},
                "id": request_id
            })

    def handle_create_task(self, request_id, params):
        task_id = str(uuid.uuid4())
        message = params.get("message", params.get("input", ""))

        # Create the task
        task = {
            "id": task_id,
            "status": "completed",
            "created_at": datetime.now().isoformat(),
            "result": {
                "echo": message,
                "received_at": datetime.now().isoformat(),
                "agent": "Echo Agent"
            }
        }
        tasks[task_id] = task

        print(f"  Created task: {task_id}")
        print(f"  Echo: {message[:50]}...")

        self.send_json({
            "jsonrpc": "2.0",
            "result": task,
            "id": request_id
        })

    def handle_get_task(self, request_id, params):
        task_id = params.get("task_id", params.get("id", ""))
        
        if task_id in tasks:
            self.send_json({
                "jsonrpc": "2.0",
                "result": tasks[task_id],
                "id": request_id
            })
        else:
            self.send_json({
                "jsonrpc": "2.0",
                "error": {"code": -32000, "message": f"Task not found: {task_id}"},
                "id": request_id
            })

    def handle_cancel_task(self, request_id, params):
        task_id = params.get("task_id", params.get("id", ""))
        
        if task_id in tasks:
            tasks[task_id]["status"] = "cancelled"
            self.send_json({
                "jsonrpc": "2.0",
                "result": {"cancelled": True},
                "id": request_id
            })
        else:
            self.send_json({
                "jsonrpc": "2.0",
                "error": {"code": -32000, "message": f"Task not found: {task_id}"},
                "id": request_id
            })

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()


def main():
    server = HTTPServer(("", PORT), A2AHandler)
    print(f"🤖 Echo Agent starting on port {PORT}")
    print(f"   Agent card: http://localhost:{PORT}/.well-known/agent.json")
    print(f"   Health: http://localhost:{PORT}/health")
    print()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n👋 Echo Agent shutting down")
        server.shutdown()


if __name__ == "__main__":
    main()

