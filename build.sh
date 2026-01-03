#!/bin/bash

# Build script for go-mcp-atlassian
# Creates binaries for multiple platforms

set -e

APP_NAME="go-mcp-atlassian"
VERSION="${VERSION:-dev}"

# Clean previous builds
rm -rf dist
mkdir -p dist

echo "Building $APP_NAME version $VERSION..."

# Build for each platform
PLATFORMS=(
    "darwin/amd64"
    "darwin/arm64"
    "linux/amd64"
    "linux/arm64"
    "windows/amd64"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"
    OUTPUT="dist/${APP_NAME}-${GOOS}-${GOARCH}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT="${OUTPUT}.exe"
    fi

    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build -ldflags="-s -w -X main.Version=$VERSION" -o "$OUTPUT" .
done

# Create universal binary for macOS
if [ -f "dist/${APP_NAME}-darwin-amd64" ] && [ -f "dist/${APP_NAME}-darwin-arm64" ]; then
    echo "Creating macOS universal binary..."
    lipo -create -output "dist/${APP_NAME}-darwin-universal" \
        "dist/${APP_NAME}-darwin-amd64" \
        "dist/${APP_NAME}-darwin-arm64"
fi

# Generate checksums
echo "Generating checksums..."
cd dist
if command -v sha256sum &> /dev/null; then
    sha256sum * > checksums.txt
elif command -v shasum &> /dev/null; then
    shasum -a 256 * > checksums.txt
fi
cd ..

echo "Build complete! Binaries are in the dist/ directory."
ls -la dist/
