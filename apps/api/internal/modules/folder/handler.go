package folder

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
	// Space-scoped folder listing (root or by parent).
	s := router.Group("/spaces/:spaceId/folders", authMW)
	s.Post("/", h.create)
	s.Get("/", h.listChildren)

	// Folder-scoped operations.
	f := router.Group("/folders", authMW)
	f.Get("/:id", h.get)
	f.Patch("/:id", h.rename)
	f.Post("/:id/move", h.move)
	f.Delete("/:id", h.delete)
	f.Get("/:id/breadcrumb", h.breadcrumb)
	f.Get("/:id/children", h.listFolderChildren)
}

func (h *Handler) create(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	var in CreateFolderInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	f, err := h.svc.Create(c.Context(), p.UserID, spaceID, in)
	if err != nil {
		return err
	}
	return lib.Created(c, f)
}

func (h *Handler) listChildren(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	var parentID *int64
	if pid := c.Query("parent_id"); pid != "" {
		v, err := strconv.ParseInt(pid, 10, 64)
		if err != nil {
			return lib.ErrBadRequest("invalid parent_id")
		}
		parentID = &v
	}
	folders, err := h.svc.ListChildren(c.Context(), p.UserID, spaceID, parentID)
	if err != nil {
		return err
	}
	return lib.OK(c, folders)
}

func (h *Handler) listFolderChildren(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	f, err := h.svc.Get(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	children, err := h.svc.ListChildren(c.Context(), p.UserID, f.SpaceID, &id)
	if err != nil {
		return err
	}
	return lib.OK(c, children)
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
	f, err := h.svc.Get(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, f)
}

func (h *Handler) rename(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in UpdateFolderInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	f, err := h.svc.Rename(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, f)
}

func (h *Handler) move(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in MoveFolderInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	f, err := h.svc.Move(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, f)
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

func (h *Handler) breadcrumb(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	items, err := h.svc.Breadcrumb(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	return lib.OK(c, items)
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
