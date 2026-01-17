#!/usr/bin/env python3
"""
Weather Agent - A mock A2A agent that returns weather data.
Demonstrates task creation with structured output.
"""

import json
import uuid
import random
from http.server import HTTPServer, BaseHTTPRequestHandler
from datetime import datetime

PORT = 8002

AGENT_CARD = {
    "name": "Weather Agent",
    "description": "Provides current weather information for any city",
    "url": f"http://localhost:{PORT}",
    "version": "1.0.0",
    "protocol_version": "0.1",
    "capabilities": {
        "streaming": False,
        "push_notifications": False
    },
    "skills": [
        {
            "id": "get_weather",
            "name": "Get Weather",
            "description": "Returns current weather for a specified city",
            "examples": [
                "What's the weather in London?",
                "Get weather for Tokyo"
            ]
        },
        {
            "id": "get_forecast",
            "name": "Get Forecast",
            "description": "Returns 5-day weather forecast",
            "examples": [
                "What's the forecast for Paris?",
                "5-day forecast for New York"
            ]
        }
    ]
}

# Mock weather data
WEATHER_CONDITIONS = ["Sunny", "Cloudy", "Rainy", "Partly Cloudy", "Overcast", "Clear"]
CITIES_TIMEZONE = {
    "london": "Europe/London",
    "tokyo": "Asia/Tokyo",
    "new york": "America/New_York",
    "paris": "Europe/Paris",
    "sydney": "Australia/Sydney",
    "mumbai": "Asia/Kolkata",
    "berlin": "Europe/Berlin",
    "toronto": "America/Toronto",
}

tasks = {}


def get_mock_weather(city):
    """Generate mock weather data for a city."""
    city_lower = city.lower()
    
    # Seed random based on city name for consistent results
    random.seed(hash(city_lower) % 1000)
    
    temp_base = random.randint(5, 30)
    
    return {
        "city": city.title(),
        "country": "Unknown" if city_lower not in CITIES_TIMEZONE else city_lower.split()[0].title(),
        "temperature": {
            "value": temp_base,
            "unit": "celsius",
            "feels_like": temp_base + random.randint(-3, 3)
        },
        "condition": random.choice(WEATHER_CONDITIONS),
        "humidity": random.randint(30, 90),
        "wind": {
            "speed": random.randint(5, 30),
            "unit": "km/h",
            "direction": random.choice(["N", "NE", "E", "SE", "S", "SW", "W", "NW"])
        },
        "updated_at": datetime.now().isoformat()
    }


def get_mock_forecast(city, days=5):
    """Generate mock forecast data."""
    forecast = []
    random.seed(hash(city.lower()) % 1000)
    
    for i in range(days):
        temp_high = random.randint(15, 35)
        forecast.append({
            "day": i + 1,
            "date": datetime.now().strftime("%Y-%m-%d"),
            "condition": random.choice(WEATHER_CONDITIONS),
            "temperature": {
                "high": temp_high,
                "low": temp_high - random.randint(5, 12),
                "unit": "celsius"
            },
            "precipitation_chance": random.randint(0, 100)
        })
    
    return {
        "city": city.title(),
        "forecast": forecast,
        "generated_at": datetime.now().isoformat()
    }


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
        city = params.get("city", params.get("location", "Unknown"))
        skill = params.get("skill", "get_weather")

        print(f"  City: {city}, Skill: {skill}")

        # Generate weather data
        if skill == "get_forecast":
            days = params.get("days", 5)
            result = get_mock_forecast(city, days)
        else:
            result = get_mock_weather(city)

        task = {
            "id": task_id,
            "status": "completed",
            "skill": skill,
            "created_at": datetime.now().isoformat(),
            "result": result
        }
        tasks[task_id] = task

        print(f"  Created task: {task_id}")
        print(f"  Weather: {result.get('condition', 'N/A')} {result.get('temperature', {}).get('value', 'N/A')}°C")

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

    def do_OPTIONS(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()


def main():
    server = HTTPServer(("", PORT), A2AHandler)
    print(f"🌤️  Weather Agent starting on port {PORT}")
    print(f"   Agent card: http://localhost:{PORT}/.well-known/agent.json")
    print(f"   Health: http://localhost:{PORT}/health")
    print()
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n👋 Weather Agent shutting down")
        server.shutdown()


if __name__ == "__main__":
    main()

