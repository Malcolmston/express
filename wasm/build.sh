#!/usr/bin/env bash
# Build the express utilities WebAssembly adapter.
set -euo pipefail
cd "$(dirname "$0")"
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o express.wasm .
echo "built express.wasm ($(du -h express.wasm | cut -f1))"
