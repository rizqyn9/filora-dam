package storage

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

type Handler struct {
	svc   *Service
	authz *auth.Authorizer
}

func NewHandler(svc *Service, authz *auth.Authorizer) *Handler {
	return &Handler{svc: svc, authz: authz}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("/storage", authMW)
	g.Get("/providers", h.list)
	g.Post("/providers", h.create)
	g.Get("/providers/:id", h.get)
	g.Patch("/providers/:id", h.update)
	g.Delete("/providers/:id", h.deactivate)
	g.Get("/usage", h.usage)
}

func (h *Handler) require(c fiber.Ctx, action string) (*auth.Principal, error) {
	p := auth.MustPrincipal(c)
	if p == nil {
		return nil, lib.ErrUnauthorized("not authenticated")
	}
	if err := h.authz.Require(c.Context(), p.UserID, "storage", action); err != nil {
		return nil, err
	}
	return p, nil
}

func (h *Handler) list(c fiber.Ctx) error {
	if _, err := h.require(c, "read"); err != nil {
		return err
	}
	providers, err := h.svc.List(c.Context())
	if err != nil {
		return err
	}
	return lib.OK(c, providers)
}

func (h *Handler) create(c fiber.Ctx) error {
	p, err := h.require(c, "create")
	if err != nil {
		return err
	}
	var in CreateProviderInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	created := p.UserID
	provider, err := h.svc.Create(c.Context(), in, &created)
	if err != nil {
		return err
	}
	return lib.Created(c, provider)
}

func (h *Handler) get(c fiber.Ctx) error {
	if _, err := h.require(c, "read"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	provider, err := h.svc.Get(c.Context(), id)
	if err != nil {
		return err
	}
	return lib.OK(c, provider)
}

func (h *Handler) update(c fiber.Ctx) error {
	if _, err := h.require(c, "update"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in UpdateProviderInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	provider, err := h.svc.Update(c.Context(), id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, provider)
}

func (h *Handler) deactivate(c fiber.Ctx) error {
	if _, err := h.require(c, "delete"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Deactivate(c.Context(), id); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"deactivated": true})
}

func (h *Handler) usage(c fiber.Ctx) error {
	if _, err := h.require(c, "read"); err != nil {
		return err
	}
	usage, err := h.svc.Usage(c.Context())
	if err != nil {
		return err
	}
	return lib.OK(c, usage)
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
