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
ENV PATH="/venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    OTEL_SERVICE_NAME="${{ values.name }}" \
    OTEL_EXPORTER_OTLP_ENDPOINT="http://otel-collector-gateway.observability.svc.cluster.local:4317" \
    OTEL_TRACES_EXPORTER=otlp \
    OTEL_METRICS_EXPORTER=none

USER appuser
EXPOSE 8080
# opentelemetry-instrument wraps the process for zero-code auto-instrumentation
# of FastAPI/Starlette/HTTP — no tracing code in app/main.py.
ENTRYPOINT ["opentelemetry-instrument", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
