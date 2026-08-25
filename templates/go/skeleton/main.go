package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const serviceName = "${{ values.name }}"

var httpRequestsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{Name: "http_requests_total", Help: "Total HTTP requests"},
	[]string{"method", "route", "status_code"},
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Hello from " + serviceName + "!",
		"version": envOr("APP_VERSION", "dev"),
	})
	httpRequestsTotal.WithLabelValues(r.Method, "/", "200").Inc()
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	ctx := context.Background()
	shutdown := initTracing(ctx)
	defer shutdown()

	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", healthzHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// otelhttp.NewHandler traces every request through the mux without any
	// per-handler span code — the Go equivalent of Node/Python's zero-code
	// auto-instrumentation.
	handler := otelhttp.NewHandler(mux, serviceName)

	log.Printf("%s listening on :8080", serviceName)
	log.Fatal(http.ListenAndServe(":8080", handler))
}
