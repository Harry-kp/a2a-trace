#!/usr/bin/env python3
"""
Approval Agent - Handles expense approval workflow.

Part of the Expense Reimbursement Workflow demo for a2a-trace.
Routes expenses to appropriate approvers and tracks approval status.
"""

import json
import time
import random
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime
import uuid

PORT = 8003

# Simulated approval queue
APPROVAL_QUEUE = {}

# Approval hierarchy
APPROVERS = {
    "level_1": {
        "threshold": 100.00,
        "approver": "Team Lead",
        "sla_hours": 24
    },
    "level_2": {
        "threshold": 500.00,
        "approver": "Department Manager",
        "sla_hours": 48
    },
    "level_3": {
        "threshold": float("inf"),
        "approver": "Finance Director",
        "sla_hours": 72
    }
}

AGENT_CARD = {
    "name": "Approval Workflow",
    "description": "Manages expense approval routing and tracking",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": True
    },
    "skills": [
        {
            "id": "submit_for_approval",
            "name": "Submit for Approval",
            "description": "Submits an expense report for approval",
            "examples": ["Submit expense report EXP-001 for approval"]
        },
        {
            "id": "check_status",
            "name": "Check Approval Status",
            "description": "Checks the current status of an approval request",
            "examples": ["What's the status of approval APR-001?"]
        },
        {
            "id": "approve",
            "name": "Approve Expense",
            "description": "Approves a pending expense (for simulation)",
            "examples": ["Approve expense APR-001"]
        }
    ]
}


class ApprovalAgentHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[Approval Agent] {args[0]}")

    def do_GET(self):
        if self.path == "/.well-known/agent.json":
            self.send_response(200)
            self.send_header("Content-Type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(AGENT_CARD).encode())
        else:
            self.send_error(404)

    def do_POST(self):
        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length)
        
        try:
            request = json.loads(body)
        except json.JSONDecodeError:
            self.send_error(400, "Invalid JSON")
            return

        method = request.get("method", "")
        params = request.get("params", {})
        request_id = request.get("id")

        # Simulate processing
        time.sleep(random.uniform(0.05, 0.15))

        if method == "tasks/create":
            response = self.handle_task(params)
        else:
            response = {"error": {"code": -32601, "message": f"Unknown method: {method}"}}

        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        
        result = {
            "jsonrpc": "2.0",
            "id": request_id,
            **response
        }
        self.wfile.write(json.dumps(result).encode())

    def handle_task(self, params):
        skill = params.get("skill", "submit_for_approval")
        
        if skill == "submit_for_approval":
            return self.submit_for_approval(params)
        elif skill == "check_status":
            return self.check_status(params)
        elif skill == "approve":
            return self.approve_expense(params)
        
        return {"error": {"code": -32602, "message": f"Unknown skill: {skill}"}}

    def submit_for_approval(self, params):
        expense_id = params.get("expense_id", str(uuid.uuid4())[:8])
        amount = params.get("amount", 0)
        submitter = params.get("submitter", "Employee")
        category = params.get("category", "Miscellaneous")
        description = params.get("description", "")
        
        # Determine approval level
        approval_level = "level_1"
        for level, config in sorted(APPROVERS.items()):
            if amount <= config["threshold"]:
                approval_level = level
                break
        
        config = APPROVERS[approval_level]
        approval_id = f"APR-{str(uuid.uuid4())[:8].upper()}"
        
        # Store in queue
        APPROVAL_QUEUE[approval_id] = {
            "id": approval_id,
            "expense_id": expense_id,
            "amount": amount,
            "category": category,
            "description": description,
            "submitter": submitter,
            "status": "pending",
            "level": approval_level,
            "approver": config["approver"],
            "sla_deadline": f"{config['sla_hours']} hours",
            "submitted_at": datetime.now().isoformat(),
            "history": [
                {
                    "action": "submitted",
                    "timestamp": datetime.now().isoformat(),
                    "actor": submitter
                }
            ]
        }
        
        # Simulate auto-approval for small amounts (demo purposes)
        if amount < 50 and random.random() > 0.3:
            APPROVAL_QUEUE[approval_id]["status"] = "auto_approved"
            APPROVAL_QUEUE[approval_id]["history"].append({
                "action": "auto_approved",
                "timestamp": datetime.now().isoformat(),
                "actor": "System",
                "reason": "Below auto-approval threshold"
            })
        
        return {
            "result": {
                "approval_id": approval_id,
                "status": APPROVAL_QUEUE[approval_id]["status"],
                "assigned_to": config["approver"],
                "sla_deadline": f"{config['sla_hours']} hours",
                "message": f"Expense submitted for {config['approver']} approval"
            }
        }

    def check_status(self, params):
        approval_id = params.get("approval_id", "")
        
        if approval_id in APPROVAL_QUEUE:
            approval = APPROVAL_QUEUE[approval_id]
            return {
                "result": {
                    "approval_id": approval_id,
                    "status": approval["status"],
                    "approver": approval["approver"],
                    "submitted_at": approval["submitted_at"],
                    "history": approval["history"]
                }
            }
        
        return {
            "result": {
                "approval_id": approval_id,
                "status": "not_found",
                "message": "Approval request not found"
            }
        }

    def approve_expense(self, params):
        approval_id = params.get("approval_id", "")
        approver = params.get("approver", "Manager")
        
        if approval_id in APPROVAL_QUEUE:
            approval = APPROVAL_QUEUE[approval_id]
            if approval["status"] == "pending":
                approval["status"] = "approved"
                approval["history"].append({
                    "action": "approved",
                    "timestamp": datetime.now().isoformat(),
                    "actor": approver
                })
                return {
                    "result": {
                        "approval_id": approval_id,
                        "status": "approved",
                        "approved_by": approver,
                        "approved_at": datetime.now().isoformat()
                    }
                }
            else:
                return {
                    "result": {
                        "approval_id": approval_id,
                        "status": approval["status"],
                        "message": f"Cannot approve - current status is {approval['status']}"
                    }
                }
        
        return {"error": {"code": -32602, "message": f"Approval not found: {approval_id}"}}


if __name__ == "__main__":
    server = HTTPServer(("", PORT), ApprovalAgentHandler)
    print(f"✅ Approval Workflow Agent running on port {PORT}")
    print(f"   Agent Card: http://localhost:{PORT}/.well-known/agent.json")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")

