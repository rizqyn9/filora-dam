package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/rizqynugroho9/filora-dam/api/internal/config"
	"github.com/rizqynugroho9/filora-dam/api/internal/database"
	"github.com/rizqynugroho9/filora-dam/api/internal/lib"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/account"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/gallery"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/rbac"
	"github.com/rizqynugroho9/filora-dam/api/internal/modules/session"
)

// Deps are the dependencies required to build the HTTP server. As modules are
// added, their handlers are added here and wired in the compose root.
type Deps struct {
	Config  *config.Config
	DB      *database.DB
	AuthMW  fiber.Handler
	Account *account.Handler
	RBAC    *rbac.Handler
	Session *session.Handler
	Gallery *gallery.Handler
}

// New builds the Fiber application with global middleware and routes.
func New(deps Deps) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Filora DAM API",
		ErrorHandler: errorHandler,
	})

	app.Use(requestid.New())
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	registerRoutes(app, deps)

	return app
}

// errorHandler converts any error into the standard error envelope.
func errorHandler(c fiber.Ctx, err error) error {
	status, code, message := lib.HTTPError(err)
	return lib.Fail(c, status, code, message)
}
