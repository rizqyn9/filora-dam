package folder

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

func (h *Handler) GetFolder(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	f, err := h.service.GetFolder(c.Context(), id)
	if err != nil {
		return lib.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "folder not found")
	}

	return lib.JSON(c, fiber.StatusOK, f)
}

func (h *Handler) ListFolders(c fiber.Ctx) error {
	spaceID, err := uuid.Parse(c.Query("space_id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "space_id is required")
	}

	var parentID *uuid.UUID
	if p := c.Query("parent_id"); p != "" {
		parsed, err := uuid.Parse(p)
		if err != nil {
			return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid parent_id")
		}
		parentID = &parsed
	}

	folders, err := h.service.ListFolders(c.Context(), spaceID, parentID)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list folders")
	}

	return lib.JSON(c, fiber.StatusOK, folders)
}

func (h *Handler) CreateFolder(c fiber.Ctx) error {
	var req CreateFolderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	f, err := h.service.CreateFolder(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create folder")
	}

	return lib.JSON(c, fiber.StatusCreated, f)
}

func (h *Handler) RenameFolder(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	var req RenameFolderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.RenameFolder(c.Context(), id, req.Name); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to rename folder")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"renamed": true})
}

func (h *Handler) MoveFolder(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	var req MoveFolderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}

	if err := h.service.MoveFolder(c.Context(), id, req.ParentID); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to move folder")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"moved": true})
}

func (h *Handler) DeleteFolder(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	if err := h.service.DeleteFolder(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete folder")
	}

	return lib.JSON(c, fiber.StatusNoContent, nil)
}

func (h *Handler) RestoreFolder(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	if err := h.service.RestoreFolder(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to restore folder")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"restored": true})
}

func (h *Handler) GetBreadcrumbs(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder ID")
	}

	breadcrumbs, err := h.service.GetBreadcrumbs(c.Context(), id)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to get breadcrumbs")
	}

	return lib.JSON(c, fiber.StatusOK, breadcrumbs)
}
