package telemetry

import (
	"log/slog"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// NewLogger creates a zerolog.Logger for console output.
// Call SetupSlog after OTel init to also route logs to Axiom via OTel.
func NewLogger() zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}).
		With().Timestamp().Logger()
}

// SetupSlog configures Go's standard slog to send logs through OTel.
// All slog.Info/Warn/Error calls are exported as OTel log records to Axiom.
// Use slog for application-level logging that should appear in Axiom.
// Use zerolog for local console output only.
func SetupSlog() {
	handler := otelslog.NewHandler("filora-api")
	slog.SetDefault(slog.New(handler))
}
