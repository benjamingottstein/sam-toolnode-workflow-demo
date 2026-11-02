# Sales Order Sanitizer

An event-driven workflow that removes personal data from a sales order event before it travels on. Four
deterministic steps, no AI model involved.

```
sales order created  →  validate_input  →  strip_pii  →  validate_output  →  republish
   (event mesh)          input contract     2 fields      output contract     sanitized event
```

- **In:** `sap/car/salesorder/created/v1/<sales org>/<order>`
- **Out:** `sap/car/salesorder/sanitized/v1/<sales org>/<order>`
- **Rejected:** `sap/car/salesorder/rejected/v1` with the reason

The output contract is an allowlist: if the e-mail address or loyalty card number were still present, the event
would fail validation instead of being republished. That is what makes the removal verifiable rather than a promise.

## Parts

| Part | Where |
|---|---|
| Workflow | `workflows/SapSalesOrderSanitizer.yaml` |
| Tool package (Go) | `toolsets/sap-event-tools/` — `validate_json`, `strip_fields`, `publish_event` |
| Event entrypoint | `entrypoints/SapCarEvents.yaml` |
| Webhook entrypoint | `entrypoints/SapCarWebhook.yaml` (same workflow over HTTP) |
| Portal | `console/index.html` |
| Example event | `samples/salesorder-created.json` |

This tool package is also used by the shipment enricher demo, so deploy this one first.

## Run it

```bash
cp .env.example .env   # then enter your broker details
```

Fetch the tool SDK once (it is not in the repository):

```bash
sam toolset sync
```

```bash
sam config apply --target desktop --env-file .env -m manifest.yaml
```

Wait about 20 seconds after an apply before sending the first event: the entrypoint restarts and has to find the
workflow again.

```bash
../scripts/serve-portals.sh
```

Open http://127.0.0.1:8790/portal/, choose **Sales order sanitizer**, press **Connect**, then **Publish event**.
The scenarios cover a valid order, one that breaks the input contract, and one with an unexpected field.
