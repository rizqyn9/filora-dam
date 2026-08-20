package asset

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	assets := router.Group("/assets")
	assets.Get("/", h.ListAssets)
	assets.Post("/upload", h.Upload)
	assets.Get("/:id", h.GetAsset)
	assets.Patch("/:id/rename", h.RenameAsset)
	assets.Post("/references", h.CreateReference)
	assets.Delete("/references/:ref_id", h.DeleteReference)
}
