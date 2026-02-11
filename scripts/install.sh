#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BINARY="linear-tui"
INSTALL_DIR="${HOME}/.local/bin"

echo "Building ${BINARY}..."
cd "$REPO_DIR"
go build -o "${BINARY}" ./cmd/linear-tui

mkdir -p "$INSTALL_DIR"
ln -sf "${REPO_DIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} -> ${INSTALL_DIR}/${BINARY}"
echo ""

if command -v "$BINARY" &>/dev/null; then
    echo "Verified: $(which "$BINARY")"
else
    echo "Warning: ${INSTALL_DIR} is not on your PATH."
    echo "Add this to your shell profile:"
    echo "  export PATH=\"\${HOME}/.local/bin:\${PATH}\""
fi
