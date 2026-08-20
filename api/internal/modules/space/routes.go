package space

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	spaces := router.Group("/spaces")
	spaces.Get("/", h.ListSpaces)
	spaces.Post("/", h.CreateSpace)
	spaces.Get("/:id", h.GetSpace)
	spaces.Put("/:id", h.UpdateSpace)
	spaces.Delete("/:id", h.DeleteSpace)
}
