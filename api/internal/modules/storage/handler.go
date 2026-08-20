package storage

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListAccounts(c fiber.Ctx) error {
	accounts, err := h.service.ListAccounts(c.Context())
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to list accounts")
	}
	return lib.JSON(c, fiber.StatusOK, accounts)
}

func (h *Handler) GetAccount(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid account ID")
	}

	account, err := h.service.GetAccount(c.Context(), id)
	if err != nil {
		return lib.JSONError(c, fiber.StatusNotFound, "NOT_FOUND", "account not found")
	}

	return lib.JSON(c, fiber.StatusOK, account)
}

func (h *Handler) CreateAccount(c fiber.Ctx) error {
	var req CreateAccountRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	account, err := h.service.CreateAccount(c.Context(), req)
	if err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to create account")
	}

	return lib.JSON(c, fiber.StatusCreated, account)
}

func (h *Handler) UpdateAccount(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid account ID")
	}

	var req UpdateAccountRequest
	if err := c.Bind().JSON(&req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := lib.Validate.Struct(req); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	}

	if err := h.service.UpdateAccount(c.Context(), id, req); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to update account")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"updated": true})
}

func (h *Handler) DeactivateAccount(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_ID", "invalid account ID")
	}

	if err := h.service.DeactivateAccount(c.Context(), id); err != nil {
		return lib.JSONError(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "failed to deactivate account")
	}

	return lib.JSON(c, fiber.StatusOK, fiber.Map{"deactivated": true})
}
