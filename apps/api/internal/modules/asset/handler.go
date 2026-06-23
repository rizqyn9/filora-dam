package asset

import (
	"errors"
	"strconv"

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
	assets := app.Group("/api/v1/assets")

	// Protected routes
	assets.Use(authMiddleware)

	assets.Get("/", h.ListAssets)
	assets.Get("/search", h.SearchAssets)
	assets.Get("/filter/:type", h.FilterAssets)
	assets.Get("/:id", h.GetAsset)
	assets.Delete("/:id", h.DeleteAsset)
	assets.Put("/:id/tags", h.UpdateTags)
}

func (h *Handler) ListAssets(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	// Get pagination params
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	response, err := h.service.ListAssets(c.Context(), userID, limit, offset)
	if err != nil {
		return lib.InternalError(c, "Failed to list assets")
	}

	return lib.Success(c, response)
}

func (h *Handler) GetAsset(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "Asset ID is required")
	}

	asset, err := h.service.GetByID(c.Context(), id, userID)
	if err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return lib.NotFound(c, "Asset not found")
		}
		if errors.Is(err, ErrAccessDenied) {
			return lib.Forbidden(c, "Access denied")
		}
		return lib.InternalError(c, "Failed to get asset")
	}

	return lib.Success(c, asset)
}

func (h *Handler) DeleteAsset(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "Asset ID is required")
	}

	if err := h.service.DeleteAsset(c.Context(), id, userID); err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return lib.NotFound(c, "Asset not found")
		}
		if errors.Is(err, ErrAccessDenied) {
			return lib.Forbidden(c, "Access denied")
		}
		return lib.InternalError(c, "Failed to delete asset")
	}

	return lib.Success(c, fiber.Map{
		"message": "Asset deleted successfully",
	})
}

func (h *Handler) UpdateTags(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "Asset ID is required")
	}

	var req UpdateTagsRequest
	if err := c.Bind().Body(&req); err != nil {
		return lib.BadRequest(c, "Invalid request body")
	}

	if req.Tags == nil {
		return lib.BadRequest(c, "Tags are required")
	}

	if err := h.service.UpdateTags(c.Context(), id, userID, req.Tags); err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return lib.NotFound(c, "Asset not found")
		}
		if errors.Is(err, ErrAccessDenied) {
			return lib.Forbidden(c, "Access denied")
		}
		return lib.InternalError(c, "Failed to update tags")
	}

	return lib.Success(c, fiber.Map{
		"message": "Tags updated successfully",
	})
}

func (h *Handler) SearchAssets(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	query := c.Query("q")
	if query == "" {
		return lib.BadRequest(c, "Search query is required")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	response, err := h.service.SearchAssets(c.Context(), userID, query, limit, offset)
	if err != nil {
		return lib.InternalError(c, "Failed to search assets")
	}

	return lib.Success(c, response)
}

func (h *Handler) FilterAssets(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	assetType := c.Params("type")
	if assetType == "" {
		return lib.BadRequest(c, "Asset type is required")
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	response, err := h.service.FilterAssets(c.Context(), userID, assetType, limit, offset)
	if err != nil {
		return lib.InternalError(c, "Failed to filter assets")
	}

	return lib.Success(c, response)
}
