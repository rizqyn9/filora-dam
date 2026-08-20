package auth

import (
	"github.com/gofiber/fiber/v3"

	iauth "github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	resp, err := h.service.Login(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "AUTH_FAILED", err.Error())
	}

	return lib.JSON(c, fiber.StatusOK, resp)
}

func (h *Handler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	resp, err := h.service.Register(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "REGISTER_FAILED", err.Error())
	}

	return lib.JSON(c, fiber.StatusCreated, resp)
}

func (h *Handler) Logout(c fiber.Ctx) error {
	user := iauth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	if err := h.service.Logout(c.Context(), user.SessionID); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to logout")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"logged_out": true})
}

func (h *Handler) Me(c fiber.Ctx) error {
	user := iauth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	me, err := h.service.GetMe(c.Context(), user.UserID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to get user")
	}

	return lib.JSON(c, fiber.StatusOK, me)
}

func (h *Handler) ChangePassword(c fiber.Ctx) error {
	user := iauth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req ChangePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.ChangePassword(c.Context(), user.UserID, req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "PASSWORD_CHANGE_FAILED", err.Error())
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"changed": true})
}
