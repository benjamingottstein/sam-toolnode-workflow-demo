#!/usr/bin/env bash
# Serves the three web portals locally.
#   http://127.0.0.1:8790/portal/   - start page with links to all three
set -euo pipefail
cd "$(dirname "$0")/.."
PORT=${1:-8790}
echo "Portals:      http://127.0.0.1:${PORT}/portal/"
echo "Press Ctrl+C to stop."
python3 -m http.server "$PORT" --bind 127.0.0.1
