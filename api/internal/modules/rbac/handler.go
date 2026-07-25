package rbac

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/auth"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// Handler exposes RBAC administration endpoints. Every action is guarded by a
// global permission (role:read / role:assign / role:manage) via the authorizer.
type Handler struct {
	svc   *Service
	authz *auth.Authorizer
}

func NewHandler(svc *Service, authz *auth.Authorizer) *Handler {
	return &Handler{svc: svc, authz: authz}
}

func (h *Handler) RegisterRoutes(router fiber.Router, authMW fiber.Handler) {
	g := router.Group("/rbac", authMW)

	g.Get("/roles", h.listRoles)
	g.Post("/roles", h.createRole)
	g.Get("/roles/:id", h.getRole)
	g.Patch("/roles/:id", h.updateRole)
	g.Delete("/roles/:id", h.deleteRole)

	g.Get("/roles/:id/permissions", h.listRolePermissions)
	g.Post("/roles/:id/permissions", h.grantPermission)
	g.Delete("/roles/:id/permissions/:permissionId", h.revokePermission)

	g.Get("/permissions", h.listPermissions)
	g.Post("/permissions", h.createPermission)

	g.Get("/users/:id/roles", h.listUserRoles)
	g.Post("/users/:id/roles", h.assignRole)
	g.Delete("/users/:id/roles/:roleId", h.revokeRole)
}

// require enforces auth + a global permission, returning the principal.
func (h *Handler) require(c fiber.Ctx, resource, action string) (*auth.Principal, error) {
	p := auth.MustPrincipal(c)
	if p == nil {
		return nil, lib.ErrUnauthorized("not authenticated")
	}
	if err := h.authz.Require(c.Context(), p.UserID, resource, action); err != nil {
		return nil, err
	}
	return p, nil
}

// --- roles ---

func (h *Handler) listRoles(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "read"); err != nil {
		return err
	}
	roles, err := h.svc.ListRoles(c.Context())
	if err != nil {
		return err
	}
	return lib.OK(c, roles)
}

func (h *Handler) getRole(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "read"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	role, err := h.svc.GetRole(c.Context(), id)
	if err != nil {
		return err
	}
	return lib.OK(c, role)
}

func (h *Handler) createRole(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	var in CreateRoleInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	role, err := h.svc.CreateRole(c.Context(), in)
	if err != nil {
		return err
	}
	return lib.Created(c, role)
}

func (h *Handler) updateRole(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in UpdateRoleInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	role, err := h.svc.UpdateRole(c.Context(), id, in)
	if err != nil {
		return err
	}
	return lib.OK(c, role)
}

func (h *Handler) deleteRole(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	if err := h.svc.DeleteRole(c.Context(), id); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"deleted": true})
}

// --- permissions ---

func (h *Handler) listPermissions(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "read"); err != nil {
		return err
	}
	perms, err := h.svc.ListPermissions(c.Context())
	if err != nil {
		return err
	}
	return lib.OK(c, perms)
}

func (h *Handler) createPermission(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	var in CreatePermissionInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	perm, err := h.svc.CreatePermission(c.Context(), in)
	if err != nil {
		return err
	}
	return lib.Created(c, perm)
}

// --- grants ---

func (h *Handler) listRolePermissions(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "read"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	perms, err := h.svc.ListRolePermissions(c.Context(), id)
	if err != nil {
		return err
	}
	return lib.OK(c, perms)
}

func (h *Handler) grantPermission(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in GrantInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	if err := h.svc.GrantPermission(c.Context(), id, in); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"granted": true})
}

func (h *Handler) revokePermission(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "manage"); err != nil {
		return err
	}
	roleID, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	permID, err := paramInt64(c, "permissionId")
	if err != nil {
		return err
	}
	if err := h.svc.RevokePermission(c.Context(), roleID, permID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"revoked": true})
}

// --- assignments ---

func (h *Handler) listUserRoles(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "read"); err != nil {
		return err
	}
	id, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	roles, err := h.svc.ListUserRoles(c.Context(), id)
	if err != nil {
		return err
	}
	return lib.OK(c, roles)
}

func (h *Handler) assignRole(c fiber.Ctx) error {
	p, err := h.require(c, "role", "assign")
	if err != nil {
		return err
	}
	userID, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	var in AssignRoleInput
	if err := c.Bind().Body(&in); err != nil {
		return lib.ErrBadRequest("invalid request body").Wrap(err)
	}
	grantedBy := p.UserID
	if err := h.svc.AssignRole(c.Context(), userID, in, &grantedBy); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"assigned": true})
}

func (h *Handler) revokeRole(c fiber.Ctx) error {
	if _, err := h.require(c, "role", "assign"); err != nil {
		return err
	}
	userID, err := paramInt64(c, "id")
	if err != nil {
		return err
	}
	roleID, err := paramInt64(c, "roleId")
	if err != nil {
		return err
	}
	if err := h.svc.RevokeRole(c.Context(), userID, roleID); err != nil {
		return err
	}
	return lib.OK(c, fiber.Map{"revoked": true})
}

func paramInt64(c fiber.Ctx, key string) (int64, error) {
	v, err := strconv.ParseInt(c.Params(key), 10, 64)
	if err != nil {
		return 0, lib.ErrBadRequest("invalid " + key)
	}
	return v, nil
}
