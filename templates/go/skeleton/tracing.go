package main

import (
	"context"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Go has no bytecode-weaving auto-instrumentation agent like Java's, so the
// SDK is wired up here. Everything below reads its configuration from
// OTEL_* environment variables injected by the Helm chart — nothing about
// the environment is compiled into the binary, so one image promotes across
// dev/stage/prod unchanged.
//
// resource.Default() already reads OTEL_SERVICE_NAME and
// OTEL_RESOURCE_ATTRIBUTES, and Kubernetes-derived attributes (pod,
// namespace, node, deployment) are added by the collector's k8sattributes
// processor rather than duplicated here.
func initTelemetry(ctx context.Context) func(context.Context) error {
	res, err := resource.Merge(resource.Default(), resource.Empty())
	if err != nil {
		res = resource.Default()
	}

	traceExp, err := otlptracegrpc.New(ctx)
	if err != nil {
		panic(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// W3C trace context, so trace IDs survive hops between services.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	metricExp, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		panic(err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return func(c context.Context) error {
		if err := tp.Shutdown(c); err != nil {
			return err
		}
		return mp.Shutdown(c)
	}
}

// traceHandler wraps a slog.Handler and stamps trace_id/span_id onto every
// record emitted with a context carrying an active span. This is what makes
// the log -> trace pivot work: the collector's filelog receiver promotes
// these two fields into first-class log-record IDs, and Grafana uses them to
// jump from a log line straight to the trace.
type traceHandler struct{ slog.Handler }

func (h traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return traceHandler{h.Handler.WithAttrs(attrs)}
}

func (h traceHandler) WithGroup(name string) slog.Handler {
	return traceHandler{h.Handler.WithGroup(name)}
}

// initLogging installs a JSON logger on stdout. Logs are collected from the
// container's stdout by the otel-collector agent rather than pushed by this
// process, so a panic or an OOM kill still produces readable output — see
// decision 1 in platform-demo-gitops/docs/observability/README.md.
func initLogging() {
	base := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Match the field names the collector's json_parser expects.
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.LevelKey:
				a.Key = "severity"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	})
	slog.SetDefault(slog.New(traceHandler{base}))
}
