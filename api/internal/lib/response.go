package lib

import "github.com/gofiber/fiber/v3"

type APIResponse struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(c fiber.Ctx, status int, data any) error {
	return c.Status(status).JSON(APIResponse{Success: true, Data: data})
}

func JSONError(c fiber.Ctx, status int, code, message string) error {
	return c.Status(status).JSON(APIResponse{
		Success: false,
		Error:   &Error{Code: code, Message: message},
	})
}
