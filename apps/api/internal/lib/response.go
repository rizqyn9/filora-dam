package lib

import "github.com/gofiber/fiber/v3"

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool       `json:"success"`
	Error   ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success sends a success response
func Success(c fiber.Ctx, data interface{}) error {
	return c.JSON(SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// Error sends an error response
func Error(c fiber.Ctx, statusCode int, code, message string) error {
	return c.Status(statusCode).JSON(ErrorResponse{
		Success: false,
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// BadRequest sends a 400 error
func BadRequest(c fiber.Ctx, message string) error {
	return Error(c, fiber.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized sends a 401 error
func Unauthorized(c fiber.Ctx, message string) error {
	return Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden sends a 403 error
func Forbidden(c fiber.Ctx, message string) error {
	return Error(c, fiber.StatusForbidden, "FORBIDDEN", message)
}

// NotFound sends a 404 error
func NotFound(c fiber.Ctx, message string) error {
	return Error(c, fiber.StatusNotFound, "NOT_FOUND", message)
}

// InternalError sends a 500 error
func InternalError(c fiber.Ctx, message string) error {
	return Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", message)
}
