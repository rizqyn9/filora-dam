package telemetry

import (
	"fmt"

	"github.com/gofiber/fiber/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "filora-api"

// FiberMiddleware creates a span for each HTTP request.
func FiberMiddleware() fiber.Handler {
	tracer := otel.Tracer(tracerName)

	return func(c fiber.Ctx) error {
		// Extract propagation context from incoming headers
		ctx := otel.GetTextMapPropagator().Extract(c.Context(), requestHeaders(c))

		spanName := fmt.Sprintf("%s %s", c.Method(), c.Route().Path)
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Method()),
				attribute.String("http.route", c.Route().Path),
				attribute.String("http.url", c.OriginalURL()),
				attribute.String("http.user_agent", c.Get("User-Agent")),
			),
		)
		defer span.End()

		c.SetContext(ctx)

		err := c.Next()

		status := c.Response().StatusCode()
		span.SetAttributes(attribute.Int("http.status_code", status))

		if status >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", status))
		}
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		return err
	}
}

// requestHeaders adapts Fiber headers to propagation.TextMapCarrier.
type headerCarrier struct {
	c fiber.Ctx
}

func requestHeaders(c fiber.Ctx) propagation.TextMapCarrier {
	return &headerCarrier{c: c}
}

func (h *headerCarrier) Get(key string) string {
	return h.c.Get(key)
}

func (h *headerCarrier) Set(key, value string) {
	// Not used for extraction
}

func (h *headerCarrier) Keys() []string {
	var keys []string
	for k := range h.c.Request().Header.All() {
		keys = append(keys, string(k))
	}
	return keys
}
