package account

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1/account")

	api.Get("/:id", h.GetUser)
	api.Get("/:id/quota", h.GetQuota)
}

func (h *Handler) GetUser(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "User ID is required")
	}

	user, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return lib.NotFound(c, "User not found")
		}
		return lib.InternalError(c, "Failed to get user")
	}

	return lib.Success(c, user)
}

func (h *Handler) GetQuota(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "User ID is required")
	}

	quota, err := h.service.GetQuota(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return lib.NotFound(c, "User not found")
		}
		return lib.InternalError(c, "Failed to get quota")
	}

	return lib.Success(c, quota)
}
