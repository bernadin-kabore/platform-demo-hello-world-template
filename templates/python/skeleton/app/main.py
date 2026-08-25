import os

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse
from prometheus_client import Counter, generate_latest, CONTENT_TYPE_LATEST

SERVICE_NAME = "${{ values.name }}"

app = FastAPI(title=SERVICE_NAME)

http_requests_total = Counter(
    "http_requests_total", "Total HTTP requests", ["method", "route", "status_code"]
)


@app.middleware("http")
async def count_requests(request, call_next):
    response = await call_next(request)
    http_requests_total.labels(request.method, request.url.path, response.status_code).inc()
    return response


@app.get("/")
def hello():
    return {"message": f"Hello from {SERVICE_NAME}!", "version": os.environ.get("APP_VERSION", "dev")}


# Liveness/readiness kept distinct even though they're identical today — this
# service has no external dependencies to wait on, but Kyverno's
# require-probes policy and the Rollout's health checks expect both.
@app.get("/healthz", response_class=PlainTextResponse)
def healthz():
    return "ok"


@app.get("/readyz", response_class=PlainTextResponse)
def readyz():
    return "ok"


@app.get("/metrics")
def metrics():
    return PlainTextResponse(generate_latest(), media_type=CONTENT_TYPE_LATEST)
