# 🔍 A2A Trace

**Visual debugger for multi-agent systems built on Google's A2A protocol**

[![CI](https://github.com/harry-kp/a2a-trace/actions/workflows/ci.yml/badge.svg)](https://github.com/harry-kp/a2a-trace/actions/workflows/ci.yml)
[![Release](https://github.com/harry-kp/a2a-trace/actions/workflows/release.yml/badge.svg)](https://github.com/harry-kp/a2a-trace/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/harry-kp/a2a-trace)](https://goreportcard.com/report/github.com/harry-kp/a2a-trace)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

<p align="center">
  <img src="docs/demo.gif" alt="A2A Trace Demo" width="800">
</p>

---

## What is A2A Trace?

A2A Trace intercepts and visualizes [Agent-to-Agent (A2A) protocol](https://github.com/google/A2A) communications, helping you debug complex multi-agent systems with ease.

### Features

- 🚀 **One-command setup** - Just prefix your command with `a2a-trace --`
- 📊 **Real-time visualization** - Watch agent interactions as they happen
- 🔍 **Message inspection** - Drill down into request/response payloads
- 🤖 **Agent discovery** - Automatically detect and display agent info
- ⚡ **Insights & analysis** - Detect slow responses, errors, and retry loops
- 📦 **Single binary** - No dependencies, works everywhere
- 🌐 **Language agnostic** - Works with any A2A agent implementation

---

## Quick Start

### Installation

#### Using Go

```bash
go install github.com/harry-kp/a2a-trace/cmd/a2a-trace@latest
```

#### Using Homebrew (macOS)

```bash
brew install harry-kp/tap/a2a-trace
```

#### Download Binary

Download from [Releases](https://github.com/harry-kp/a2a-trace/releases):

```bash
# macOS (Apple Silicon)
curl -L https://github.com/harry-kp/a2a-trace/releases/latest/download/a2a-trace-darwin-arm64.tar.gz | tar xz
sudo mv a2a-trace-darwin-arm64 /usr/local/bin/a2a-trace

# Linux (x86_64)
curl -L https://github.com/harry-kp/a2a-trace/releases/latest/download/a2a-trace-linux-amd64.tar.gz | tar xz
sudo mv a2a-trace-linux-amd64 /usr/local/bin/a2a-trace
```

### Usage

Wrap your A2A agent command with `a2a-trace --`:

```bash
# Node.js agent
a2a-trace -- node my-agent.js

# Python agent
a2a-trace -- python agent.py

# Go agent
a2a-trace -- ./my-go-agent

# With custom port
a2a-trace --port 9000 -- npm start
```

Then open **http://localhost:8080/ui** in your browser to see the trace UI.

---

## How It Works

```
┌──────────────────────────────────────────────────────────────────┐
│                                                                  │
│   $ a2a-trace -- node my-agent.js                                │
│                                                                  │
│   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐     │
│   │   Your      │      │  A2A Trace  │      │  External   │     │
│   │   Agent     │─────▶│   (proxy)   │─────▶│   Agents    │     │
│   │             │◀─────│             │◀─────│             │     │
│   └─────────────┘      └──────┬──────┘      └─────────────┘     │
│                               │                                  │
│                               ▼                                  │
│                        ┌─────────────┐                          │
│                        │  Browser UI │                          │
│                        │ (real-time) │                          │
│                        └─────────────┘                          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

1. **A2A Trace** spawns your agent as a child process
2. Sets `HTTP_PROXY` environment variable automatically  
3. All A2A HTTP traffic flows through the trace proxy
4. Messages are logged to SQLite and broadcast via WebSocket
5. The web UI displays everything in real-time

---

## CLI Reference

```
Usage:
  a2a-trace [flags] -- <command> [args...]

Flags:
  -p, --port int      Proxy port (default 8080)
      --ui-port int   UI port (default: same as proxy)
      --db string     SQLite database path (default: in-memory)
  -v, --verbose       Verbose output
      --no-ui         Don't serve the web UI
  -h, --help          Help for a2a-trace
      --version       Version info
```

### Examples

```bash
# Basic usage
a2a-trace -- node agent.js

# Custom proxy port
a2a-trace --port 9000 -- python agent.py

# Persist traces to file
a2a-trace --db ./traces.db -- ./agent

# Verbose mode (see all requests in terminal)
a2a-trace --verbose -- npm run agent

# Without UI (CLI only)
a2a-trace --no-ui -- ./agent
```

---

## API Endpoints

The trace server exposes REST endpoints:

| Endpoint | Description |
|----------|-------------|
| `GET /api/messages` | List all intercepted messages |
| `GET /api/agents` | List discovered agents |
| `GET /api/insights` | List detected issues |
| `GET /api/trace` | Current trace info |
| `GET /api/summary` | Statistics summary |
| `GET /api/export` | Export trace as JSON |
| `WS /ws` | WebSocket for real-time updates |

---

## Development

### Prerequisites

- Go 1.22+
- Node.js 20+
- npm

### Building from Source

```bash
git clone https://github.com/harry-kp/a2a-trace.git
cd a2a-trace

# Build everything
./scripts/build.sh

# Or manually:
cd ui && npm install && npm run build && cd ..
mkdir -p cmd/a2a-trace/ui && cp -r ui/out cmd/a2a-trace/ui/
go build -o bin/a2a-trace ./cmd/a2a-trace
```

### Running Tests

```bash
# Go tests
go test ./...

# UI type check
cd ui && npm run build
```

---

## Roadmap

- [ ] **HAR export** - Export traces in HTTP Archive format
- [ ] **Replay** - Replay captured requests
- [ ] **Diff mode** - Compare two trace sessions
- [ ] **OpenTelemetry** - Export traces to OTEL backends
- [ ] **Streaming support** - Better SSE/streaming visualization
- [ ] **VSCode extension** - Integrated debugging experience

---

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License - see [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [Google A2A Protocol](https://github.com/google/A2A) - The protocol this tool debugs
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Gorilla WebSocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [modernc.org/sqlite](https://modernc.org/sqlite) - Pure Go SQLite driver

---

<p align="center">
  Made with ❤️ by <a href="https://github.com/harry-kp">Harshit Chaudhary</a>
</p>

