package tag

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
	byGallery := router.Group("/galleries/:galleryId/tags", authMW)
	byGallery.Post("/", h.create)
	byGallery.Get("/", h.list)

	t := router.Group("/tags", authMW)
	t.Patch("/:id", h.update)
	t.Delete("/:id", h.delete)
	t.Post("/:id/assets", h.attach)
	t.Delete("/:id/assets/:assetId", h.detach)
}

func (h *Handler) create(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	galleryID, err := paramInt64(c, "galleryId")
	if err != nil {
		return err
	}
	var in CreateTagInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	t, err := h.svc.Create(c.Context(), p.UserID, galleryID, in)
	if err != nil {
		return err
	}
	return lib.Created(c, t)
}

func (h *Handler) list(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	galleryID, err := paramInt64(c, "galleryId")
	if err != nil {
		return err
	}
	tags, err := h.svc.ListByGallery(c.Context(), p.UserID, galleryID)
	if err != nil {
		return err
	}
	return lib.OK(c, tags)
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
	var in UpdateTagInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	t, err := h.svc.Update(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, t)
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

func (h *Handler) attach(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in AttachInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	if err := h.svc.Attach(c.Context(), p.UserID, id, in); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"attached": true})
}

func (h *Handler) detach(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
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
	if err := h.svc.Detach(c.Context(), p.UserID, id, assetID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"detached": true})
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
