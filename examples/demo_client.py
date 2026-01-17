#!/usr/bin/env python3
"""
Demo Client - Makes requests to A2A agents for testing.
Use with a2a-trace to visualize the agent interactions:

    a2a-trace -- python demo_client.py
"""

import json
import urllib.request
import urllib.error
import time
import sys
import os

# Agent URLs
ECHO_AGENT = os.environ.get("ECHO_AGENT_URL", "http://localhost:8001")
WEATHER_AGENT = os.environ.get("WEATHER_AGENT_URL", "http://localhost:8002")
ORCHESTRATOR_AGENT = os.environ.get("ORCHESTRATOR_AGENT_URL", "http://localhost:8003")


def call_agent(agent_url, method, params):
    """Make a JSON-RPC call to an A2A agent."""
    request_data = {
        "jsonrpc": "2.0",
        "method": method,
        "id": "demo-" + str(int(time.time() * 1000)),
        "params": params
    }
    
    print(f"\n📤 Calling {agent_url}")
    print(f"   Method: {method}")
    print(f"   Params: {json.dumps(params)}")
    
    data = json.dumps(request_data).encode()
    req = urllib.request.Request(
        agent_url,
        data=data,
        headers={"Content-Type": "application/json"}
    )
    
    try:
        start = time.time()
        with urllib.request.urlopen(req, timeout=30) as response:
            result = json.loads(response.read().decode())
            elapsed = (time.time() - start) * 1000
            
            print(f"📥 Response ({elapsed:.0f}ms)")
            
            if "result" in result:
                print(f"   ✅ Success")
                # Print key parts of result
                task_result = result["result"]
                if isinstance(task_result, dict):
                    if "result" in task_result:
                        inner = task_result["result"]
                        if isinstance(inner, dict):
                            for key, value in list(inner.items())[:3]:
                                print(f"   {key}: {json.dumps(value)[:50]}")
            else:
                print(f"   ❌ Error: {result.get('error', {}).get('message', 'Unknown')}")
            
            return result
    except urllib.error.URLError as e:
        print(f"   ❌ Connection error: {e}")
        return None
    except Exception as e:
        print(f"   ❌ Error: {e}")
        return None


def demo_echo_agent():
    """Demo: Echo Agent"""
    print("\n" + "="*60)
    print("Demo 1: Echo Agent")
    print("="*60)
    
    call_agent(ECHO_AGENT, "tasks/create", {
        "message": "Hello from the demo client! This is a test message."
    })
    time.sleep(0.5)


def demo_weather_agent():
    """Demo: Weather Agent"""
    print("\n" + "="*60)
    print("Demo 2: Weather Agent")
    print("="*60)
    
    cities = ["London", "Tokyo", "New York"]
    
    for city in cities:
        call_agent(WEATHER_AGENT, "tasks/create", {
            "city": city,
            "skill": "get_weather"
        })
        time.sleep(0.3)


def demo_orchestrator():
    """Demo: Orchestrator Agent (multi-hop)"""
    print("\n" + "="*60)
    print("Demo 3: Orchestrator Agent (Multi-hop)")
    print("="*60)
    
    # This will trigger Echo Agent + Weather Agent calls
    call_agent(ORCHESTRATOR_AGENT, "tasks/create", {
        "skill": "greet_with_weather",
        "name": "Developer",
        "city": "San Francisco"
    })
    
    time.sleep(0.5)
    
    # Multi-city weather
    call_agent(ORCHESTRATOR_AGENT, "tasks/create", {
        "skill": "multi_city_weather",
        "cities": ["Paris", "Berlin", "Sydney"]
    })


def demo_agent_discovery():
    """Demo: Agent Discovery"""
    print("\n" + "="*60)
    print("Demo 4: Agent Discovery")
    print("="*60)
    
    agents = [
        ("Echo Agent", ECHO_AGENT),
        ("Weather Agent", WEATHER_AGENT),
        ("Orchestrator Agent", ORCHESTRATOR_AGENT)
    ]
    
    for name, url in agents:
        try:
            req = urllib.request.Request(f"{url}/.well-known/agent.json")
            with urllib.request.urlopen(req, timeout=5) as response:
                card = json.loads(response.read().decode())
                print(f"\n🤖 {card.get('name', name)}")
                print(f"   URL: {card.get('url', url)}")
                print(f"   Version: {card.get('version', 'unknown')}")
                skills = card.get("skills", [])
                if skills:
                    print(f"   Skills: {', '.join(s.get('name', s.get('id', '?')) for s in skills)}")
        except Exception as e:
            print(f"\n⚠️  {name} not available: {e}")


def main():
    print("🔍 A2A Trace Demo Client")
    print("   Running with: a2a-trace -- python demo_client.py")
    print("   View trace: http://localhost:8080/ui")
    
    # Check if agents are running
    print("\n📡 Checking agent availability...")
    
    available = True
    for name, url in [("Echo", ECHO_AGENT), ("Weather", WEATHER_AGENT), ("Orchestrator", ORCHESTRATOR_AGENT)]:
        try:
            req = urllib.request.Request(f"{url}/health")
            urllib.request.urlopen(req, timeout=2)
            print(f"   ✅ {name} Agent ({url})")
        except:
            print(f"   ❌ {name} Agent ({url}) - not running")
            available = False
    
    if not available:
        print("\n⚠️  Some agents are not running.")
        print("   Start them with:")
        print("     python examples/echo_agent.py")
        print("     python examples/weather_agent.py")
        print("     python examples/orchestrator_agent.py")
        print("\n   Continuing with available agents...")
    
    time.sleep(1)
    
    # Run demos
    try:
        demo_agent_discovery()
        time.sleep(0.5)
        
        demo_echo_agent()
        time.sleep(0.5)
        
        demo_weather_agent()
        time.sleep(0.5)
        
        demo_orchestrator()
        
    except KeyboardInterrupt:
        print("\n\n👋 Demo interrupted")
        sys.exit(0)
    
    print("\n" + "="*60)
    print("Demo Complete!")
    print("="*60)
    print("\n📊 Check http://localhost:8080/ui to see the trace visualization")
    print("   You should see all the agent interactions in the timeline.")


if __name__ == "__main__":
    main()

