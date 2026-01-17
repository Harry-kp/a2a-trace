#!/usr/bin/env python3
"""
Policy Checker Agent - Validates expenses against company policies.

Part of the Expense Reimbursement Workflow demo for a2a-trace.
Enforces expense policies like daily limits, approved categories, etc.
"""

import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

PORT = 8002

# Company expense policies
POLICIES = {
    "Lodging": {
        "daily_limit": 300.00,
        "requires_approval_above": 250.00,
        "allowed_vendors": None,  # Any vendor
        "documentation_required": True
    },
    "Transportation": {
        "daily_limit": 600.00,
        "requires_approval_above": 500.00,
        "allowed_vendors": ["United Airlines", "Delta", "American", "Uber", "Lyft"],
        "documentation_required": True
    },
    "Meals": {
        "daily_limit": 75.00,
        "requires_approval_above": 50.00,
        "per_person_limit": 30.00,
        "requires_client_name": True,
        "documentation_required": True
    },
    "Miscellaneous": {
        "daily_limit": 100.00,
        "requires_approval_above": 50.00,
        "documentation_required": True
    }
}

AGENT_CARD = {
    "name": "Policy Checker",
    "description": "Validates expenses against company reimbursement policies and limits",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "check_policy",
            "name": "Check Policy Compliance",
            "description": "Validates an expense against company policies",
            "examples": ["Check if this expense complies with policy"]
        },
        {
            "id": "get_limits",
            "name": "Get Category Limits",
            "description": "Returns the spending limits for a category",
            "examples": ["What are the limits for Meals?"]
        }
    ]
}


class PolicyAgentHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[Policy Agent] {args[0]}")

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

        # Simulate policy lookup time
        time.sleep(0.05)

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
        skill = params.get("skill", "check_policy")
        
        if skill == "check_policy":
            expense = params.get("expense", {})
            return self.check_policy_compliance(expense)
        
        elif skill == "get_limits":
            category = params.get("category", "Miscellaneous")
            policy = POLICIES.get(category, POLICIES["Miscellaneous"])
            return {
                "result": {
                    "category": category,
                    "limits": policy
                }
            }
        
        return {"error": {"code": -32602, "message": f"Unknown skill: {skill}"}}

    def check_policy_compliance(self, expense):
        category = expense.get("category", "Miscellaneous")
        amount = expense.get("amount", 0)
        vendor = expense.get("vendor", "")
        attendees = expense.get("attendees", [])
        
        policy = POLICIES.get(category, POLICIES["Miscellaneous"])
        violations = []
        warnings = []
        requires_approval = False
        
        # Check daily limit
        if amount > policy["daily_limit"]:
            violations.append({
                "type": "OVER_LIMIT",
                "message": f"Amount ${amount:.2f} exceeds daily limit of ${policy['daily_limit']:.2f}",
                "severity": "error"
            })
        
        # Check approval threshold
        if amount > policy.get("requires_approval_above", float("inf")):
            requires_approval = True
            warnings.append({
                "type": "REQUIRES_APPROVAL",
                "message": f"Amount ${amount:.2f} requires manager approval (threshold: ${policy['requires_approval_above']:.2f})",
                "severity": "warning"
            })
        
        # Check allowed vendors
        allowed = policy.get("allowed_vendors")
        if allowed and vendor not in allowed:
            warnings.append({
                "type": "UNAPPROVED_VENDOR",
                "message": f"Vendor '{vendor}' is not on the pre-approved list",
                "severity": "warning"
            })
        
        # Special rules for meals
        if category == "Meals":
            per_person = policy.get("per_person_limit", float("inf"))
            num_people = max(len(attendees), 1)
            per_person_actual = amount / num_people
            
            if per_person_actual > per_person:
                warnings.append({
                    "type": "PER_PERSON_EXCEEDED",
                    "message": f"Per-person cost ${per_person_actual:.2f} exceeds limit of ${per_person:.2f}",
                    "severity": "warning"
                })
            
            if policy.get("requires_client_name") and not any("client" in a.lower() for a in attendees):
                violations.append({
                    "type": "MISSING_CLIENT",
                    "message": "Client meals require client name in attendee list",
                    "severity": "error"
                })
        
        is_compliant = len(violations) == 0
        
        return {
            "result": {
                "compliant": is_compliant,
                "requires_approval": requires_approval,
                "violations": violations,
                "warnings": warnings,
                "checked_at": datetime.now().isoformat(),
                "policy_version": "2026.1"
            }
        }


if __name__ == "__main__":
    server = HTTPServer(("", PORT), PolicyAgentHandler)
    print(f"📋 Policy Checker Agent running on port {PORT}")
    print(f"   Agent Card: http://localhost:{PORT}/.well-known/agent.json")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")

