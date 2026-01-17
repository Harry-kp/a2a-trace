# Expense Reimbursement Workflow Demo

A realistic multi-agent workflow demonstrating enterprise expense processing using the A2A protocol. This is perfect for showcasing **a2a-trace** capabilities.

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    User / Client                              │
└─────────────────────────┬────────────────────────────────────┘
                          │
                          ▼
┌──────────────────────────────────────────────────────────────┐
│              Expense Orchestrator (8004)                      │
│        Coordinates the complete workflow                      │
└───────┬─────────────────┬─────────────────┬──────────────────┘
        │                 │                 │
        ▼                 ▼                 ▼
┌───────────────┐ ┌───────────────┐ ┌───────────────┐
│   Receipt     │ │    Policy     │ │   Approval    │
│   Analyzer    │ │    Checker    │ │   Workflow    │
│    (8001)     │ │    (8002)     │ │    (8003)     │
│               │ │               │ │               │
│ • OCR/Extract │ │ • Limits      │ │ • Routing     │
│ • Validate    │ │ • Compliance  │ │ • SLA         │
│ • Categorize  │ │ • Warnings    │ │ • History     │
└───────────────┘ └───────────────┘ └───────────────┘
```

## 🚀 Quick Start

### 1. Start the Agent Network

```bash
cd examples/expense_workflow

# Start all 4 agents
./run_demo.sh
```

### 2. Run with a2a-trace

```bash
# In a new terminal, from the a2a-trace root
./bin/a2a-trace -- python3 examples/expense_workflow/run_client.py
```

### 3. Open the UI

Navigate to **http://localhost:8080/ui** to see:
- 📊 **Timeline**: All 13+ request/response pairs
- 🤖 **Agents**: 4 discovered agents with capabilities
- 💡 **Insights**: Workflow patterns and timing analysis

## 🎬 What the Demo Shows

The demo processes an expense report with 4 receipts through a realistic workflow:

1. **Agent Discovery**: Fetches agent cards from all 4 agents
2. **Receipt Analysis**: Calls Receipt Agent 4x to extract expense data
3. **Policy Checking**: Calls Policy Agent 4x to validate compliance
4. **Approval Submission**: Calls Approval Agent to route for approval

### Sample Receipts

| ID | Vendor | Category | Amount | Notes |
|----|--------|----------|--------|-------|
| rcpt-001 | Hilton Hotels | Lodging | $289.00 | San Francisco |
| rcpt-002 | United Airlines | Transportation | $542.50 | SFO → JFK |
| rcpt-003 | The French Laundry | Meals | $385.00 | Client dinner (over limit!) |
| rcpt-004 | Uber | Transportation | $45.80 | Airport transfer |

### Policy Rules

- **Lodging**: $300/day limit, approval above $250
- **Transportation**: $600/day, pre-approved vendors only
- **Meals**: $75/day, $30/person, requires client name
- **Misc**: $100/day

## 📁 Files

```
expense_workflow/
├── receipt_agent.py      # OCR/data extraction agent
├── policy_agent.py       # Policy compliance agent
├── approval_agent.py     # Approval routing agent
├── expense_orchestrator.py # Workflow coordinator
├── run_demo.sh           # Starts all agents
├── run_client.py         # Demo client script
└── README.md             # This file
```

## 🔧 Manual Testing

```bash
# Analyze a receipt
curl --proxy http://localhost:8080 -X POST http://localhost:8001/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tasks/create","id":"1","params":{"skill":"analyze_receipt","receipt_id":"rcpt-001"}}'

# Check policy
curl --proxy http://localhost:8080 -X POST http://localhost:8002/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tasks/create","id":"2","params":{"skill":"check_policy","expense":{"category":"Meals","amount":150,"vendor":"Restaurant"}}}'

# Full workflow
curl --proxy http://localhost:8080 -X POST http://localhost:8004/ \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"tasks/create","id":"3","params":{"skill":"process_expense","receipt_ids":["rcpt-001","rcpt-002"],"submitter":"Demo User"}}'
```

## 🎯 Why This Demo?

This demo is designed to showcase real-world A2A patterns:

1. **Multi-agent coordination**: Orchestrator calling multiple specialized agents
2. **Sequential workflows**: Receipt → Policy → Approval chain
3. **Business logic**: Real expense policies with limits and validations
4. **Error handling**: Policy violations, approval routing
5. **Realistic data**: Actual expense categories, vendors, amounts

Perfect for understanding how A2A enables complex enterprise workflows!

