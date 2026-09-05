import json
import logging
import os
import sys
from datetime import datetime, timezone

from fastapi import FastAPI
from fastapi.responses import PlainTextResponse

SERVICE_NAME = "${{ values.name }}"


class JsonFormatter(logging.Formatter):
    """Emits structured JSON on stdout for the otel-collector agent to tail.

    OTEL_PYTHON_LOG_CORRELATION=true (set in the Dockerfile) makes the
    OpenTelemetry logging instrumentation attach otelTraceID/otelSpanID to
    every record, which this formatter promotes to the trace_id/span_id
    field names the collector's json_parser expects. That is what makes the
    log -> trace pivot work in Grafana.
    """

    def format(self, record: logging.LogRecord) -> str:
        payload = {
            # logging.formatTime() delegates to time.strftime(), which has no
            # %f directive for sub-second precision, so build this directly.
            "timestamp": datetime.fromtimestamp(record.created, tz=timezone.utc)
            .isoformat(timespec="milliseconds")
            .replace("+00:00", "Z"),
            "severity": record.levelname,
            "message": record.getMessage(),
            "logger": record.name,
        }
        trace_id = getattr(record, "otelTraceID", None)
        span_id = getattr(record, "otelSpanID", None)
        # "0" * 32 is what the instrumentation emits when no span is active;
        # a zero ID is worse than no ID because it looks like a real trace.
        if trace_id and trace_id.strip("0"):
            payload["trace_id"] = trace_id
        if span_id and span_id.strip("0"):
            payload["span_id"] = span_id
        if record.exc_info:
            payload["exception"] = self.formatException(record.exc_info)
        return json.dumps(payload)


def _configure_logging() -> None:
    handler = logging.StreamHandler(sys.stdout)
    handler.setFormatter(JsonFormatter())
    root = logging.getLogger()
    root.handlers = [handler]
    root.setLevel(os.environ.get("LOG_LEVEL", "INFO"))


_configure_logging()
logger = logging.getLogger(SERVICE_NAME)

app = FastAPI(title=SERVICE_NAME)


@app.get("/")
def hello():
    logger.info("handled hello request")
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


# No /metrics endpoint: metrics leave over OTLP via opentelemetry-instrument,
# so this service depends on OpenTelemetry alone and never on Prometheus.
