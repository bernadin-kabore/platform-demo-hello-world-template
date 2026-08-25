package main

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Go has no bytecode-weaving auto-instrumentation agent like Java's, so
// wiring the SDK here is the closest equivalent to "zero code in main.go":
// every handler wrapped in otelhttp.NewHandler (see main.go) is traced
// without calling the OTel API directly.
func initTracing(ctx context.Context) func() {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector-gateway.observability.svc.cluster.local:4317"
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(endpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	res, _ := resource.New(ctx, resource.WithAttributes(
		semconv.ServiceName("${{ values.name }}"),
	))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	return func() { _ = tp.Shutdown(ctx) }
}
