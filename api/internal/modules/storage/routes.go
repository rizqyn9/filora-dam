package storage

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(router fiber.Router, h *Handler) {
	accounts := router.Group("/storage/accounts")
	accounts.Get("/", h.ListAccounts)
	accounts.Post("/", h.CreateAccount)
	accounts.Get("/:id", h.GetAccount)
	accounts.Put("/:id", h.UpdateAccount)
	accounts.Post("/:id/deactivate", h.DeactivateAccount)
}
