package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

// The request handlers -- the part of a scaffolded service a team actually
// writes, and the only file the coverage gate measures. main.go and tracing.go
// are platform bootstrap: every service inherits them unchanged, nobody edits
// them, and counting them made a fully tested service score 11%.

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
