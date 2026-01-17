#!/usr/bin/env python3
"""
Receipt Analyzer Agent - Extracts expense data from receipts.

Part of the Expense Reimbursement Workflow demo for a2a-trace.
Simulates OCR and data extraction from expense receipts.
"""

import json
import random
import time
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse

PORT = 8001

# Simulated receipt database
SAMPLE_RECEIPTS = {
    "rcpt-001": {
        "vendor": "Hilton Hotels",
        "category": "Lodging",
        "amount": 289.00,
        "currency": "USD",
        "date": "2026-01-15",
        "location": "San Francisco, CA",
        "tax": 28.90,
        "confidence": 0.98
    },
    "rcpt-002": {
        "vendor": "United Airlines",
        "category": "Transportation",
        "amount": 542.50,
        "currency": "USD",
        "date": "2026-01-14",
        "location": "SFO → JFK",
        "tax": 45.00,
        "confirmation": "UA789456",
        "confidence": 0.99
    },
    "rcpt-003": {
        "vendor": "The French Laundry",
        "category": "Meals",
        "amount": 385.00,
        "currency": "USD",
        "date": "2026-01-15",
        "location": "Yountville, CA",
        "tax": 32.50,
        "attendees": ["John Smith", "Client: Sarah Chen"],
        "confidence": 0.95
    },
    "rcpt-004": {
        "vendor": "Uber",
        "category": "Transportation",
        "amount": 45.80,
        "currency": "USD",
        "date": "2026-01-14",
        "location": "SFO to Downtown SF",
        "confidence": 0.97
    }
}

AGENT_CARD = {
    "name": "Receipt Analyzer",
    "description": "Extracts and validates expense data from receipt images using OCR and ML",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "analyze_receipt",
            "name": "Analyze Receipt",
            "description": "Extracts vendor, amount, date, category from receipt image",
            "examples": ["Analyze receipt rcpt-001", "Extract data from this receipt"]
        },
        {
            "id": "validate_receipt",
            "name": "Validate Receipt",
            "description": "Validates receipt data for completeness and accuracy",
            "examples": ["Validate this receipt data"]
        }
    ]
}


class ReceiptAgentHandler(BaseHTTPRequestHandler):
    def log_message(self, format, *args):
        print(f"[Receipt Agent] {args[0]}")

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

        # Simulate processing time
        time.sleep(random.uniform(0.1, 0.3))

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
        skill = params.get("skill", "analyze_receipt")
        receipt_id = params.get("receipt_id", "rcpt-001")
        
        if skill == "analyze_receipt":
            if receipt_id in SAMPLE_RECEIPTS:
                receipt = SAMPLE_RECEIPTS[receipt_id]
                return {
                    "result": {
                        "status": "success",
                        "receipt_id": receipt_id,
                        "extracted_data": receipt,
                        "processing_time_ms": random.randint(150, 400),
                        "ocr_confidence": receipt.get("confidence", 0.95)
                    }
                }
            else:
                # Generate random receipt for unknown IDs
                return {
                    "result": {
                        "status": "success",
                        "receipt_id": receipt_id,
                        "extracted_data": {
                            "vendor": "Unknown Vendor",
                            "category": "Miscellaneous",
                            "amount": round(random.uniform(10, 200), 2),
                            "currency": "USD",
                            "date": "2026-01-16",
                            "confidence": round(random.uniform(0.7, 0.9), 2)
                        },
                        "processing_time_ms": random.randint(200, 500)
                    }
                }
        
        elif skill == "validate_receipt":
            data = params.get("data", {})
            issues = []
            
            if not data.get("vendor"):
                issues.append("Missing vendor name")
            if not data.get("amount") or data.get("amount", 0) <= 0:
                issues.append("Invalid or missing amount")
            if not data.get("date"):
                issues.append("Missing date")
            
            return {
                "result": {
                    "status": "valid" if not issues else "invalid",
                    "issues": issues,
                    "validated_at": "2026-01-17T10:00:00Z"
                }
            }
        
        return {"error": {"code": -32602, "message": f"Unknown skill: {skill}"}}


if __name__ == "__main__":
    server = HTTPServer(("", PORT), ReceiptAgentHandler)
    print(f"🧾 Receipt Analyzer Agent running on port {PORT}")
    print(f"   Agent Card: http://localhost:{PORT}/.well-known/agent.json")
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nShutting down...")

