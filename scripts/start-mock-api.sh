#!/usr/bin/env bash
# Starts the mock master data REST API used by the shipment enricher demo.
set -euo pipefail
cd "$(dirname "$0")/../shipment-enricher/mock-api"
docker build -q -t northwind-master-data-mock:1.0 . >/dev/null
docker rm -f northwind-mock >/dev/null 2>&1 || true
docker run -d --name northwind-mock --restart unless-stopped -p 127.0.0.1:8099:8099 northwind-master-data-mock:1.0 >/dev/null
until curl -sf -o /dev/null http://127.0.0.1:8099/health; do sleep 1; done
echo "Mock master data API: http://127.0.0.1:8099 (customers, carriers, openapi.json)"
