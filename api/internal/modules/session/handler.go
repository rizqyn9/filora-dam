package session

import (
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("/cli/sessions", authMW)
	g.Post("/", h.issue)
	g.Get("/", h.list)
	g.Post("/revoke-all", h.revokeAll)
	g.Delete("/:id", h.revoke)
}

func (h *Handler) issue(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	var in IssueInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}

	ip := optional(c.IP())
	ua := optional(c.Get("User-Agent"))

	res, err := h.svc.Issue(c.Context(), p.UserID, in, ip, ua)
	if err != nil {
		return err
	}
	return lib.Created(c, res)
}

func (h *Handler) list(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	sessions, err := h.svc.List(c.Context(), p.UserID)
	if err != nil {
		return err
	}
	return lib.OK(c, sessions)
}

func (h *Handler) revoke(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return lib.ErrBadRequest("invalid session id")
	}
	if err := h.svc.Revoke(c.Context(), id, p.UserID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"revoked": true})
}

func (h *Handler) revokeAll(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	if err := h.svc.RevokeAll(c.Context(), p.UserID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"revoked": true})
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
