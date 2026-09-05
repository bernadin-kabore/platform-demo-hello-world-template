# syntax=docker/dockerfile:1
FROM python:3.12.4-slim AS builder
WORKDIR /app
RUN python -m venv /venv
ENV PATH="/venv/bin:$PATH"
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt \
    && opentelemetry-bootstrap -a install

FROM python:3.12.4-slim AS runtime
RUN useradd --uid 65532 --create-home --shell /usr/sbin/nologin appuser
COPY --from=builder /venv /venv
WORKDIR /app
COPY app ./app
# OTEL_SERVICE_NAME, the OTLP endpoint and resource attributes are injected by
# the Helm chart at deploy time rather than baked in here, so the same image
# promotes across environments unchanged.
# OTEL_PYTHON_LOG_CORRELATION makes the instrumentation attach
# otelTraceID/otelSpanID to every log record, which app/main.py's JsonFormatter
# emits as trace_id/span_id for the collector to promote.
ENV PATH="/venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=otlp \
    OTEL_LOGS_EXPORTER=none \
    OTEL_PYTHON_LOG_CORRELATION=true

USER appuser
EXPOSE 8080
# opentelemetry-instrument wraps the process for zero-code auto-instrumentation
# of FastAPI/Starlette/HTTP - no tracing code in app/main.py.
ENTRYPOINT ["opentelemetry-instrument", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
