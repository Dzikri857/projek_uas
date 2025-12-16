package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
)

// CSRFMiddleware provides CSRF protection
func CSRFMiddleware() fiber.Handler {
	return csrf.New(csrf.Config{
		KeyLookup:      "header:X-CSRF-Token",
		CookieName:     "csrf_",
		CookieSameSite: "Strict",
		Expiration:     3600, // 1 hour
		KeyGenerator:   func() string { return "random-csrf-key" },
	})
}
