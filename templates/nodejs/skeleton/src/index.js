const express = require("express");
const pino = require("pino");
const { trace } = require("@opentelemetry/api");

const app = express();
const port = process.env.PORT || 8080;
const serviceName = "${{ values.name }}";

// Structured JSON to stdout. The otel-collector agent tails the container log
// file and promotes trace_id/span_id into first-class log-record fields, which
// is what makes the log -> trace pivot work in Grafana. Logs are collected
// rather than pushed so that a crash still produces readable output — see
// platform-demo-gitops/docs/observability/README.md, decision 1.
const logger = pino({
  level: process.env.LOG_LEVEL || "info",
  messageKey: "message",
  timestamp: () => `,"timestamp":"${new Date().toISOString()}"`,
  formatters: {
    // Field names here must match the collector's json_parser configuration.
    level: (label) => ({ severity: label.toUpperCase() }),
  },
  // Runs per log call, so it picks up whichever span is active at that moment.
  mixin() {
    const span = trace.getActiveSpan();
    if (!span) return {};
    const ctx = span.spanContext();
    return { trace_id: ctx.traceId, span_id: ctx.spanId };
  },
});

app.get("/", (req, res) => {
  logger.info({ route: "/" }, "handled hello request");
  res.json({ message: `Hello from ${serviceName}!`, version: process.env.APP_VERSION || "dev" });
});

// Liveness: process is up. Readiness: process is up AND ready for traffic —
// identical here since this service has no external dependencies to wait
// on, but kept separate so the Rollout/Kyverno probe requirements are real.
app.get("/healthz", (req, res) => res.status(200).send("ok"));
app.get("/readyz", (req, res) => res.status(200).send("ok"));

if (require.main === module) {
  app.listen(port, () => {
    logger.info({ port }, `${serviceName} listening`);
  });
}

module.exports = app;
