package account

import (
	"github.com/gofiber/fiber/v3"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

// WebhookHandler handles Clerk webhook events for user sync.
type WebhookHandler struct {
	service *Service
}

func NewWebhookHandler(service *Service) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// HandleWebhook processes Clerk webhook events.
// In production, verify the webhook signature with the Clerk SDK.
func (h *WebhookHandler) HandleWebhook(c fiber.Ctx) error {
	// TODO: verify Clerk webhook signature (svix)

	var payload struct {
		Type string `json:"type"`
		Data struct {
			ID            string  `json:"id"`
			EmailAddress  string  `json:"email_address,omitempty"`
			FirstName     string  `json:"first_name,omitempty"`
			LastName      string  `json:"last_name,omitempty"`
			ImageURL      *string `json:"image_url,omitempty"`
			// Clerk sends email_addresses as array; simplified here
			EmailAddresses []struct {
				EmailAddress string `json:"email_address"`
			} `json:"email_addresses,omitempty"`
		} `json:"data"`
	}

	if err := c.Bind().JSON(&payload); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid webhook payload")
	}

	email := payload.Data.EmailAddress
	if email == "" && len(payload.Data.EmailAddresses) > 0 {
		email = payload.Data.EmailAddresses[0].EmailAddress
	}

	name := payload.Data.FirstName
	if payload.Data.LastName != "" {
		name += " " + payload.Data.LastName
	}

	switch payload.Type {
	case "user.created", "user.updated":
		_, err := h.service.SyncUser(c.Context(), payload.Data.ID, email, name, payload.Data.ImageURL)
		if err != nil {
			return lib.JSONError(c, fiber.StatusInternalServerError, "SYNC_FAILED", "failed to sync user")
		}
	case "user.deleted":
		if err := h.service.DeleteUser(c.Context(), payload.Data.ID); err != nil {
			return lib.JSONError(c, fiber.StatusInternalServerError, "DELETE_FAILED", "failed to delete user")
		}
	}

	return c.SendStatus(fiber.StatusOK)
}

func RegisterRoutes(router fiber.Router, wh *WebhookHandler) {
	router.Post("/webhooks/clerk", wh.HandleWebhook)
}
