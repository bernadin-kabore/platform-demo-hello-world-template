package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const serviceName = "${{ values.name }}"

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Logging with the request context is what attaches trace_id/span_id —
	// slog.Info(...) without a context would produce an uncorrelated line.
	slog.InfoContext(r.Context(), "handled hello request", "route", "/")

	w.Header().Set("Content-Type", "application/json")
	// The error is checked because the platform's golangci-lint runs errcheck,
	// and because a write that fails here is worth a log line: the status has
	// already been sent, so there is nothing left to tell the client.
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from " + serviceName + "!",
		"version": envOr("APP_VERSION", "dev"),
	}); err != nil {
		slog.ErrorContext(r.Context(), "failed to write response", "error", err)
	}
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("ok")); err != nil {
		slog.ErrorContext(r.Context(), "failed to write health response", "error", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

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
