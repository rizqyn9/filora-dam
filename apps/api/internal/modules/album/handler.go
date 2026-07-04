package album

import (
	"strconv"

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
	byGallery := router.Group("/galleries/:galleryId/albums", authMW)
	byGallery.Post("/", h.create)
	byGallery.Get("/", h.listByGallery)

	a := router.Group("/albums", authMW)
	a.Get("/:id", h.get)
	a.Patch("/:id", h.update)
	a.Delete("/:id", h.delete)
	a.Get("/:id/members", h.listMembers)
	a.Post("/:id/members", h.addMember)
	a.Delete("/:id/members/:userId", h.removeMember)
	a.Get("/:id/assets", h.listAssets)
	a.Post("/:id/assets", h.addAsset)
	a.Delete("/:id/assets/:assetId", h.removeAsset)
}

func (h *Handler) create(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	galleryID, err := paramInt64(c, "galleryId")
	if err != nil {
		return err
	}
	var in CreateAlbumInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	a, err := h.svc.Create(c.Context(), p.UserID, galleryID, in)
	if err != nil {
		return err
	}
	return lib.Created(c, a)
}

func (h *Handler) listByGallery(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	galleryID, err := paramInt64(c, "galleryId")
	if err != nil {
		return err
	}
	albums, err := h.svc.ListByGallery(c.Context(), p.UserID, galleryID)
	if err != nil {
		return err
	}
	return lib.OK(c, albums)
}

func (h *Handler) get(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	a, err := h.svc.Get(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, a)
}

func (h *Handler) update(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in UpdateAlbumInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	a, err := h.svc.Update(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, a)
}

func (h *Handler) delete(c fiber.Ctx) error {
	p := principal(c)
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
	p := principal(c)
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

func (h *Handler) addMember(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in AddMemberInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	if err := h.svc.AddMember(c.Context(), p.UserID, id, in); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"added": true})
}

func (h *Handler) removeMember(c fiber.Ctx) error {
	p := principal(c)
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

func (h *Handler) listAssets(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	ids, err := h.svc.ListAssetIDs(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"asset_ids": ids})
}

func (h *Handler) addAsset(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in AddAssetInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	if err := h.svc.AddAsset(c.Context(), p.UserID, id, in); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"added": true})
}

func (h *Handler) removeAsset(c fiber.Ctx) error {
	p := principal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	assetID, err := uuid.Parse(c.Params("assetId"))
	if err != nil {
		return lib.ErrBadRequest("invalid asset id")
	}
	if err := h.svc.RemoveAsset(c.Context(), p.UserID, id, assetID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"removed": true})
}

func principal(c fiber.Ctx) *auth.Principal {
	return auth.MustPrincipal(c)
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
