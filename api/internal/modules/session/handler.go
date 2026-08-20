package session

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateSession(c fiber.Ctx) error {
	user := auth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req CreateSessionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	raw, sess, err := h.service.CreateSession(c.Context(), user.UserID, req.Label)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create session")
	}

	return lib.JSON(c, fiber.StatusCreated, SessionResponse{
		ID:    sess.ID,
		Label: sess.Label,
		Token: raw, // only time the raw token is returned
	})
}

func (h *Handler) ListSessions(c fiber.Ctx) error {
	user := auth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	sessions, err := h.service.ListSessions(c.Context(), user.UserID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list sessions")
	}

	return lib.JSON(c, fiber.StatusOK, sessions)
}

func (h *Handler) RevokeSession(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid session ID")
	}

	if err := h.service.RevokeSession(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to revoke session")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"revoked": true})
}
