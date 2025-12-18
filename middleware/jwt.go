package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// JWTMiddleware is an alias for AuthMiddleware for consistency
func JWTMiddleware(jwtSecret string) fiber.Handler {
	return AuthMiddleware(jwtSecret)
}
