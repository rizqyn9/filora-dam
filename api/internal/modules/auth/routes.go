package auth

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(app fiber.Router, h *Handler) {
	auth := app.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/register", h.Register)
}

func RegisterProtectedRoutes(api fiber.Router, h *Handler) {
	auth := api.Group("/auth")
	auth.Post("/logout", h.Logout)
	auth.Get("/me", h.Me)
	auth.Post("/change-password", h.ChangePassword)
}
