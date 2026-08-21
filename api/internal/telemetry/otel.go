package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.16.0"
)

// Config holds OTel configuration.
type Config struct {
	ServiceName    string // e.g. "filora-api"
	ServiceVersion string
	Environment    string // development, production
	Endpoint       string // Axiom OTLP endpoint: "api.axiom.co"
	Token          string // Axiom API token
	Dataset        string // Axiom dataset name
}

// Shutdown is returned by Init and should be called on application exit.
type Shutdown func(ctx context.Context) error

// Init configures OpenTelemetry with OTLP/HTTP exporters pointed at Axiom.
// Returns a shutdown function to flush pending telemetry.
func Init(ctx context.Context, cfg Config) (Shutdown, error) {
	if cfg.Endpoint == "" || cfg.Token == "" {
		// OTel disabled — return no-op shutdown
		return func(_ context.Context) error { return nil }, nil
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	// --- Trace exporter ---
	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization":   "Bearer " + cfg.Token,
			"X-Axiom-Dataset": cfg.Dataset,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(traceExp, trace.WithBatchTimeout(5*time.Second)),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	// --- Metric exporter ---
	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		otlpmetrichttp.WithHeaders(map[string]string{
			"Authorization":   "Bearer " + cfg.Token,
			"X-Axiom-Dataset": cfg.Dataset,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(30*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// --- Propagator ---
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Shutdown flushes both providers
	shutdown := func(ctx context.Context) error {
		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("otel shutdown errors: %v", errs)
		}
		return nil
	}

	return shutdown, nil
}
