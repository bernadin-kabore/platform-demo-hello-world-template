package main

// Process bootstrap: logging, telemetry, routing and the listener. Excluded
// from the coverage denominator along with tracing.go -- see handlers.go.

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	ctx := context.Background()

	initLogging()
	shutdown := initTelemetry(ctx)
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(c); err != nil {
			slog.Error("telemetry shutdown failed", "error", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", healthzHandler)

	// otelhttp emits both spans AND the standard HTTP server metrics
	// (http.server.request.duration, http.server.active_requests) against
	// the global MeterProvider set in initTelemetry. There is no hand-rolled
	// counter and no /metrics endpoint: metrics leave over OTLP like every
	// other signal, so the application depends on OpenTelemetry alone.
	handler := otelhttp.NewHandler(mux, serviceName)

	slog.Info("starting http server", "service", serviceName, "port", 8080)
	if err := http.ListenAndServe(":8080", handler); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}
