#!/bin/bash
# Expense Workflow Demo - Starts all agents and runs the demo

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "🚀 Starting Expense Workflow Demo"
echo "=================================="
echo ""

# Kill any existing agents
pkill -f "receipt_agent.py" 2>/dev/null || true
pkill -f "policy_agent.py" 2>/dev/null || true
pkill -f "approval_agent.py" 2>/dev/null || true
pkill -f "expense_orchestrator.py" 2>/dev/null || true
sleep 1

# Start agents in background
echo "Starting agents..."
python3 receipt_agent.py &
sleep 0.5
python3 policy_agent.py &
sleep 0.5
python3 approval_agent.py &
sleep 0.5
python3 expense_orchestrator.py &
sleep 1

echo ""
echo "✅ All agents started:"
echo "   - Receipt Analyzer:    http://localhost:8001"
echo "   - Policy Checker:      http://localhost:8002"
echo "   - Approval Workflow:   http://localhost:8003"
echo "   - Expense Orchestrator: http://localhost:8004"
echo ""
echo "Now run a2a-trace to see the workflow:"
echo ""
echo "  a2a-trace -- python3 run_client.py"
echo ""
echo "Or manually trigger the workflow:"
echo ""
echo '  curl --proxy http://localhost:8080 -X POST http://localhost:8004/ \'
echo '    -H "Content-Type: application/json" \'
echo '    -d '"'"'{"jsonrpc":"2.0","method":"tasks/create","id":"demo-1","params":{"skill":"process_expense","receipt_ids":["rcpt-001","rcpt-002","rcpt-003"],"submitter":"John Smith"}}'"'"''
echo ""

# Wait for user to stop
trap "echo 'Stopping agents...'; pkill -f 'receipt_agent.py|policy_agent.py|approval_agent.py|expense_orchestrator.py' 2>/dev/null; exit 0" INT TERM

echo "Press Ctrl+C to stop all agents..."
wait

