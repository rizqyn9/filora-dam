package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type Server struct {
	App    *fiber.App
	Logger zerolog.Logger
}

func New(logger zerolog.Logger) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: defaultErrorHandler,
	})

	return &Server{
		App:    app,
		Logger: logger,
	}
}

func defaultErrorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"success": false,
		"error": fiber.Map{
			"code":    "INTERNAL_ERROR",
			"message": err.Error(),
		},
	})
}
