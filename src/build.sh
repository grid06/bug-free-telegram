#!/bin/bash

echo "🔨 Building Jarvis for Windows..."

# Environment
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=0

# ===== ODDIY BUILD (avval test) =====
echo "📦 Step 1: Normal build..."
go build \
    -trimpath \
    -ldflags="-s -w -H=windowsgui" \
    -o build/jarvis_normal.exe \
    .

echo "✅ Normal build done: build/jarvis_normal.exe"

# ===== GARBLE BUILD =====
echo "🔒 Step 2: Garble obfuscated build..."
garble \
    -tiny \
    -literals \
    -seed=random \
    build \
    -trimpath \
    -ldflags="-s -w -H=windowsgui" \
    -o build/jarvis.exe \
    .

echo "✅ Garble build done: build/jarvis.exe"

# File size ko'rsat
echo ""
echo "📊 File sizes:"
ls -lh build/*.exe

echo ""
echo "🎉 Build complete!"