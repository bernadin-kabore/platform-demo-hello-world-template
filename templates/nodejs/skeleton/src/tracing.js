// Loaded before anything else via `node --require ./src/tracing.js src/index.js`
// (see the Dockerfile ENTRYPOINT). Auto-instruments Express/HTTP and exports
// traces and metrics over OTLP to the platform's collector gateway.
//
// Every endpoint and resource attribute comes from OTEL_* environment
// variables injected by the Helm chart — the SDK reads them natively, so
// nothing about the environment is baked into the image.
const { NodeSDK } = require("@opentelemetry/sdk-node");
const { getNodeAutoInstrumentations } = require("@opentelemetry/auto-instrumentations-node");
const { OTLPTraceExporter } = require("@opentelemetry/exporter-trace-otlp-grpc");
const { OTLPMetricExporter } = require("@opentelemetry/exporter-metrics-otlp-grpc");
const { PeriodicExportingMetricReader } = require("@opentelemetry/sdk-metrics");

const sdk = new NodeSDK({
  traceExporter: new OTLPTraceExporter(),
  // Auto-instrumentation emits the standard HTTP server metrics against this
  // reader, so there is no hand-rolled counter and no /metrics endpoint. The
  // application depends on OpenTelemetry alone.
  metricReader: new PeriodicExportingMetricReader({
    exporter: new OTLPMetricExporter(),
    exportIntervalMillis: 30000,
  }),
  instrumentations: [getNodeAutoInstrumentations()],
});

sdk.start();

process.on("SIGTERM", () => sdk.shutdown().finally(() => process.exit(0)));
