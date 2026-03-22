#!/bin/sh
# Install rift as a Docker CLI plugin.
# Usage: ./scripts/install-docker-plugin.sh [path-to-rift-binary]
#
# After installation, use: docker rift nginx:1.24 nginx:1.25

set -e

BINARY="${1:-rift}"
PLUGIN_DIR="${HOME}/.docker/cli-plugins"

if [ ! -f "$BINARY" ]; then
  echo "Error: binary '$BINARY' not found."
  echo "Usage: $0 [path-to-rift-binary]"
  exit 1
fi

mkdir -p "$PLUGIN_DIR"
cp "$BINARY" "$PLUGIN_DIR/docker-rift"
chmod +x "$PLUGIN_DIR/docker-rift"

echo "Installed rift as Docker CLI plugin."
echo "Usage: docker rift <image1> <image2>"
