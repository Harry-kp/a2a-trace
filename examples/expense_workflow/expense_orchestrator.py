#!/usr/bin/env python3
"""
Expense Workflow Orchestrator - Coordinates the expense reimbursement process.

Part of the Expense Reimbursement Workflow demo for a2a-trace.
Demonstrates multi-agent coordination with real-world business logic.
"""

import json
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime
import urllib.request
import urllib.error
import os

PORT = 8004

# Agent endpoints (will use proxy if HTTP_PROXY is set)
RECEIPT_AGENT = "http://localhost:8001"
POLICY_AGENT = "http://localhost:8002"
APPROVAL_AGENT = "http://localhost:8003"

AGENT_CARD = {
    "name": "Expense Orchestrator",
    "description": "Coordinates the complete expense reimbursement workflow across multiple agents",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "process_expense",
            "name": "Process Expense Report",
            "description": "Processes a complete expense report through the workflow",
            "examples": ["Process expense report with receipts rcpt-001, rcpt-002"]
        },
        {
            "id": "check_status",
            "name": "Check Expense Status",
            "description": "Checks the status of an expense report",
            "examples": ["What's the status of expense EXP-001?"]
        }
    ],
    "dependencies": [
        {"url": RECEIPT_AGENT, "name": "Receipt Analyzer"},
        {"url": POLICY_AGENT, "name": "Policy Checker"},
        {"url": APPROVAL_AGENT, "name": "Approval Workflow"}
    ]
}


def call_agent(url, method, params, request_id):
    """Make a JSON-RPC call to another agent."""
    proxy_handler = urllib.request.ProxyHandler()
    
    # Use proxy if configured
    http_proxy = os.environ.get("HTTP_PROXY") or os.environ.get("http_proxy")
    if http_proxy:
        proxy_handler = urllib.request.ProxyHandler({
            "http": http_proxy,
            "https": http_proxy
        })
    
    opener = urllib.request.build_opener(proxy_handler)
    
    payload = json.dumps({
        "jsonrpc": "2.0",
        "method": method,
        "params": params,
        "id": request_id
    }).encode()
    
    req = urllib.request.Request(
        url,
        data=payload,
        headers={"Content-Type": "application/json"}
    )
    
    try:
        with opener.open(req, timeout=10) as response:
            return json.loads(response.read().decode())
    except urllib.error.URLError as e:
        return {"error": {"code": -32000, "message": str(e)}}


class ExpenseOrchestratorHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[Orchestrator] {args[0]}")

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

        if method == "tasks/create":
            response = self.handle_task(params, request_id)
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

    def handle_task(self, params, request_id):
        skill = params.get("skill", "process_expense")
        
        if skill == "process_expense":
            return self.process_expense_report(params, request_id)
        elif skill == "check_status":
            approval_id = params.get("approval_id", "")
            return call_agent(
                APPROVAL_AGENT, 
                "tasks/create", 
                {"skill": "check_status", "approval_id": approval_id},
                f"{request_id}-status"
            )
        
        return {"error": {"code": -32602, "message": f"Unknown skill: {skill}"}}

    def process_expense_report(self, params, request_id):
        receipt_ids = params.get("receipt_ids", ["rcpt-001"])
        submitter = params.get("submitter", "John Smith")
        description = params.get("description", "Business trip expenses")
        
        workflow_log = []
        total_amount = 0
        all_expenses = []
        all_policy_results = []
        requires_approval = False
        has_violations = False
        
        print(f"\n{'='*60}")
        print(f"Processing expense report from {submitter}")
        print(f"Receipts: {receipt_ids}")
        print(f"{'='*60}\n")
        
        # Step 1: Analyze each receipt
        workflow_log.append({
            "step": 1,
            "action": "Analyzing receipts",
            "timestamp": datetime.now().isoformat()
        })
        
        for idx, receipt_id in enumerate(receipt_ids):
            print(f"[Step 1.{idx+1}] Analyzing receipt: {receipt_id}")
            
            receipt_result = call_agent(
                RECEIPT_AGENT,
                "tasks/create",
                {"skill": "analyze_receipt", "receipt_id": receipt_id},
                f"{request_id}-receipt-{idx}"
            )
            
            if "result" in receipt_result:
                expense_data = receipt_result["result"].get("extracted_data", {})
                all_expenses.append({
                    "receipt_id": receipt_id,
                    "data": expense_data
                })
                total_amount += expense_data.get("amount", 0)
                print(f"    ✓ Extracted: {expense_data.get('vendor')} - ${expense_data.get('amount', 0):.2f}")
            else:
                print(f"    ✗ Failed to analyze receipt: {receipt_result}")
        
        # Step 2: Check policy compliance for each expense
        workflow_log.append({
            "step": 2,
            "action": "Checking policy compliance",
            "timestamp": datetime.now().isoformat()
        })
        
        for idx, expense in enumerate(all_expenses):
            print(f"[Step 2.{idx+1}] Checking policy for: {expense['data'].get('vendor')}")
            
            policy_result = call_agent(
                POLICY_AGENT,
                "tasks/create",
                {"skill": "check_policy", "expense": expense["data"]},
                f"{request_id}-policy-{idx}"
            )
            
            if "result" in policy_result:
                result = policy_result["result"]
                all_policy_results.append({
                    "receipt_id": expense["receipt_id"],
                    "result": result
                })
                
                if result.get("requires_approval"):
                    requires_approval = True
                    print(f"    ⚠ Requires approval")
                
                if not result.get("compliant"):
                    has_violations = True
                    print(f"    ✗ Policy violation: {result.get('violations')}")
                else:
                    print(f"    ✓ Compliant")
        
        # Step 3: Submit for approval if needed
        approval_result = None
        if has_violations:
            workflow_log.append({
                "step": 3,
                "action": "Expense report has violations - requires review",
                "timestamp": datetime.now().isoformat()
            })
            print(f"[Step 3] Expense report flagged for review due to violations")
        else:
            workflow_log.append({
                "step": 3,
                "action": "Submitting for approval",
                "timestamp": datetime.now().isoformat()
            })
            print(f"[Step 3] Submitting for approval (total: ${total_amount:.2f})")
            
            approval_result = call_agent(
                APPROVAL_AGENT,
                "tasks/create",
                {
                    "skill": "submit_for_approval",
                    "expense_id": f"EXP-{request_id[:8]}",
                    "amount": total_amount,
                    "submitter": submitter,
                    "category": "Mixed",
                    "description": description
                },
                f"{request_id}-approval"
            )
            
            if "result" in approval_result:
                result = approval_result["result"]
                print(f"    ✓ Submitted: {result.get('approval_id')} - {result.get('status')}")
        
        # Compile final result
        return {
            "result": {
                "status": "submitted" if not has_violations else "needs_review",
                "expense_id": f"EXP-{request_id[:8]}",
                "submitter": submitter,
                "total_amount": total_amount,
                "receipts_processed": len(all_expenses),
                "policy_compliant": not has_violations,
                "requires_approval": requires_approval,
                "approval": approval_result.get("result") if approval_result else None,
                "expenses": [
                    {
                        "receipt_id": e["receipt_id"],
                        "vendor": e["data"].get("vendor"),
                        "amount": e["data"].get("amount"),
                        "category": e["data"].get("category")
                    }
                    for e in all_expenses
                ],
                "policy_results": all_policy_results,
                "workflow_log": workflow_log,
                "processed_at": datetime.now().isoformat()
            }
        }


if __name__ == "__main__":
    server = HTTPServer(("", PORT), ExpenseOrchestratorHandler)
    print(f"🎯 Expense Orchestrator running on port {PORT}")
    print(f"   Agent Card: http://localhost:{PORT}/.well-known/agent.json")
    print(f"\n   Dependencies:")
    print(f"   - Receipt Analyzer: {RECEIPT_AGENT}")
    print(f"   - Policy Checker:   {POLICY_AGENT}")
    print(f"   - Approval Agent:   {APPROVAL_AGENT}")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")

