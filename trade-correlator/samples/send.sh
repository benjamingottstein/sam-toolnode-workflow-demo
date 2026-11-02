#!/usr/bin/env bash
# Publishes one trade lifecycle event to the mesh.
#   ./samples/send.sh executed  T-123
#   ./samples/send.sh confirmed T-123
#   ./samples/send.sh allocated T-123   <- the third one triggers TradeReadyForSettlement
set -euo pipefail
cd "$(dirname "$0")/.."
set -a; . ./.env; set +a
step=${1:?executed|confirmed|allocated}; id=${2:?tradeId}
now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
case $step in
  executed)  body="{\"tradeId\":\"$id\",\"eventType\":\"TradeExecuted\",\"source\":\"OMS\",\"time\":\"$now\",\"isin\":\"DE0007164600\",\"instrument\":\"SAP SE\",\"side\":\"BUY\",\"quantity\":1200,\"price\":231.45,\"currency\":\"EUR\",\"venue\":\"XETR\",\"trader\":\"FX-DESK-07\"}" ;;
  confirmed) body="{\"tradeId\":\"$id\",\"eventType\":\"TradeConfirmed\",\"source\":\"ConfirmationPlatform\",\"time\":\"$now\",\"counterparty\":\"Deutsche Bank AG\",\"counterpartyLei\":\"7LTWFZYICNSX8D621K86\",\"confirmationRef\":\"CNF-$id\"}" ;;
  allocated) body="{\"tradeId\":\"$id\",\"eventType\":\"TradeAllocated\",\"source\":\"MiddleOffice\",\"time\":\"$now\",\"account\":\"FUND-EU-EQ-01\",\"custodian\":\"Clearstream Banking\",\"settlementDate\":\"$(date -u -v+2d +%Y-%m-%d 2>/dev/null || date -u -d '+2 days' +%Y-%m-%d)\"}" ;;
  *) echo "unknown step $step" >&2; exit 1 ;;
esac
curl -sf -o /dev/null -u "$TRADE_BROKER_USERNAME:$TRADE_BROKER_PASSWORD" \
  -H 'Content-Type: application/json' -H 'Solace-Delivery-Mode: persistent' \
  --data "$body" "$TRADE_BROKER_REST_URL/TOPIC/trade/$step/v1/$id"
echo "published trade/$step/v1/$id"
