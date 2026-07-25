package asset

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
	// Space-scoped asset operations.
	g := router.Group("/spaces/:spaceId/assets", authMW)
	g.Post("/", h.upload)
	g.Get("/", h.list)
	g.Get("/all", h.listAll)
	g.Get("/search", h.search)
	g.Get("/trash", h.trash)
	g.Get("/filter/:type", h.filterByType)

	// Asset-scoped operations.
	a := router.Group("/assets", authMW)
	a.Get("/:id", h.get)
	a.Patch("/:id", h.update)
	a.Post("/:id/move", h.move)
	a.Delete("/:id", h.delete)
	a.Post("/:id/restore", h.restore)
	a.Get("/:id/download", h.download)
}

func (h *Handler) upload(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}

	// Optional folder_id from form field.
	var folderID *int64
	if fid := c.FormValue("folder_id"); fid != "" {
		v, err := strconv.ParseInt(fid, 10, 64)
		if err != nil {
			return lib.ErrBadRequest("invalid folder_id")
		}
		folderID = &v
	}

	fh, err := c.FormFile("file")
	if err != nil {
		return lib.ErrBadRequest("file is required (multipart field 'file')")
	}
	f, err := fh.Open()
	if err != nil {
		return lib.ErrBadRequest("could not read uploaded file").Wrap(err)
	}
	defer f.Close()

	hash, err := lib.HashReader(f)
	if err != nil {
		return lib.ErrInternal("failed to hash file").Wrap(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return lib.ErrInternal("failed to rewind file").Wrap(err)
	}

	head := make([]byte, 512)
	n, _ := f.Read(head)
	mime := lib.DetectContentType(head[:n])
	if _, err := f.Seek(0, 0); err != nil {
		return lib.ErrInternal("failed to rewind file").Wrap(err)
	}

	a, err := h.svc.Upload(c.Context(), p.UserID, UploadInput{
		SpaceID:  spaceID,
		FolderID: folderID,
		Name:     fh.Filename,
		Type:     lib.ClassifyType(mime),
		MimeType: mime,
		Size:     fh.Size,
		Hash:     hash,
		Reader:   f,
	})
	if err != nil {
		return err
	}
	return lib.Created(c, a)
}

func (h *Handler) list(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}

	// Optional folder_id query param. If absent, list root assets.
	var folderID *int64
	if fid := c.Query("folder_id"); fid != "" {
		v, err := strconv.ParseInt(fid, 10, 64)
		if err != nil {
			return lib.ErrBadRequest("invalid folder_id")
		}
		folderID = &v
	}

	res, err := h.svc.List(c.Context(), p.UserID, spaceID, folderID, lib.ParsePage(c))
	if err != nil {
		return err
	}
	return lib.OK(c, res)
}

func (h *Handler) listAll(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	res, err := h.svc.ListAll(c.Context(), p.UserID, spaceID, lib.ParsePage(c))
	if err != nil {
		return err
	}
	return lib.OK(c, res)
}

func (h *Handler) search(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	q := c.Query("q")
	if q == "" {
		return lib.ErrBadRequest("query parameter 'q' is required")
	}
	res, err := h.svc.Search(c.Context(), p.UserID, spaceID, q, lib.ParsePage(c))
	if err != nil {
		return err
	}
	return lib.OK(c, res)
}

func (h *Handler) filterByType(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	res, err := h.svc.FilterByType(c.Context(), p.UserID, spaceID, c.Params("type"), lib.ParsePage(c))
	if err != nil {
		return err
	}
	return lib.OK(c, res)
}

func (h *Handler) trash(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	spaceID, err := paramInt64(c, "spaceId")
	if err != nil {
		return err
	}
	res, err := h.svc.Trash(c.Context(), p.UserID, spaceID, lib.ParsePage(c))
	if err != nil {
		return err
	}
	return lib.OK(c, res)
}

func (h *Handler) get(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
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
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return err
	}
	var in UpdateAssetInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	a, err := h.svc.UpdateName(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, a)
}

func (h *Handler) move(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return err
	}
	var in MoveAssetInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	a, err := h.svc.Move(c.Context(), p.UserID, id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, a)
}

func (h *Handler) delete(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Context(), p.UserID, id); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"deleted": true})
}

func (h *Handler) restore(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.Restore(c.Context(), p.UserID, id); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"restored": true})
}

func (h *Handler) download(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	id, err := paramUUID(c, "id")
	if err != nil {
		return err
	}
	url, err := h.svc.DownloadURL(c.Context(), p.UserID, id)
	if err != nil {
		return err
	}
	c.Set(fiber.HeaderLocation, url)
	return c.SendStatus(fiber.StatusFound)
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}

func paramUUID(c fiber.Ctx, key string) (uuid.UUID, error) {
	v, err := uuid.Parse(c.Params(key))
	if err != nil {
		return uuid.UUID{}, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
