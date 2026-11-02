"""Mock master-data REST API for the Northwind shipment enrichment demo.

Serves customer and carrier master data plus its own OpenAPI specification.
Standard library only; all data is invented.
"""
import json
import os
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

CUSTOMERS = {
    "C-10045": {"customerId": "C-10045", "name": "Baltic Fresh Foods AB", "country": "SE", "serviceLevel": "gold",
                "incoterms": "DAP", "paymentTerms": "net30", "creditStatus": "ok",
                "accountManager": {"name": "Lena Hartmann", "email": "lena.hartmann@northwind-freight.example"}},
    "C-10233": {"customerId": "C-10233", "name": "Rhein Pharma Logistik GmbH", "country": "DE", "serviceLevel": "platinum",
                "incoterms": "DDP", "paymentTerms": "net14", "creditStatus": "ok",
                "accountManager": {"name": "Marco Pereira", "email": "marco.pereira@northwind-freight.example"}},
    "C-10877": {"customerId": "C-10877", "name": "Costa Verde Produce SL", "country": "ES", "serviceLevel": "silver",
                "incoterms": "FCA", "paymentTerms": "net45", "creditStatus": "on_hold",
                "accountManager": {"name": "Ana Ruiz", "email": "ana.ruiz@northwind-freight.example"}},
}
CARRIERS = {
    "NWX": {"carrierCode": "NWX", "name": "Northwind Express", "mode": "road", "scac": "NWXP",
            "onTimePercent": 94.2, "co2GramsPerTonneKm": 62,
            "dispatchContact": {"phone": "+49 40 555 0113", "email": "dispatch@northwind-express.example"}},
    "BLS": {"carrierCode": "BLS", "name": "Baltic Line Shipping", "mode": "ocean", "scac": "BLSH",
            "onTimePercent": 81.5, "co2GramsPerTonneKm": 14,
            "dispatchContact": {"phone": "+46 31 555 0190", "email": "ops@baltic-line.example"}},
    "AIR7": {"carrierCode": "AIR7", "name": "Skyway Air Cargo", "mode": "air", "scac": "SKYC",
             "onTimePercent": 88.9, "co2GramsPerTonneKm": 540,
             "dispatchContact": {"phone": "+31 20 555 0144", "email": "cargo@skyway-air.example"}},
}

OPENAPI = {
    "openapi": "3.0.3",
    "info": {"title": "Northwind Master Data API (mock)", "version": "1.0.0",
             "description": "Customer and carrier master data used to enrich shipment events."},
    "servers": [{"url": "http://127.0.0.1:8099"}],
    "paths": {
        "/customers/{customerId}": {
            "get": {
                "operationId": "getCustomer",
                "summary": "Customer master data for a customer number",
                "parameters": [{"name": "customerId", "in": "path", "required": True, "schema": {"type": "string"},
                                "description": "Customer number, e.g. C-10045"}],
                "responses": {
                    "200": {"description": "Customer record", "content": {"application/json": {"schema": {
                        "type": "object",
                        "properties": {
                            "customerId": {"type": "string"}, "name": {"type": "string"}, "country": {"type": "string"},
                            "serviceLevel": {"type": "string"}, "incoterms": {"type": "string"},
                            "paymentTerms": {"type": "string"}, "creditStatus": {"type": "string"},
                            "accountManager": {"type": "object", "properties": {
                                "name": {"type": "string"}, "email": {"type": "string"}}},
                        }}}}},
                    "404": {"description": "Unknown customer"},
                },
            }
        },
        "/carriers/{carrierCode}": {
            "get": {
                "operationId": "getCarrier",
                "summary": "Carrier master data for a carrier code",
                "parameters": [{"name": "carrierCode", "in": "path", "required": True, "schema": {"type": "string"},
                                "description": "Carrier code, e.g. NWX"}],
                "responses": {
                    "200": {"description": "Carrier record", "content": {"application/json": {"schema": {
                        "type": "object",
                        "properties": {
                            "carrierCode": {"type": "string"}, "name": {"type": "string"}, "mode": {"type": "string"},
                            "scac": {"type": "string"}, "onTimePercent": {"type": "number"},
                            "co2GramsPerTonneKm": {"type": "number"},
                            "dispatchContact": {"type": "object", "properties": {
                                "phone": {"type": "string"}, "email": {"type": "string"}}},
                        }}}}},
                    "404": {"description": "Unknown carrier"},
                },
            }
        },
        "/health": {"get": {"operationId": "health", "summary": "Health check",
                            "responses": {"200": {"description": "ok"}}}},
    },
}


class Handler(BaseHTTPRequestHandler):
    protocol_version = "HTTP/1.1"

    def _send(self, status, payload):
        body = json.dumps(payload).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = self.path.split("?")[0].rstrip("/") or "/"
        if path in ("/openapi.json", "/openapi"):
            return self._send(200, OPENAPI)
        if path == "/health":
            return self._send(200, {"status": "ok"})
        if path.startswith("/customers/"):
            key = path.rsplit("/", 1)[-1]
            row = CUSTOMERS.get(key)
            return self._send(200, row) if row else self._send(404, {"error": "unknown customer", "customerId": key})
        if path.startswith("/carriers/"):
            key = path.rsplit("/", 1)[-1]
            row = CARRIERS.get(key)
            return self._send(200, row) if row else self._send(404, {"error": "unknown carrier", "carrierCode": key})
        if path == "/":
            return self._send(200, {"service": "Northwind master data (mock)",
                                    "endpoints": ["/customers/{customerId}", "/carriers/{carrierCode}", "/openapi.json"],
                                    "customers": list(CUSTOMERS), "carriers": list(CARRIERS)})
        return self._send(404, {"error": "not found", "path": path})

    def log_message(self, fmt, *args):
        print("%s - %s" % (self.address_string(), fmt % args), flush=True)


if __name__ == "__main__":
    ThreadingHTTPServer(("0.0.0.0", int(os.environ.get("PORT", "8099"))), Handler).serve_forever()
