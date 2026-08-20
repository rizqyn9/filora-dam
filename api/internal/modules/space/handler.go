package space

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSpace(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid space ID")
	}

	space, err := h.service.GetSpace(c.Context(), id)
	if err != nil {
		return lib.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "space not found")
	}

	return lib.JSON(c, fiber.StatusOK, space)
}

func (h *Handler) ListSpaces(c fiber.Ctx) error {
	// TODO: extract user ID from auth context
	userID := int64(1) // placeholder

	spaces, err := h.service.ListSpaces(c.Context(), userID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list spaces")
	}

	return lib.JSON(c, fiber.StatusOK, spaces)
}

func (h *Handler) CreateSpace(c fiber.Ctx) error {
	var req CreateSpaceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// TODO: extract user ID from auth context
	userID := int64(1) // placeholder

	space, err := h.service.CreateSpace(c.Context(), userID, req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create space")
	}

	return lib.JSON(c, fiber.StatusCreated, space)
}

func (h *Handler) UpdateSpace(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid space ID")
	}

	var req UpdateSpaceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	space, err := h.service.UpdateSpace(c.Context(), id, req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to update space")
	}

	return lib.JSON(c, fiber.StatusOK, space)
}

func (h *Handler) DeleteSpace(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid space ID")
	}

	if err := h.service.DeleteSpace(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete space")
	}

	return lib.JSON(c, fiber.StatusNoContent, nil)
}
