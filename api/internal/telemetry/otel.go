package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otellog "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.16.0"
)

// Config holds OTel configuration.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Endpoint       string // Axiom: api.axiom.co
	Token          string // Axiom API token
	Dataset        string // Axiom dataset for traces + logs (X-Axiom-Dataset)
	MetricsDataset string // Axiom dataset for metrics (X-Axiom-Metrics-Dataset)
}

// Shutdown is returned by Init and should be called on application exit.
type Shutdown func(ctx context.Context) error

// Init configures OpenTelemetry with OTLP/HTTP exporters for traces, metrics, and logs.
// Per Axiom docs:
// - Traces + Logs use header: X-Axiom-Dataset
// - Metrics use header: X-Axiom-Metrics-Dataset
// - Each signal type should target a separate dataset.
func Init(ctx context.Context, cfg Config) (Shutdown, error) {
	if cfg.Endpoint == "" || cfg.Token == "" {
		return func(_ context.Context) error { return nil }, nil
	}

	// Common headers (auth only)
	authHeader := map[string]string{
		"Authorization": "Bearer " + cfg.Token,
	}

	// Traces + Logs headers
	dataHeaders := map[string]string{
		"Authorization":   "Bearer " + cfg.Token,
		"X-Axiom-Dataset": cfg.Dataset,
	}

	// Metrics headers (different header name per Axiom docs)
	metricsHeaders := map[string]string{
		"Authorization":           "Bearer " + cfg.Token,
		"X-Axiom-Metrics-Dataset": cfg.MetricsDataset,
	}
	_ = authHeader // not used directly

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

	// --- Traces ---
	traceExp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.Endpoint),
		otlptracehttp.WithHeaders(dataHeaders),
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

	// --- Metrics ---
	metricExp, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(cfg.Endpoint),
		otlpmetrichttp.WithHeaders(metricsHeaders),
	)
	if err != nil {
		return nil, fmt.Errorf("create metric exporter: %w", err)
	}

	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(30*time.Second))),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	// --- Logs ---
	logExp, err := otlploghttp.New(ctx,
		otlploghttp.WithEndpoint(cfg.Endpoint),
		otlploghttp.WithHeaders(dataHeaders),
	)
	if err != nil {
		return nil, fmt.Errorf("create log exporter: %w", err)
	}

	lp := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExp, log.WithExportInterval(5*time.Second))),
		log.WithResource(res),
	)
	otellog.SetLoggerProvider(lp)

	// --- Propagator ---
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func(ctx context.Context) error {
		var errs []error
		if err := tp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := mp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if err := lp.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		if len(errs) > 0 {
			return fmt.Errorf("otel shutdown errors: %v", errs)
		}
		return nil
	}

	return shutdown, nil
}
