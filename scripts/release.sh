#!/bin/bash

VERSION="1.0.1"
BINARY_NAME="jp"
PLATFORMS=("darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

rm -rf dist
mkdir -p dist

for platform in "${PLATFORMS[@]}"; do
    OS="${platform%/*}"
    ARCH="${platform#*/}"
    output_name="$BINARY_NAME-$OS-$ARCH"
    
    echo "Building for $OS/$ARCH..."
    GOOS=$OS GOARCH=$ARCH go build -o "dist/$BINARY_NAME" ./cmd/jp
    
    cd dist
    tar czf "$output_name.tar.gz" "$BINARY_NAME"
    shasum -a 256 "$output_name.tar.gz"
    cd ..
done

echo "Build complete. Please create a new release on GitHub with version v$VERSION"
echo "Upload the .tar.gz files from the dist directory"
echo "Update the Formula/jp.rb file in homebrew-tap with the new SHA256 values" 