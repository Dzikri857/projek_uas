package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
)

// HelmetMiddleware adds security headers
func HelmetMiddleware() fiber.Handler {
	return helmet.New()
}
