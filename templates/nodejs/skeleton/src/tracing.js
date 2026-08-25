// Loaded before anything else via `node --require ./src/tracing.js src/index.js`
// (see the Dockerfile CMD). Auto-instruments Express/HTTP and exports spans
// via OTLP/gRPC to the platform's otel-collector gateway — no manual span
// creation needed for basic request tracing.
const { NodeSDK } = require("@opentelemetry/sdk-node");
const { getNodeAutoInstrumentations } = require("@opentelemetry/auto-instrumentations-node");
const { OTLPTraceExporter } = require("@opentelemetry/exporter-trace-otlp-grpc");

const sdk = new NodeSDK({
  serviceName: process.env.OTEL_SERVICE_NAME || "${{ values.name }}",
  traceExporter: new OTLPTraceExporter({
    url: process.env.OTEL_EXPORTER_OTLP_ENDPOINT || "http://otel-collector-gateway.observability.svc.cluster.local:4317",
  }),
  instrumentations: [getNodeAutoInstrumentations()],
});

sdk.start();

process.on("SIGTERM", () => sdk.shutdown().finally(() => process.exit(0)));
