#!/bin/bash
set -e

# A2A Trace cross-platform build script
# Builds binaries for Linux, macOS, and Windows

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🔨 Building A2A Trace for all platforms${NC}"
echo ""

# Build UI first
echo -e "${YELLOW}📦 Building UI...${NC}"
cd "$PROJECT_ROOT/ui"
npm install --silent
npm run build

# Copy UI build to Go embed location
echo -e "${YELLOW}📁 Preparing UI for embedding...${NC}"
mkdir -p "$PROJECT_ROOT/cmd/a2a-trace/ui"
rm -rf "$PROJECT_ROOT/cmd/a2a-trace/ui/out"
cp -r "$PROJECT_ROOT/ui/out" "$PROJECT_ROOT/cmd/a2a-trace/ui/"

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

# Create output directory
mkdir -p "$PROJECT_ROOT/dist"

# Build targets
TARGETS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

echo -e "${YELLOW}🔧 Building binaries...${NC}"

for target in "${TARGETS[@]}"; do
    GOOS="${target%/*}"
    GOARCH="${target#*/}"
    
    OUTPUT_NAME="a2a-trace-${GOOS}-${GOARCH}"
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    echo "  Building ${OUTPUT_NAME}..."
    
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags="$LDFLAGS" \
        -o "dist/${OUTPUT_NAME}" \
        ./cmd/a2a-trace
done

echo ""
echo -e "${GREEN}✅ All builds complete!${NC}"
echo ""
ls -la "$PROJECT_ROOT/dist/"

