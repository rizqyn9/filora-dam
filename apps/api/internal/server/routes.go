package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
)

// registerRoutes mounts all HTTP routes. Module route groups are added here as
// modules are built.
func registerRoutes(app *fiber.App, deps Deps) {
	app.Get("/", func(c fiber.Ctx) error {
		return lib.OK(c, fiber.Map{
			"name":    "Filora DAM API",
			"version": "0.1.0",
			"status":  "ok",
		})
	})

	app.Get("/health", func(c fiber.Ctx) error {
		dbStatus := "ok"
		if deps.DB != nil {
			ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
			defer cancel()
			if err := deps.DB.Pool.Ping(ctx); err != nil {
				dbStatus = "down"
			}
		} else {
			dbStatus = "unconfigured"
		}

		return lib.OK(c, fiber.Map{
			"status":   "ok",
			"database": dbStatus,
		})
	})

	// Versioned API
	v1 := app.Group("/api/v1")
	if deps.Account != nil {
		deps.Account.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.RBAC != nil {
		deps.RBAC.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.Session != nil {
		deps.Session.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.Gallery != nil {
		deps.Gallery.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.Album != nil {
		deps.Album.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.Tag != nil {
		deps.Tag.RegisterRoutes(v1, deps.AuthMW)
	}
	if deps.Storage != nil {
		deps.Storage.RegisterRoutes(v1, deps.AuthMW)
	}
}
