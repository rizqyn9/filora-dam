package lib

import "github.com/gofiber/fiber/v3"

// Envelope is the standard API response wrapper.
//
//	success: { "success": true,  "data": {...} }
//	error:   { "success": false, "error": { "code": "...", "message": "..." } }
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

// ErrorBody is the error payload of an unsuccessful response.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 success response.
func OK(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(Envelope{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(Envelope{Success: true, Data: data})
}

// Fail writes an error response with the given status, code, and message.
func Fail(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	})
}
