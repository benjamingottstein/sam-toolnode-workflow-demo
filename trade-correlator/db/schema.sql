-- One row per (trade, event type): a redelivered or duplicate event is a no-op.
CREATE TABLE IF NOT EXISTS trade_events (
  trade_id    text        NOT NULL,
  event_type  text        NOT NULL,
  payload     jsonb       NOT NULL,
  received_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (trade_id, event_type)
);

-- One row per trade; ready_at is set exactly once, by whichever event
-- completes the set, so the follow-up event is published exactly once.
CREATE TABLE IF NOT EXISTS trades (
  trade_id   text        PRIMARY KEY,
  first_seen timestamptz NOT NULL DEFAULT now(),
  ready_at   timestamptz
);
