package lib

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v3"
)

// AppError is a domain error carrying an HTTP status, a stable machine code, and
// a human-readable message. Services return these; the central error handler
// converts them into the standard error envelope.
type AppError struct {
	Status  int
	Code    string
	Message string
	wrapped error
}

func (e *AppError) Error() string {
	if e.wrapped != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.wrapped)
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.wrapped }

// Wrap attaches an underlying error for logging without exposing it to clients.
func (e *AppError) Wrap(err error) *AppError {
	return &AppError{Status: e.Status, Code: e.Code, Message: e.Message, wrapped: err}
}

// NewAppError builds an AppError.
func NewAppError(status int, code, message string) *AppError {
	return &AppError{Status: status, Code: code, Message: message}
}

// Common error constructors. Use .Wrap(err) to attach a cause.
func ErrBadRequest(msg string) *AppError {
	return NewAppError(fiber.StatusBadRequest, "BAD_REQUEST", msg)
}
func ErrValidation(msg string) *AppError {
	return NewAppError(fiber.StatusBadRequest, "VALIDATION_ERROR", msg)
}
func ErrUnauthorized(msg string) *AppError {
	return NewAppError(fiber.StatusUnauthorized, "UNAUTHORIZED", msg)
}
func ErrForbidden(msg string) *AppError { return NewAppError(fiber.StatusForbidden, "FORBIDDEN", msg) }
func ErrNotFound(msg string) *AppError  { return NewAppError(fiber.StatusNotFound, "NOT_FOUND", msg) }
func ErrConflict(msg string) *AppError  { return NewAppError(fiber.StatusConflict, "CONFLICT", msg) }
func ErrInternal(msg string) *AppError {
	return NewAppError(fiber.StatusInternalServerError, "INTERNAL", msg)
}

// HTTPError maps any error to (status, code, clientMessage). Unknown errors are
// reported as a generic 500 so internals are never leaked to clients.
func HTTPError(err error) (int, string, string) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Status, appErr.Code, appErr.Message
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code, "ERROR", fiberErr.Message
	}

	return fiber.StatusInternalServerError, "INTERNAL", "Internal Server Error"
}
