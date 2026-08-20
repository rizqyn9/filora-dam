package session

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	sessions := router.Group("/sessions")
	sessions.Post("/", h.CreateSession)
	sessions.Get("/", h.ListSessions)
	sessions.Delete("/:id", h.RevokeSession)
}
