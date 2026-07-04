package dashboard

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
	router.Get("/galleries/:galleryId/dashboard", authMW, h.gallery)
	router.Get("/dashboard/system", authMW, h.system)
}

func (h *Handler) gallery(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	galleryID, err := strconv.ParseInt(c.Params("galleryId"), 10, 64)
	if err != nil {
		return lib.ErrBadRequest("invalid galleryId")
	}
	dash, err := h.svc.Gallery(c.Context(), p.UserID, galleryID)
	if err != nil {
		return err
	}
	return lib.OK(c, dash)
}

func (h *Handler) system(c fiber.Ctx) error {
	p := auth.MustPrincipal(c)
	if p == nil {
		return lib.ErrUnauthorized("not authenticated")
	}
	dash, err := h.svc.System(c.Context(), p.UserID)
	if err != nil {
		return err
	}
	return lib.OK(c, dash)
}
