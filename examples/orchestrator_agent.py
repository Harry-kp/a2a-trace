#!/usr/bin/env python3
"""
Orchestrator Agent - An A2A agent that coordinates other agents.
Demonstrates multi-hop communication patterns.
"""

import json
import uuid
import urllib.request
import urllib.error
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime
import os

PORT = 8003

# Other agent URLs
ECHO_AGENT = os.environ.get("ECHO_AGENT_URL", "http://localhost:8001")
WEATHER_AGENT = os.environ.get("WEATHER_AGENT_URL", "http://localhost:8002")

AGENT_CARD = {
    "name": "Orchestrator Agent",
    "description": "Coordinates multiple agents to complete complex tasks",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "greet_with_weather",
            "name": "Greet with Weather",
            "description": "Greets a user and includes weather information for their city",
            "examples": [
                "Greet John from London with the weather"
            ]
        },
        {
            "id": "multi_city_weather",
            "name": "Multi-City Weather",
            "description": "Gets weather for multiple cities at once",
            "examples": [
                "Get weather for London, Tokyo, and New York"
            ]
        }
    ],
    "dependencies": [
        {"url": ECHO_AGENT, "name": "Echo Agent"},
        {"url": WEATHER_AGENT, "name": "Weather Agent"}
    ]
}

tasks = {}


def call_agent(agent_url, method, params):
    """Make a JSON-RPC call to another A2A agent."""
    request_data = {
        "jsonrpc": "2.0",
        "method": method,
        "id": str(uuid.uuid4()),
        "params": params
    }
    
    print(f"    → Calling {agent_url}")
    print(f"      Method: {method}")
    
    data = json.dumps(request_data).encode()
    req = urllib.request.Request(
        agent_url,
        data=data,
        headers={"Content-Type": "application/json"}
    )
    
    try:
        with urllib.request.urlopen(req, timeout=30) as response:
            result = json.loads(response.read().decode())
            print(f"    ← Response received")
            return result
    except urllib.error.URLError as e:
        print(f"    ✗ Error: {e}")
        return {"error": {"code": -32000, "message": str(e)}}
    except Exception as e:
        print(f"    ✗ Error: {e}")
        return {"error": {"code": -32000, "message": str(e)}}


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
        print(f"  Params: {json.dumps(params)}")

        if method == "tasks/create":
            self.handle_create_task(request_id, params)
        elif method == "tasks/get":
            self.handle_get_task(request_id, params)
        else:
            self.send_json({
                "jsonrpc": "2.0",
                "error": {"code": -32601, "message": f"Method not found: {method}"},
                "id": request_id
            })

    def handle_create_task(self, request_id, params):
        task_id = str(uuid.uuid4())
        skill = params.get("skill", "greet_with_weather")

        print(f"  Task ID: {task_id}")
        print(f"  Skill: {skill}")

        if skill == "greet_with_weather":
            result = self.skill_greet_with_weather(params)
        elif skill == "multi_city_weather":
            result = self.skill_multi_city_weather(params)
        else:
            result = {"error": f"Unknown skill: {skill}"}

        task = {
            "id": task_id,
            "status": "completed" if "error" not in result else "failed",
            "skill": skill,
            "created_at": datetime.now().isoformat(),
            "result": result
        }
        tasks[task_id] = task

        self.send_json({
            "jsonrpc": "2.0",
            "result": task,
            "id": request_id
        })

    def skill_greet_with_weather(self, params):
        """Greet a user with weather information."""
        name = params.get("name", "Friend")
        city = params.get("city", "London")

        print(f"  Orchestrating: greet {name} with weather for {city}")

        # Step 1: Get greeting from Echo Agent
        echo_result = call_agent(ECHO_AGENT, "tasks/create", {
            "message": f"Hello, {name}! Welcome!"
        })

        if "error" in echo_result:
            return {"error": f"Echo Agent failed: {echo_result['error']}"}

        greeting = echo_result.get("result", {}).get("result", {}).get("echo", "Hello!")

        # Step 2: Get weather from Weather Agent
        weather_result = call_agent(WEATHER_AGENT, "tasks/create", {
            "city": city,
            "skill": "get_weather"
        })

        if "error" in weather_result:
            return {"error": f"Weather Agent failed: {weather_result['error']}"}

        weather = weather_result.get("result", {}).get("result", {})
        condition = weather.get("condition", "unknown")
        temp = weather.get("temperature", {}).get("value", "?")

        # Combine results
        return {
            "greeting": greeting,
            "weather_summary": f"The weather in {city} is {condition} with {temp}°C",
            "weather_details": weather,
            "orchestrated_at": datetime.now().isoformat()
        }

    def skill_multi_city_weather(self, params):
        """Get weather for multiple cities."""
        cities = params.get("cities", ["London", "Tokyo", "New York"])

        print(f"  Orchestrating: weather for {len(cities)} cities")

        results = {}
        errors = []

        for city in cities:
            weather_result = call_agent(WEATHER_AGENT, "tasks/create", {
                "city": city,
                "skill": "get_weather"
            })

            if "error" in weather_result:
                errors.append(f"{city}: {weather_result['error']}")
            else:
                weather = weather_result.get("result", {}).get("result", {})
                results[city] = {
                    "condition": weather.get("condition"),
                    "temperature": weather.get("temperature", {}).get("value"),
                    "unit": "celsius"
                }

        return {
            "cities": results,
            "errors": errors if errors else None,
            "orchestrated_at": datetime.now().isoformat()
        }

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

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()


def main():
    server = HTTPServer(("", PORT), A2AHandler)
    print(f"🎭 Orchestrator Agent starting on port {PORT}")
    print(f"   Agent card: http://localhost:{PORT}/.well-known/agent.json")
    print(f"   Health: http://localhost:{PORT}/health")
    print()
    print(f"   Dependencies:")
    print(f"   - Echo Agent: {ECHO_AGENT}")
    print(f"   - Weather Agent: {WEATHER_AGENT}")
    print()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n👋 Orchestrator Agent shutting down")
        server.shutdown()


if __name__ == "__main__":
    main()

