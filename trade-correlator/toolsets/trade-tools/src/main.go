package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	sdk "github.com/SolaceDev/solace-agent-mesh-go/pkg/samtoolsdk"
	"github.com/jackc/pgx/v5"
)

type SQLParams struct {
	Statement string         `json:"statement" desc:"One SQL statement; reference arguments as @name, e.g. @event::jsonb"`
	Args      map[string]any `json:"args,omitempty" desc:"Named arguments for the statement, e.g. {event: {...}}"`
}

// runSQL executes one statement in its own transaction (autocommit) and
// returns the affected/returned row count plus any rows it returns, so a
// workflow can keep its data logic as visible SQL in the YAML.
func runSQL(ctx context.Context, p SQLParams, tc *sdk.ToolContext) (*sdk.Result, error) {
	dsn := tc.GetConfigString("dsn", "")
	if dsn == "" {
		return sdk.Error("sql is not configured: set dsn on the trade-tools toolset"), nil
	}
	if strings.TrimSpace(p.Statement) == "" {
		return sdk.Error("statement must not be empty"), nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return sdk.Error("connect to postgres: " + err.Error()), nil
	}
	defer conn.Close(context.Background())

	rows, err := conn.Query(ctx, p.Statement, pgx.NamedArgs(p.Args))
	if err != nil {
		return sdk.Error("sql: " + err.Error()), nil
	}
	result, err := pgx.CollectRows(rows, pgx.RowToMap)
	if err != nil {
		return sdk.Error("sql: " + err.Error()), nil
	}
	count := rows.CommandTag().RowsAffected()
	var first any
	if len(result) > 0 {
		first = result[0]
	}
	tc.SendStatus(fmt.Sprintf("%s: %d row(s)", rows.CommandTag().String(), count))
	return sdk.OK(fmt.Sprintf("%s, %d row(s).", rows.CommandTag().String(), count),
		sdk.WithData(map[string]any{"rowCount": count, "rows": result, "row": first}),
	), nil
}

type PublishEventParams struct {
	Topic string         `json:"topic" desc:"Solace topic to publish to, e.g. trade/settlement/ready/v1/T-123"`
	Event map[string]any `json:"event" desc:"The JSON object to publish as the message body"`
}

// publishEvent uses the broker's REST messaging API: the tool runtime has no
// Solace transport (the Event Mesh connector panics there).
func publishEvent(ctx context.Context, p PublishEventParams, tc *sdk.ToolContext) (*sdk.Result, error) {
	restURL := strings.TrimRight(tc.GetConfigString("rest_url", ""), "/")
	user := tc.GetConfigString("username", "")
	pass := tc.GetConfigString("password", "")
	if restURL == "" || user == "" || pass == "" {
		return sdk.Error("publish_event is not configured: set rest_url, username and password on the trade-tools toolset"), nil
	}
	if p.Topic == "" || strings.ContainsAny(p.Topic, "*>") {
		return sdk.Error("topic must be a concrete topic without wildcards"), nil
	}
	levels := strings.Split(p.Topic, "/")
	for i, l := range levels {
		levels[i] = url.PathEscape(l)
	}
	body, err := json.Marshal(p.Event)
	if err != nil {
		return sdk.Error("encode event: " + err.Error()), nil
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, restURL+"/TOPIC/"+strings.Join(levels, "/"), bytes.NewReader(body))
	if err != nil {
		return sdk.Error("build request: " + err.Error()), nil
	}
	req.SetBasicAuth(user, pass)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Solace-Delivery-Mode", "persistent")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return sdk.Error("publish failed: " + err.Error()), nil
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return sdk.Error(fmt.Sprintf("broker rejected the publish: HTTP %d %s", resp.StatusCode, strings.TrimSpace(string(msg)))), nil
	}
	tc.SendStatus("published to " + p.Topic)
	return sdk.OK("Published to "+p.Topic+".",
		sdk.WithData(map[string]any{"topic": p.Topic, "bytes": len(body)}),
	), nil
}

func main() {
	sdk.Run(
		sdk.NewTool("sql",
			"Executes one SQL statement against the toolset's Postgres database with named @arguments and returns rowCount, rows and the first row.",
			runSQL,
			sdk.WithConfigSchema(
				sdk.ConfigSchemaField{Key: "dsn", Type: "string", Required: true, Secret: true, Description: "Postgres connection string"},
			),
		),
		sdk.NewTool("publish_event",
			"Publishes a JSON object to a Solace topic through the broker's REST messaging API.",
			publishEvent,
			sdk.WithConfigSchema(
				sdk.ConfigSchemaField{Key: "rest_url", Type: "string", Required: true, Description: "Broker REST messaging URL, e.g. https://<host>:9443"},
				sdk.ConfigSchemaField{Key: "username", Type: "string", Required: true, Description: "Broker client username"},
				sdk.ConfigSchemaField{Key: "password", Type: "string", Required: true, Secret: true, Description: "Broker client password"},
			),
		),
	)
}
