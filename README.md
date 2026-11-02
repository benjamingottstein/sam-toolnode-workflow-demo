# Solace Agent Mesh — tool node workflow demos

Three event-driven workflows for Solace Agent Mesh that run **entirely on deterministic tool steps**: no AI model is
involved. Each run takes a fraction of a second, produces the same result every time, and costs nothing per run.

Every demo follows the same shape — an event arrives from a Solace broker, a workflow of tool steps processes it, and
the result is published back onto the mesh as a new event. Each demo ships with a small web portal that publishes the
incoming event and shows the event that comes back.

| Demo | What it does | Shows |
|---|---|---|
| [Sales order sanitizer](sap-car-sanitizer/) | Removes personal data from a sales order event and republishes it | Validation against a contract, before/after comparison |
| [Shipment enricher](shipment-enricher/) | Adds customer and carrier master data from a REST API | Calling an external REST API from a workflow step |
| [Trade settlement correlator](trade-correlator/) | Waits for three events about the same trade, then publishes a fourth | State across events, held in a database |

## What you need

- **Solace Agent Mesh desktop** running (the portals and the `sam` command talk to it on `http://127.0.0.1:8800`)
- **The `sam` command** on your PATH — it ships with the desktop app
- **A Solace broker** — Solace Cloud, a local PubSub+ container, or any existing one. All three demos can share one.
- **Python 3** for the portals, **Docker** for the mock API and the database, **Go** only if you change a tool package

## Getting started

> **One-time step after cloning:** the Go tool packages need the Agent Mesh tool SDK, which is not part of this
> repository. Fetch it from your own installation (no network needed):
>
> ```bash
> sam toolset sync
> ```

```bash
git clone https://github.com/solacese/sam-toolnode-workflow-demo.git && cd sam-toolnode-workflow-demo
```

Each demo keeps its own settings. Copy the example and fill in your broker:

```bash
cp sap-car-sanitizer/.env.example sap-car-sanitizer/.env
```

Deploy a demo into the desktop app:

```bash
cd sap-car-sanitizer && sam config apply --target desktop --env-file .env -m manifest.yaml
```

Start the portals (one server, all three):

```bash
./scripts/serve-portals.sh
```

Open **http://127.0.0.1:8790/portal/**, pick a demo, press **Connect**, enter the same broker details, publish an event.

> After every deploy, wait about 20 seconds before sending the first event. The entrypoint restarts and has to find
> the workflow again; an event sent too early runs into a timeout.

Demo-specific steps are in each demo's own README — the trade correlator needs a Postgres container, the shipment
enricher a mock REST API (`./scripts/start-mock-api.sh`).

## Why tool steps instead of agents

An AI agent is the right tool when a decision needs judgement. These three cases don't: a field is either present or
not, a REST lookup either answers or doesn't, three events are either complete or not. Using tool steps makes the
workflows fast, repeatable and auditable — and every step can be checked against a contract.

Each workflow still runs inside Agent Mesh, with the same event entrypoints, the same deployment, and the same run
history in the user interface as an agent-based workflow.

## Repository layout

```
sap-car-sanitizer/     workflow, Go tool package, event + webhook entrypoint, portal
shipment-enricher/     workflow, REST connection + mock API, event entrypoint, portal
trade-correlator/      workflow, Go tool package, database schema, event entrypoint, portal
portal/                start page linking the three portals
scripts/               serve-portals.sh, start-mock-api.sh
```

Credentials never belong in this repository: every project reads them from its own `.env`, which is git-ignored.
The portals ask for the broker details in the page itself.
