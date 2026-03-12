package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/prabalesh/zenko-backend/internal/api/handlers"
	"github.com/prabalesh/zenko-backend/internal/api/middleware"
	"github.com/prabalesh/zenko-backend/internal/config"
	"github.com/prabalesh/zenko-backend/internal/services/auth"
)

func SetupRouter(
	cfg *config.Config,
	authHandler *handlers.AuthHandler,
	profileHandler *handlers.ProfileHandler,
	jwtService auth.JWTService,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	// Global Middleware
	app.Use(middleware.Logger())
	app.Use(middleware.RateLimit())

	// Health check
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
		})
	})

	// API routes
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Auth routes
	authGroup := v1.Group("/auth")
	authGroup.Get("/google", authHandler.GoogleLogin)
	authGroup.Get("/google/callback", authHandler.GoogleCallback)
	authGroup.Post("/refresh", authHandler.Refresh)
	authGroup.Post("/logout", authHandler.Logout)

	// Protected routes
	protected := v1.Group("/", middleware.Auth(jwtService))
	protected.Get("/me", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{"status": "authenticated", "user_id": userID})
	})

	// Profile routes
	profileGroup := protected.Group("/profile")
	profileGroup.Post("/username", profileHandler.SetUsername)
	profileGroup.Get("/", profileHandler.GetProfile)
	profileGroup.Patch("/", profileHandler.UpdateProfile)
	profileGroup.Patch("/username", profileHandler.ChangeUsername)
	profileGroup.Patch("/social-links", profileHandler.UpdateSocialLinks)

	// Public profile route
	v1.Get("/profile/:username", profileHandler.GetPublicProfile)

	return app
}
