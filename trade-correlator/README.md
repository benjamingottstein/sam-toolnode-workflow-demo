# Trade Settlement Correlator

Stateful, LLM-free SAM workflow: three systems report on the same trade, in any
order. Only when all three events for a `tradeId` have arrived does the workflow
create a fourth event, `TradeReadyForSettlement`.

| # | Event | Source | Topic |
|---|---|---|---|
| 1 | `TradeExecuted` | OMS | `trade/executed/v1/<tradeId>` |
| 2 | `TradeConfirmed` | Confirmation platform | `trade/confirmed/v1/<tradeId>` |
| 3 | `TradeAllocated` | Middle office | `trade/allocated/v1/<tradeId>` |
| 4 | `TradeReadyForSettlement` (published by the workflow) | SAM | `trade/settlement/ready/v1/<tradeId>` |

```
store_event (SQL) → check_complete (SQL) → all_events_in? ──rowCount 1──→ publish_ready (4th event)
                                                         └─otherwise: run ends
```

All data logic is SQL in the workflow YAML; the generic `trade-tools__sql`
tool only executes it (named `@arguments`, returns `rowCount`, `rows`, `row`).

- `store_event`: `INSERT … ON CONFLICT DO NOTHING` - duplicates and unknown
  event types are not stored.
- `check_complete`: `UPDATE trades SET ready_at = now() WHERE ready_at IS NULL
  AND (count of events) = 3 RETURNING …` - atomic, so even when all three
  events arrive at once exactly one run gets a row back. That row is event 4,
  built with `jsonb_build_object`.
- `all_events_in` is a switch on `rowCount`; `publish_ready` publishes event 4
  via the broker's REST API.

## Setup

```bash
docker run -d --name trade-correlator-pg --restart unless-stopped \
  -e POSTGRES_USER=trades -e POSTGRES_PASSWORD=trades -e POSTGRES_DB=trades \
  -p 127.0.0.1:55432:5432 postgres:18.0-trixie
docker exec -i trade-correlator-pg psql -U trades -d trades < db/schema.sql
Fetch the tool SDK once (it is not in the repository):

```bash
sam toolset sync
```

sam config apply --target desktop --env-file .env -m manifest.yaml
```

Wait ~20 s after an apply before sending events (the entrypoint gateway
restarts and has to rediscover the workflow).

## Demo

Console (connects to the broker over WSS, listens on `trade/>`):

```bash
cd console && python3 -m http.server 8791 --bind 127.0.0.1
```

Open http://127.0.0.1:8791, click **Connect**, then send the three events one
by one (event 4 appears after the third), or use **Send all 3 at once** to show
that concurrent events still produce exactly one event 4.

Or from the shell:

```bash
./samples/send.sh executed  T-1001
./samples/send.sh confirmed T-1001
./samples/send.sh allocated T-1001   # third event → TradeReadyForSettlement
```

Watch `trade/>` on the broker (e.g. Try-Me) and the state in Postgres:

```bash
docker exec trade-correlator-pg psql -U trades -d trades \
  -c "select trade_id, event_type, received_at from trade_events order by received_at"
```
