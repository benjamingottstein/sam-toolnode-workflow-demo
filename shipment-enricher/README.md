# Northwind Shipment Enricher

Event enrichment demo for Solace Agent Mesh desktop, built like the SAP sanitizer and trade correlator demos:
a workflow of deterministic tool steps, no AI model involved, so a run takes well under a second.

## What happens

1. A transport system publishes a bare shipment event on
   `northwind/tms/shipment/created/v1/<site>/<shipment>` — shipment number, customer number, carrier code, route.
2. The **ShipmentEvents** entrypoint picks it up from the Solace Cloud broker and starts the workflow.
3. **ShipmentEnricher** runs four steps:
   - `validate_input` — the event must match the agreed format
   - `get_customer` — REST call for customer master data
   - `get_carrier` — REST call for carrier master data
   - `publish_enriched` — publishes the enriched event on
     `northwind/tms/shipment/enriched/v1/<site>/<shipment>`
4. Anything that fails goes to `northwind/tms/shipment/enrichment-failed/v1` with the reason.

The enriched event keeps the original content and adds `data.customer` (name, country, service level, incoterms,
payment terms, credit status, account manager) and `data.carrier` (name, mode, code, on-time percentage, CO2 per
tonne-kilometre, dispatch contact).

## Parts

| Part | Where |
|---|---|
| Workflow | `workflows/ShipmentEnricher.yaml` |
| REST connection | `connectors/northwind-master-data.yaml` + `openapi/northwind-master-data.json` |
| Event entrypoint | `entrypoints/ShipmentEvents.yaml` |
| Mock REST API | `mock-api/` — runs as the container `northwind-mock` on `http://127.0.0.1:8099` |
| Web page | `console/index.html` — served by the container `northwind-console` on `http://127.0.0.1:8098` |
| Example event | `samples/shipment-created.json` |

The workflow reuses the tool package `sap-event-tools` (already on the instance) for `validate_json` and
`publish_event`. Broker settings come from the SAP demo's `.env`; this project's `.env` only holds the two
addresses of the mock API.

## Running it

Both containers restart by themselves. If they are ever gone:

```bash
docker start northwind-mock northwind-console
```

Open the page, press **Connect**, pick a scenario, press **Publish event**:

- **Road shipment** / **Ocean shipment** — enriched event comes back, the added fields are highlighted
- **Unknown customer** — the REST API answers 404, the run stops, failure event with the reason
- **Breaks the format** — validation rejects the event before any REST call

## Deploying changes

```bash
cd ~/Documents/Codex/2026-09-09/ich/outputs/SAM-workshop/shipment-enricher
sam config apply -m manifest.yaml --env-file .env
```

Test without the broker (runs the workflow directly):

```bash
sam task send --url http://127.0.0.1:8800 --insecure -a ShipmentEnricher --data @samples/shipment-created.json "Enrich this shipment event"
```

## Things worth knowing

- The REST spec travels **as a file** (`specification_file`). The platform refuses to fetch a spec from
  `127.0.0.1`, and a workflow step bound to a REST connection needs the operation names at deploy time.
- The platform renames REST operations to lower case with underscores: `getCustomer` becomes `get_customer`
  and its parameter `customerId` becomes `customer_id`. The workflow uses those names.
- A workflow is only redeployed when its file actually changes; bump `version:` to force it.
- Measured end to end via the broker: enrichment about 180 ms, rejections 70–130 ms.
