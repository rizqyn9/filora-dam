package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "filora-api"

// Metrics holds all application-level OTel metrics.
type Metrics struct {
	UploadsTotal    metric.Int64Counter
	UploadsBytes    metric.Int64Counter
	UploadsDuration metric.Float64Histogram
	ArchiveJobs     metric.Int64Counter
	AuthLogins      metric.Int64Counter
}

// NewMetrics initializes all application metrics.
func NewMetrics() *Metrics {
	meter := otel.Meter(meterName)

	uploadsTotal, _ := meter.Int64Counter("filora.uploads.total",
		metric.WithDescription("Total upload operations"),
	)
	uploadsBytes, _ := meter.Int64Counter("filora.uploads.bytes",
		metric.WithDescription("Total bytes uploaded"),
	)
	uploadsDuration, _ := meter.Float64Histogram("filora.uploads.duration_ms",
		metric.WithDescription("Upload duration in milliseconds"),
	)
	archiveJobs, _ := meter.Int64Counter("filora.archive.jobs.total",
		metric.WithDescription("Archive sync jobs by status"),
	)
	authLogins, _ := meter.Int64Counter("filora.auth.logins.total",
		metric.WithDescription("Login attempts by client and result"),
	)

	return &Metrics{
		UploadsTotal:    uploadsTotal,
		UploadsBytes:    uploadsBytes,
		UploadsDuration: uploadsDuration,
		ArchiveJobs:     archiveJobs,
		AuthLogins:      authLogins,
	}
}
