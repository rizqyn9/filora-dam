package account

import (
	"encoding/json"
	"net/http"

	svix "github.com/svix/svix-webhooks/go"

	"github.com/gofiber/fiber/v3"
	"github.com/rizqyn9/filora-dam/api/internal/lib"
)

// WebhookHandler handles Clerk webhook events for user sync.
type WebhookHandler struct {
	service       *Service
	webhookSecret string // Clerk webhook signing secret (whsec_...)
}

func NewWebhookHandler(service *Service, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{service: service, webhookSecret: webhookSecret}
}

// HandleWebhook processes Clerk webhook events with signature verification.
func (h *WebhookHandler) HandleWebhook(c fiber.Ctx) error {
	body := c.Body()

	// Verify webhook signature if secret is configured
	if h.webhookSecret != "" {
		wh, err := svix.NewWebhook(h.webhookSecret)
		if err != nil {
			return lib.JSONError(c, fiber.StatusInternalServerError, "WEBHOOK_ERROR", "invalid webhook config")
		}

		headers := http.Header{}
		headers.Set("svix-id", c.Get("svix-id"))
		headers.Set("svix-timestamp", c.Get("svix-timestamp"))
		headers.Set("svix-signature", c.Get("svix-signature"))

		if err := wh.Verify(body, headers); err != nil {
			return lib.JSONError(c, fiber.StatusUnauthorized, "WEBHOOK_INVALID", "invalid webhook signature")
		}
	}

	var payload struct {
		Type string `json:"type"`
		Data struct {
			ID             string  `json:"id"`
			FirstName      string  `json:"first_name"`
			LastName       string  `json:"last_name"`
			ImageURL       *string `json:"image_url"`
			EmailAddresses []struct {
				EmailAddress string `json:"email_address"`
			} `json:"email_addresses"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return lib.JSONError(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid webhook payload")
	}

	var email string
	if len(payload.Data.EmailAddresses) > 0 {
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
