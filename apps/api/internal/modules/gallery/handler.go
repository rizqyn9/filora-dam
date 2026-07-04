package gallery

import (
	"strconv"

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
	g := router.Group("/galleries", authMW)
	g.Post("/", h.create)
	g.Get("/", h.list)
	g.Get("/:id", h.get)
	g.Patch("/:id", h.update)
	g.Delete("/:id", h.delete)

	g.Get("/:id/members", h.listMembers)
	g.Patch("/:id/members/:userId", h.updateMemberRole)
	g.Delete("/:id/members/:userId", h.removeMember)

	g.Post("/:id/invitations", h.invite)
	g.Get("/:id/invitations", h.listInvitations)
	g.Delete("/:id/invitations/:invId", h.revokeInvitation)

	// Accepting an invitation is done by the invited (authenticated) user.
	router.Post("/invitations/accept", authMW, h.acceptInvitation)
}

func (h *Handler) create(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	var in CreateGalleryInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	g, err := h.svc.Create(c.Context(), p.UserID, in)
	if err != nil {
		return err
	}
	return lib.Created(c, g)
}

func (h *Handler) list(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	gs, err := h.svc.List(c.Context(), p.UserID)
	if err != nil {
		return err
	}
	return lib.OK(c, gs)
}

func (h *Handler) get(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	g, err := h.svc.Get(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, g)
}

func (h *Handler) update(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in UpdateGalleryInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	g, err := h.svc.Update(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, g)
}

func (h *Handler) delete(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Context(), p.UserID, id); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"deleted": true})
}

func (h *Handler) listMembers(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	members, err := h.svc.ListMembers(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, members)
}

func (h *Handler) updateMemberRole(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	userID, err := paramInt64(c, "userId")
	if err != nil {
		return err
	}
	var in UpdateMemberRoleInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	if err := h.svc.UpdateMemberRole(c.Context(), p.UserID, id, userID, in); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"updated": true})
}

func (h *Handler) removeMember(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	userID, err := paramInt64(c, "userId")
	if err != nil {
		return err
	}
	if err := h.svc.RemoveMember(c.Context(), p.UserID, id, userID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"removed": true})
}

func (h *Handler) invite(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in InviteInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	inv, err := h.svc.Invite(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.Created(c, inv)
}

func (h *Handler) listInvitations(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	invs, err := h.svc.ListInvitations(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, invs)
}

func (h *Handler) revokeInvitation(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	invID, err := paramInt64(c, "invId")
	if err != nil {
		return err
	}
	if err := h.svc.RevokeInvitation(c.Context(), p.UserID, id, invID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"revoked": true})
}

func (h *Handler) acceptInvitation(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	var in AcceptInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	g, err := h.svc.Accept(c.Context(), p.UserID, p.Email, in)
	if err != nil {
		return err
	}
	return lib.OK(c, g)
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
