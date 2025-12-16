package middleware

import (
	"projek_uas/config"
	"projek_uas/helper"

	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware is an alias for AuthMiddleware for consistency
func JWTMiddleware(cfg *config.Config) fiber.Handler {
	return AuthMiddleware(cfg)
}
