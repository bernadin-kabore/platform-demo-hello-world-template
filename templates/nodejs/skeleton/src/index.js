const express = require("express");
const client = require("prom-client");

const app = express();
const port = process.env.PORT || 8080;
const serviceName = "${{ values.name }}";

const register = new client.Registry();
client.collectDefaultMetrics({ register });

const httpRequestsTotal = new client.Counter({
  name: "http_requests_total",
  help: "Total HTTP requests",
  labelNames: ["method", "route", "status_code"],
  registers: [register],
});

app.use((req, res, next) => {
  res.on("finish", () => {
    httpRequestsTotal.inc({ method: req.method, route: req.path, status_code: res.statusCode });
  });
  next();
});

app.get("/", (req, res) => {
  res.json({ message: `Hello from ${serviceName}!`, version: process.env.APP_VERSION || "dev" });
});

// Liveness: process is up. Readiness: process is up AND ready for traffic —
// identical here since this service has no external dependencies to wait
// on, but kept separate so the Rollout/Kyverno probe requirements are real.
app.get("/healthz", (req, res) => res.status(200).send("ok"));
app.get("/readyz", (req, res) => res.status(200).send("ok"));

app.get("/metrics", async (req, res) => {
  res.set("Content-Type", register.contentType);
  res.end(await register.metrics());
});

if (require.main === module) {
  app.listen(port, () => {
    console.log(`${serviceName} listening on :${port}`);
  });
}

module.exports = app;
