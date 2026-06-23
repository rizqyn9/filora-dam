package storage

import (
	"errors"

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
	storage := app.Group("/api/v1/storage")

	// Protected routes
	storage.Use(authMiddleware)

	// Provider management
	storage.Get("/providers", h.ListProviders)
	storage.Post("/providers", h.CreateProvider)
	storage.Get("/providers/:id", h.GetProvider)
	storage.Delete("/providers/:id", h.DeactivateProvider)

	// Upload
	storage.Post("/upload", h.Upload)

	// Download
	storage.Get("/download/:id", h.Download)
}

func (h *Handler) ListProviders(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	activeOnly := c.Query("active") == "true"

	providers, err := h.service.ListProviders(c.Context(), userID, activeOnly)
	if err != nil {
		return lib.InternalError(c, "Failed to list providers")
	}

	// Remove credentials from response
	for _, p := range providers {
		p.Credentials = nil
	}

	return lib.Success(c, providers)
}

func (h *Handler) GetProvider(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "Provider ID is required")
	}

	provider, err := h.service.GetProviderByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return lib.NotFound(c, "Provider not found")
		}
		return lib.InternalError(c, "Failed to get provider")
	}

	// Verify provider belongs to user
	if provider.UserID != userID {
		return lib.Forbidden(c, "Access denied")
	}

	// Remove credentials from response
	provider.Credentials = nil

	return lib.Success(c, provider)
}

func (h *Handler) CreateProvider(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	var req CreateProviderRequest
	if err := c.Bind().Body(&req); err != nil {
		return lib.BadRequest(c, "Invalid request body")
	}

	// Basic validation
	if req.Name == "" || req.Type == "" || req.Credentials == nil {
		return lib.BadRequest(c, "Name, type, and credentials are required")
	}

	provider, err := h.service.CreateProvider(c.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrInvalidProviderType) {
			return lib.BadRequest(c, "Invalid provider type. Must be: cloudinary, imagekit, or r2")
		}
		return lib.InternalError(c, "Failed to create provider")
	}

	// Remove credentials from response
	provider.Credentials = nil

	return c.Status(fiber.StatusCreated).JSON(lib.SuccessResponse{
		Success: true,
		Data:    provider,
	})
}

func (h *Handler) DeactivateProvider(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	id := c.Params("id")
	if id == "" {
		return lib.BadRequest(c, "Provider ID is required")
	}

	// Verify provider belongs to user
	provider, err := h.service.GetProviderByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, ErrProviderNotFound) {
			return lib.NotFound(c, "Provider not found")
		}
		return lib.InternalError(c, "Failed to get provider")
	}

	if provider.UserID != userID {
		return lib.Forbidden(c, "Access denied")
	}

	if err := h.service.DeactivateProvider(c.Context(), id); err != nil {
		return lib.InternalError(c, "Failed to deactivate provider")
	}

	return lib.Success(c, fiber.Map{
		"message": "Provider deactivated successfully",
	})
}

func (h *Handler) Upload(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	// Parse multipart form
	file, err := c.FormFile("file")
	if err != nil {
		return lib.BadRequest(c, "File is required")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return lib.InternalError(c, "Failed to open file")
	}
	defer src.Close()

	// Upload
	asset, err := h.service.UploadFile(c.Context(), userID, src, file.Filename, file.Size)
	if err != nil {
		return lib.InternalError(c, "Failed to upload file")
	}

	return c.Status(fiber.StatusCreated).JSON(lib.SuccessResponse{
		Success: true,
		Data:    asset,
	})
}

func (h *Handler) Download(c fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return lib.Unauthorized(c, "User not authenticated")
	}

	assetID := c.Params("id")
	if assetID == "" {
		return lib.BadRequest(c, "Asset ID is required")
	}

	// Download file
	rc, err := h.service.DownloadFile(c.Context(), assetID, userID)
	if err != nil {
		if errors.Is(err, ErrNoActiveProvider) {
			return lib.BadRequest(c, "No storage provider available")
		}
		return lib.InternalError(c, "Failed to download file")
	}
	defer rc.Close()

	// Stream file
	return c.SendStream(rc)
}
