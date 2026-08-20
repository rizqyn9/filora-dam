package tag

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	tags := router.Group("/tags")
	tags.Get("/", h.ListTags)
	tags.Post("/", h.CreateTag)
	tags.Delete("/:id", h.DeleteTag)
	tags.Post("/asset", h.TagAsset)
	tags.Delete("/asset", h.UntagAsset)
	tags.Get("/asset/:asset_id", h.ListAssetTags)
}
