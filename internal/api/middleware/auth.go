package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/prabalesh/zenko-backend/internal/pkg/errors"
	"github.com/prabalesh/zenko-backend/internal/services/auth"
)

func Auth(jwtService auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return errors.Unauthorized("Missing authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return errors.Unauthorized("Invalid authorization header format")
		}

		tokenStr := parts[1]
		claims, err := jwtService.ValidateAccessToken(tokenStr)
		if err != nil {
			return errors.Unauthorized("Invalid or expired token: " + err.Error())
		}

		c.Locals("user_id", (*claims)["sub"])
		c.Locals("username", (*claims)["username"])

		return c.Next()
	}
}
