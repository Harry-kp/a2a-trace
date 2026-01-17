#!/bin/bash
set -e

# A2A Trace build script
# Builds both the UI and the Go binary

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔨 Building A2A Trace${NC}"
echo ""

# Build UI
echo -e "${YELLOW}📦 Building UI...${NC}"
cd "$PROJECT_ROOT/ui"
npm install --silent
npm run build

# Copy UI build to Go embed location
echo -e "${YELLOW}📁 Preparing UI for embedding...${NC}"
mkdir -p "$PROJECT_ROOT/cmd/a2a-trace/ui"
rm -rf "$PROJECT_ROOT/cmd/a2a-trace/ui/out"
cp -r "$PROJECT_ROOT/ui/out" "$PROJECT_ROOT/cmd/a2a-trace/ui/"

# Build Go binary
echo -e "${YELLOW}🔧 Building Go binary...${NC}"
cd "$PROJECT_ROOT"

# Get version info
VERSION="${VERSION:-dev}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')}"
BUILD_DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

# Build flags
LDFLAGS="-s -w"
LDFLAGS="$LDFLAGS -X github.com/harry-kp/a2a-trace/internal/cli.Version=$VERSION"
LDFLAGS="$LDFLAGS -X github.com/harry-kp/a2a-trace/internal/cli.Commit=$COMMIT"
LDFLAGS="$LDFLAGS -X github.com/harry-kp/a2a-trace/internal/cli.BuildDate=$BUILD_DATE"

# Build for current platform
go build -ldflags="$LDFLAGS" -o bin/a2a-trace ./cmd/a2a-trace

echo ""
echo -e "${GREEN}✅ Build complete!${NC}"
echo -e "   Binary: ${PROJECT_ROOT}/bin/a2a-trace"
echo ""

# Show version
./bin/a2a-trace --version

