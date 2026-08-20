package asset

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/rizqyn9/filora-dam/api/internal/auth"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
	authz   auth.SpaceMemberChecker
	logger  zerolog.Logger
}

func NewHandler(service *Service, authz auth.SpaceMemberChecker, logger zerolog.Logger) *Handler {
	return &Handler{service: service, authz: authz, logger: logger}
}

func (h *Handler) GetAsset(c fiber.Ctx) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
	}

	a, err := h.service.GetAsset(c.Context(), id)
	if err != nil {
		h.logger.Error().Err(err).Str("asset_id", id.String()).Msg("get asset failed")
		return lib.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "asset not found")
	}

	return lib.JSON(c, fiber.StatusOK, a)
}

func (h *Handler) ListAssets(c fiber.Ctx) error {
	user := auth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	spaceID, err := parseQueryUUID(c, "space_id")
	if err != nil {
		return err
	}

	// Authz: user must be member of space
	if err := auth.RequireSpaceAccess(c.Context(), h.authz, spaceID, user.UserID); err != nil {
		return lib.JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "no access to this space")
	}

	var folderID *uuid.UUID
	if f := c.Query("folder_id"); f != "" {
		parsed, perr := uuid.Parse(f)
		if perr != nil {
			return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid folder_id")
		}
		folderID = &parsed
	}

	limit := clampInt(c.Query("limit", "50"), 1, 200)
	offset := clampInt(c.Query("offset", "0"), 0, 1_000_000)

	assets, err := h.service.ListAssets(c.Context(), ListAssetsParams{
		SpaceID:  spaceID,
		FolderID: folderID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		h.logger.Error().Err(err).Msg("list assets failed")
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

	// Authz
	if err := auth.RequireSpaceAccess(c.Context(), h.authz, spaceID, user.UserID); err != nil {
		return lib.JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "no access to this space")
	}

	var folderID *uuid.UUID
	if f := c.FormValue("folder_id"); f != "" {
		parsed, perr := uuid.Parse(f)
		if perr != nil {
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
		h.logger.Error().Err(err).Msg("open uploaded file failed")
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to read file")
	}
	defer f.Close()

	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	a, err := h.service.Upload(c.Context(), user.UserID, UploadInput{
		SpaceID:     spaceID,
		FolderID:    folderID,
		Filename:    file.Filename,
		ContentType: contentType,
		Size:        file.Size,
		Body:        f,
	})
	if err != nil {
		// Distinguish quota errors from internal errors
		if strings.Contains(err.Error(), "quota exceeded") {
			return lib.JSONError(c, fiber.StatusUnprocessableEntity, "QUOTA_EXCEEDED", err.Error())
		}
		h.logger.Error().Err(err).Msg("upload failed")
		// If asset was created but archive job failed, still return success with warning
		if a != nil {
			return lib.JSON(c, fiber.StatusCreated, a)
		}
		return lib.JSONError(c, fiber.StatusInternalServerError, "UPLOAD_FAILED", "failed to upload file")
	}

	return lib.JSON(c, fiber.StatusCreated, a)
}

func (h *Handler) CreateReference(c fiber.Ctx) error {
	user := auth.GetUser(c)
	if user == nil {
		return lib.JSONError(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}

	var req CreateReferenceRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	// Authz
	if err := auth.RequireSpaceAccess(c.Context(), h.authz, req.SpaceID, user.UserID); err != nil {
		return lib.JSONError(c, fiber.StatusForbidden, "FORBIDDEN", "no access to this space")
	}

	ref, err := h.service.CreateReference(c.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("create reference failed")
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create reference")
	}

	return lib.JSON(c, fiber.StatusCreated, ref)
}

func (h *Handler) DeleteReference(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("ref_id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid reference ID")
	}

	// TODO: look up reference to get spaceID + sizeBytes for proper authz + quota decrement
	if err := h.service.DeleteReference(c.Context(), id, uuid.Nil, 0); err != nil {
		h.logger.Error().Err(err).Int64("ref_id", id).Msg("delete reference failed")
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to delete reference")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) RenameAsset(c fiber.Ctx) error {
	id, err := parseUUID(c, "id")
	if err != nil {
		return err
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
		h.logger.Error().Err(err).Str("asset_id", id.String()).Msg("rename asset failed")
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to rename asset")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"renamed": true})
}

// --- Helpers ---

func parseUUID(c fiber.Ctx, param string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(param))
	if err != nil {
		return uuid.Nil, lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid "+param)
	}
	return id, nil
}

func parseQueryUUID(c fiber.Ctx, key string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Query(key))
	if err != nil {
		return uuid.Nil, lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", key+" is required and must be a valid UUID")
	}
	return id, nil
}

func clampInt(s string, min, max int) int {
	v, err := strconv.Atoi(s)
	if err != nil || v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
