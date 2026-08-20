package asset

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAsset(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid asset ID")
	}

	a, err := h.service.GetAsset(c.Context(), id)
	if err != nil {
		return lib.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "asset not found")
	}

	return lib.JSON(c, fiber.StatusOK, a)
}

func (h *Handler) ListAssets(c fiber.Ctx) error {
	spaceID, err := uuid.Parse(c.Query("space_id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "space_id is required")
	}

	var folderID *uuid.UUID
	if f := c.Query("folder_id"); f != "" {
		parsed, err := uuid.Parse(f)
		if err != nil {
			return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder_id")
		}
		folderID = &parsed
	}

	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	assets, err := h.service.ListAssets(c.Context(), ListAssetsParams{
		SpaceID:  spaceID,
		FolderID: folderID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list assets")
	}

	return lib.JSON(c, fiber.StatusOK, assets)
}

func (h *Handler) Upload(c fiber.Ctx) error {
	user := auth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	spaceID, err := uuid.Parse(c.FormValue("space_id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "space_id is required")
	}

	var folderID *uuid.UUID
	if f := c.FormValue("folder_id"); f != "" {
		parsed, err := uuid.Parse(f)
		if err != nil {
			return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder_id")
		}
		folderID = &parsed
	}

	file, err := c.FormFile("file")
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "MISSING_FILE", "file is required")
	}

	f, err := file.Open()
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to read file")
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	a, err := h.service.Upload(c.Context(), user.UserID, spaceID, folderID, file.Filename, contentType, file.Size, f)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "UPLOAD_FAILED", err.Error())
	}

	return lib.JSON(c, fiber.StatusCreated, a)
}

func (h *Handler) CreateReference(c fiber.Ctx) error {
	var req CreateReferenceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	ref, err := h.service.CreateReference(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create reference")
	}

	return lib.JSON(c, fiber.StatusCreated, ref)
}

func (h *Handler) DeleteReference(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("ref_id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid reference ID")
	}

	if err := h.service.DeleteReference(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete reference")
	}

	return lib.JSON(c, fiber.StatusNoContent, nil)
}

func (h *Handler) RenameAsset(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid asset ID")
	}

	var req struct {
		Name string `json:"name" validate:"required,min=1,max=255"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.RenameAsset(c.Context(), id, req.Name); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to rename asset")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"renamed": true})
}
