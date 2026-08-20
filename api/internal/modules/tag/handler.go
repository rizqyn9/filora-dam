package tag

import (
	"strconv"

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

func (h *Handler) ListTags(c fiber.Ctx) error {
	spaceID, err := uuid.Parse(c.Query("space_id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "space_id is required")
	}

	tags, err := h.service.ListTags(c.Context(), spaceID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list tags")
	}

	return lib.JSON(c, fiber.StatusOK, tags)
}

func (h *Handler) CreateTag(c fiber.Ctx) error {
	var req CreateTagRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	t, err := h.service.CreateTag(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create tag")
	}

	return lib.JSON(c, fiber.StatusCreated, t)
}

func (h *Handler) DeleteTag(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid tag ID")
	}

	if err := h.service.DeleteTag(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete tag")
	}

	return lib.JSON(c, fiber.StatusNoContent, nil)
}

func (h *Handler) TagAsset(c fiber.Ctx) error {
	var req TagAssetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.TagAsset(c.Context(), req.AssetID, req.TagID); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to tag asset")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"tagged": true})
}

func (h *Handler) UntagAsset(c fiber.Ctx) error {
	var req TagAssetRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := h.service.UntagAsset(c.Context(), req.AssetID, req.TagID); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to untag asset")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"untagged": true})
}

func (h *Handler) ListAssetTags(c fiber.Ctx) error {
	assetID, err := uuid.Parse(c.Params("asset_id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid asset ID")
	}

	tags, err := h.service.ListAssetTags(c.Context(), assetID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list asset tags")
	}

	return lib.JSON(c, fiber.StatusOK, tags)
}
