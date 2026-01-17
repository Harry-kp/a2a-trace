# Contributing to A2A Trace

Thank you for your interest in contributing to A2A Trace! This document provides guidelines and information about contributing.

## Code of Conduct

By participating in this project, you agree to abide by our code of conduct: be respectful, inclusive, and constructive in all interactions.

## How to Contribute

### Reporting Bugs

1. Check existing issues to avoid duplicates
2. Use the bug report template
3. Include:
   - A2A Trace version (`a2a-trace --version`)
   - Operating system and version
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or screenshots

### Suggesting Features

1. Check existing issues/discussions
2. Use the feature request template
3. Describe:
   - The problem you're trying to solve
   - Your proposed solution
   - Alternative approaches considered

### Pull Requests

1. Fork the repository
2. Create a feature branch from `main`
3. Make your changes
4. Ensure tests pass
5. Update documentation if needed
6. Submit a PR with a clear description

## Development Setup

### Prerequisites

- Go 1.22 or later
- Node.js 20 or later
- npm

### Building

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/a2a-trace.git
cd a2a-trace

# Install Go dependencies
go mod download

# Build UI
cd ui
npm install
npm run build
cd ..

# Build binary
mkdir -p cmd/a2a-trace/ui
cp -r ui/out cmd/a2a-trace/ui/
go build -o bin/a2a-trace ./cmd/a2a-trace
```

### Testing

```bash
# Run Go tests
go test -v ./...

# Run with race detector
go test -race ./...

# UI type checking
cd ui && npm run build
```

### Code Style

#### Go
- Follow standard Go formatting (`gofmt`)
- Use meaningful variable names
- Add comments for exported functions
- Keep functions focused and small

#### TypeScript/React
- Use TypeScript strict mode
- Follow React best practices
- Use functional components
- Prefer named exports

## Project Structure

```
a2a-trace/
├── cmd/a2a-trace/     # Main binary entrypoint
├── internal/          # Internal packages
│   ├── cli/           # CLI parsing
│   ├── proxy/         # HTTP proxy
│   ├── process/       # Child process management
│   ├── store/         # SQLite storage
│   ├── websocket/     # WebSocket server
│   └── analyzer/      # Pattern detection
├── ui/                # Next.js frontend
│   ├── src/
│   │   ├── app/       # Pages
│   │   ├── components/# React components
│   │   ├── hooks/     # Custom hooks
│   │   └── lib/       # Utilities & store
│   └── ...
├── scripts/           # Build scripts
└── .github/           # CI/CD workflows
```

## Commit Messages

Follow conventional commits:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

Examples:
```
feat(proxy): add support for gRPC interception
fix(ui): correct timeline sorting order
docs: update installation instructions
```

## Release Process

Releases are automated via GitHub Actions when a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Getting Help

- Open an issue for bugs or features
- Start a discussion for questions
- Check existing docs and issues first

## Recognition

Contributors are recognized in:
- The README acknowledgments section
- Release notes
- GitHub contributors page

Thank you for contributing to A2A Trace!

