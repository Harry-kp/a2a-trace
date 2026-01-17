# A2A Trace Demo Agents

Simple A2A agents for testing and demonstrating a2a-trace.

## Prerequisites

- Python 3.9+
- No external dependencies (uses standard library only)

## Agents

### 1. Echo Agent (Port 8001)

A simple agent that echoes back any message it receives.

```bash
python echo_agent.py
```

### 2. Weather Agent (Port 8002)

Returns mock weather data for any city.

```bash
python weather_agent.py
```

### 3. Orchestrator Agent (Port 8003)

Calls other agents to demonstrate multi-hop communication.

```bash
python orchestrator_agent.py
```

## Running with A2A Trace

Start the agents in separate terminals:

```bash
# Terminal 1
python examples/echo_agent.py

# Terminal 2
python examples/weather_agent.py

# Terminal 3
python examples/orchestrator_agent.py
```

Then run the demo client through a2a-trace:

```bash
# Terminal 4
a2a-trace -- python examples/demo_client.py
```

Open http://localhost:8080/ui to see the trace visualization.

## Testing Individual Agents

### Echo Agent

```bash
curl -X POST http://localhost:8001 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tasks/create",
    "id": "1",
    "params": {
      "message": "Hello, Agent!"
    }
  }'
```

### Weather Agent

```bash
curl -X POST http://localhost:8002 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tasks/create",
    "id": "1",
    "params": {
      "city": "London"
    }
  }'
```

### Agent Cards

Each agent serves its card at `/.well-known/agent.json`:

```bash
curl http://localhost:8001/.well-known/agent.json
```

