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
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type ValidateJSONParams struct {
	Document map[string]any `json:"document" desc:"The JSON object to validate"`
	Schema   map[string]any `json:"schema" desc:"JSON Schema (draft 2020-12) the document must satisfy"`
	Label    *string        `json:"label,omitempty" desc:"Name used in messages, e.g. input or output"`
}

func validateJSON(ctx context.Context, p ValidateJSONParams, tc *sdk.ToolContext) (*sdk.Result, error) {
	_ = ctx
	label := "document"
	if p.Label != nil && *p.Label != "" {
		label = *p.Label
	}
	schemaID := "urn:sap-event-tools:" + label
	c := jsonschema.NewCompiler()
	if err := c.AddResource(schemaID, p.Schema); err != nil {
		return sdk.Error("invalid schema: " + err.Error()), nil
	}
	schema, err := c.Compile(schemaID)
	if err != nil {
		return sdk.Error("invalid schema: " + err.Error()), nil
	}
	// A schema violation is a normal result, not a tool error, so a workflow
	// switch node can branch on status instead of the workflow failing.
	if err := schema.Validate(p.Document); err != nil {
		tc.SendStatus(label + " failed schema validation")
		reason := fmt.Sprintf("%s does not match its schema: %v", label, err)
		return sdk.OK(reason,
			sdk.WithData(map[string]any{"status": "invalid", "reason": reason, "document": p.Document}),
		), nil
	}
	tc.SendStatus(label + " matches schema")
	return sdk.OK(label+" matches its schema.",
		sdk.WithData(map[string]any{"status": "valid", "reason": "", "document": p.Document}),
	), nil
}

type StripFieldsParams struct {
	Event  map[string]any `json:"event" desc:"The event JSON object to sanitize"`
	Fields []string       `json:"fields" desc:"Dot-separated paths of the properties to remove, e.g. data.CustomerEmailAddress"`
}

// removePath deletes a dot-separated path from nested JSON objects and
// reports whether anything was removed.
func removePath(obj map[string]any, path string) bool {
	parts := strings.Split(path, ".")
	for _, key := range parts[:len(parts)-1] {
		next, ok := obj[key].(map[string]any)
		if !ok {
			return false
		}
		obj = next
	}
	last := parts[len(parts)-1]
	if _, ok := obj[last]; !ok {
		return false
	}
	delete(obj, last)
	return true
}

func stripFields(ctx context.Context, p StripFieldsParams, tc *sdk.ToolContext) (*sdk.Result, error) {
	_ = ctx
	removed := []string{}
	missing := []string{}
	for _, path := range p.Fields {
		if removePath(p.Event, path) {
			removed = append(removed, path)
		} else {
			missing = append(missing, path)
		}
	}
	tc.SendStatus(fmt.Sprintf("removed %d field(s)", len(removed)))
	return sdk.OK(fmt.Sprintf("Removed %d of %d field(s).", len(removed), len(p.Fields)),
		sdk.WithData(map[string]any{
			"sanitizedEvent": p.Event,
			"removedFields":  removed,
			"missingFields":  missing,
		}),
	), nil
}

type PublishEventParams struct {
	Topic string         `json:"topic" desc:"Solace topic to publish to, e.g. sap/car/salesorder/sanitized/v1/1010/0000481516"`
	Event map[string]any `json:"event" desc:"The JSON object to publish as the message body"`
}

// publishEvent uses the broker's REST messaging API rather than a Solace
// client library: the tool runtime has no Solace transport (and the Event
// Mesh connector panics there), while plain HTTPS works everywhere.
func publishEvent(ctx context.Context, p PublishEventParams, tc *sdk.ToolContext) (*sdk.Result, error) {
	restURL := strings.TrimRight(tc.GetConfigString("rest_url", ""), "/")
	user := tc.GetConfigString("username", "")
	pass := tc.GetConfigString("password", "")
	if restURL == "" || user == "" || pass == "" {
		return sdk.Error("publish_event is not configured: set rest_url, username and password on the sap-event-tools toolset"), nil
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
		sdk.NewTool("validate_json",
			"Validates a JSON object against a JSON Schema and fails with the violations if it does not match.",
			validateJSON,
		),
		sdk.NewTool("strip_fields",
			"Removes the given dot-path properties from a JSON event and returns the sanitized event.",
			stripFields,
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
