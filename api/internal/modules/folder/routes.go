package folder

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	folders := router.Group("/folders")
	folders.Get("/", h.ListFolders)
	folders.Post("/", h.CreateFolder)
	folders.Get("/:id", h.GetFolder)
	folders.Patch("/:id/rename", h.RenameFolder)
	folders.Patch("/:id/move", h.MoveFolder)
	folders.Delete("/:id", h.DeleteFolder)
	folders.Post("/:id/restore", h.RestoreFolder)
	folders.Get("/:id/breadcrumbs", h.GetBreadcrumbs)
}
