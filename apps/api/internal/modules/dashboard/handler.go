package dashboard

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(app *fiber.App, authMiddleware fiber.Handler) {
	dashboard := app.Group("/api/v1/dashboard")

	// Protected routes
	dashboard.Use(authMiddleware)

	dashboard.Get("/", h.GetDashboard)
}

func (h *Handler) GetDashboard(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	response, err := h.service.GetDashboard(c.Context(), userID)
	if err != nil {
		return lib.InternalError(c, "Failed to get dashboard")
	}

	return lib.Success(c, response)
}
