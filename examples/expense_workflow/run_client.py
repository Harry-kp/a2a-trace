#!/usr/bin/env python3
"""
Expense Workflow Demo Client - Demonstrates the full expense reimbursement workflow.

This script submits expense reports and shows the multi-agent coordination.
Run this with a2a-trace to see all the inter-agent communication:

    a2a-trace -- python3 run_client.py
"""

import json
import time
import os
import urllib.request
import urllib.error

# Use proxy if available (set by a2a-trace)
def get_opener():
    proxy = os.environ.get("HTTP_PROXY") or os.environ.get("http_proxy")
    if proxy:
        print(f"Using proxy: {proxy}")
        return urllib.request.build_opener(
            urllib.request.ProxyHandler({"http": proxy, "https": proxy})
        )
    return urllib.request.build_opener()


def call_agent(url, method, params, request_id):
    """Make a JSON-RPC call to an agent."""
    opener = get_opener()
    
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
        with opener.open(req, timeout=30) as response:
            return json.loads(response.read().decode())
    except urllib.error.URLError as e:
        return {"error": {"code": -32000, "message": str(e)}}


def discover_agents():
    """Discover available agents via their agent cards."""
    print("\n" + "="*60)
    print("🔍 Discovering Agents")
    print("="*60 + "\n")
    
    agents = [
        ("Receipt Analyzer", "http://localhost:8001"),
        ("Policy Checker", "http://localhost:8002"),
        ("Approval Workflow", "http://localhost:8003"),
        ("Expense Orchestrator", "http://localhost:8004"),
    ]
    
    opener = get_opener()
    
    for name, url in agents:
        try:
            req = urllib.request.Request(f"{url}/.well-known/agent.json")
            with opener.open(req, timeout=5) as response:
                card = json.loads(response.read().decode())
                print(f"✓ {name}")
                print(f"  URL: {url}")
                print(f"  Description: {card.get('description', 'N/A')}")
                skills = card.get("skills", [])
                print(f"  Skills: {', '.join(s.get('id', '') for s in skills)}")
                print()
        except Exception as e:
            print(f"✗ {name} - {url}")
            print(f"  Error: {e}")
            print()
    
    time.sleep(1)


def demo_individual_agents():
    """Demo calling individual agents."""
    print("\n" + "="*60)
    print("📋 Demo: Individual Agent Calls")
    print("="*60 + "\n")
    
    # 1. Analyze a receipt
    print("1. Analyzing receipt rcpt-001...")
    result = call_agent(
        "http://localhost:8001",
        "tasks/create",
        {"skill": "analyze_receipt", "receipt_id": "rcpt-001"},
        "demo-receipt-1"
    )
    if "result" in result:
        data = result["result"]["extracted_data"]
        print(f"   ✓ {data['vendor']} - ${data['amount']:.2f} ({data['category']})")
    time.sleep(0.5)
    
    # 2. Check policy
    print("\n2. Checking policy for a $500 lodging expense...")
    result = call_agent(
        "http://localhost:8002",
        "tasks/create",
        {"skill": "check_policy", "expense": {
            "category": "Lodging",
            "amount": 500.00,
            "vendor": "Ritz Carlton"
        }},
        "demo-policy-1"
    )
    if "result" in result:
        r = result["result"]
        print(f"   Compliant: {r['compliant']}")
        print(f"   Requires Approval: {r['requires_approval']}")
        if r['warnings']:
            print(f"   Warnings: {[w['message'] for w in r['warnings']]}")
    time.sleep(0.5)
    
    # 3. Submit for approval
    print("\n3. Submitting an expense for approval...")
    result = call_agent(
        "http://localhost:8003",
        "tasks/create",
        {"skill": "submit_for_approval", "expense_id": "EXP-DEMO", 
         "amount": 150.00, "submitter": "Demo User", "category": "Meals"},
        "demo-approval-1"
    )
    if "result" in result:
        r = result["result"]
        print(f"   Approval ID: {r['approval_id']}")
        print(f"   Status: {r['status']}")
        print(f"   Assigned to: {r['assigned_to']}")
    time.sleep(1)


def demo_full_workflow():
    """Demo the full expense workflow through the orchestrator."""
    print("\n" + "="*60)
    print("🎯 Demo: Full Expense Workflow")
    print("="*60 + "\n")
    
    print("Processing expense report with 4 receipts...")
    print("This will trigger 9+ inter-agent calls:\n")
    print("  Orchestrator → Receipt Agent (x4)")
    print("  Orchestrator → Policy Agent (x4)")
    print("  Orchestrator → Approval Agent (x1)")
    print()
    
    result = call_agent(
        "http://localhost:8004",
        "tasks/create",
        {
            "skill": "process_expense",
            "receipt_ids": ["rcpt-001", "rcpt-002", "rcpt-003", "rcpt-004"],
            "submitter": "Sarah Chen",
            "description": "Q1 2026 Client Meeting Trip - San Francisco"
        },
        "workflow-demo-1"
    )
    
    print("\n" + "-"*60)
    print("📊 Workflow Result:")
    print("-"*60)
    
    if "result" in result:
        r = result["result"]
        print(f"\n  Expense ID:      {r['expense_id']}")
        print(f"  Status:          {r['status']}")
        print(f"  Total Amount:    ${r['total_amount']:.2f}")
        print(f"  Receipts:        {r['receipts_processed']}")
        print(f"  Policy OK:       {r['policy_compliant']}")
        
        if r['approval']:
            print(f"\n  Approval ID:     {r['approval']['approval_id']}")
            print(f"  Approval Status: {r['approval']['status']}")
            print(f"  Approver:        {r['approval']['assigned_to']}")
        
        print("\n  Expenses:")
        for exp in r['expenses']:
            print(f"    - {exp['vendor']}: ${exp['amount']:.2f} ({exp['category']})")
    else:
        print(f"  Error: {result.get('error', 'Unknown error')}")


def main():
    print("\n" + "="*60)
    print("💼 A2A Trace - Expense Workflow Demo")
    print("="*60)
    print("\nThis demo shows a realistic multi-agent expense reimbursement")
    print("workflow. Watch the a2a-trace UI to see all communication!\n")
    
    # First, discover all agents (triggers agent card requests)
    discover_agents()
    
    # Then demo individual agents
    demo_individual_agents()
    
    # Finally, run the full workflow
    demo_full_workflow()
    
    print("\n" + "="*60)
    print("✅ Demo Complete!")
    print("="*60)
    print("\nCheck the A2A Trace UI to see:")
    print("  - All request/response pairs in the Timeline")
    print("  - 4 discovered agents in the Agents tab")
    print("  - Insights about the workflow patterns")
    print()


if __name__ == "__main__":
    main()

