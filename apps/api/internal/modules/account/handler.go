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
	// Account routes
	account := app.Group("/api/v1/account")
	account.Get("/:id", h.GetUser)
	account.Get("/:id/quota", h.GetQuota)

	// Auth routes
	auth := app.Group("/api/v1/auth")
	auth.Post("/register", h.Register)
	auth.Post("/login", h.Login)
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

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return lib.BadRequest(c, "Invalid request body")
	}

	// Validate request
	// Note: Fiber v3 validation will be added later
	if req.Email == "" || req.Name == "" || req.Password == "" {
		return lib.BadRequest(c, "Email, name, and password are required")
	}

	if len(req.Password) < 8 {
		return lib.BadRequest(c, "Password must be at least 8 characters")
	}

	response, err := h.service.Register(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyTaken) {
			return lib.Error(c, fiber.StatusConflict, "EMAIL_TAKEN", "Email already taken")
		}
		return lib.InternalError(c, "Failed to register user")
	}

	return c.Status(fiber.StatusCreated).JSON(lib.SuccessResponse{
		Success: true,
		Data:    response,
	})
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return lib.BadRequest(c, "Invalid request body")
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		return lib.BadRequest(c, "Email and password are required")
	}

	response, err := h.service.Login(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return lib.Unauthorized(c, "Invalid email or password")
		}
		return lib.InternalError(c, "Failed to login")
	}

	return lib.Success(c, response)
}
