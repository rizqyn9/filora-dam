package account

import (
	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/clerk"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Handler exposes account HTTP endpoints.
type Handler struct {
	svc           *Service
	webhookSecret string
}

func NewHandler(svc *Service, webhookSecret string) *Handler {
	return &Handler{svc: svc, webhookSecret: webhookSecret}
}

// RegisterRoutes mounts protected /me routes (behind authMW) and the public,
// signature-verified Clerk webhook.
func (h *Handler) RegisterRoutes(router fiber.Router, authMW fiber.Handler) {
	me := router.Group("/me", authMW)
	me.Get("/", h.getMe)
	me.Patch("/", h.updateMe)

	router.Post("/webhooks/clerk", h.clerkWebhook)
}

func (h *Handler) getMe(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	u, err := h.svc.GetByID(c.Context(), p.UserID)
	if err != nil {
		return err
	}
	return lib.OK(c, u)
}

func (h *Handler) updateMe(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	var in UpdateProfileInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	u, err := h.svc.UpdateProfile(c.Context(), p.UserID, in)
	if err != nil {
		return err
	}
	return lib.OK(c, u)
}

func (h *Handler) clerkWebhook(c fiber.Ctx) error {
	if h.webhookSecret == "" {
		return lib.ErrInternal("clerk webhook is not configured")
	}

	body := c.Body()
	headers := clerk.SvixHeaders{
		ID:        c.Get("svix-id"),
		Timestamp: c.Get("svix-timestamp"),
		Signature: c.Get("svix-signature"),
	}
	if err := clerk.VerifySvix(h.webhookSecret, headers, body); err != nil {
		return lib.ErrUnauthorized("invalid webhook signature").Wrap(err)
	}

	eventType, ident, err := clerk.ParseWebhook(body)
	if err != nil {
		return lib.ErrBadRequest("invalid webhook payload").Wrap(err)
	}

	if _, err := h.svc.ProcessWebhook(c.Context(), headers.ID, eventType, body, ident); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"received": true})
}
