package middleware

import (
	"projek_uas/helper"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(jwtSecret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Missing authorization header")
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := helper.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			return helper.ErrorResponse(c, fiber.StatusUnauthorized, "Invalid token")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("roleID", claims.RoleID)
		c.Locals("roleName", claims.RoleName)
		c.Locals("permissions", claims.Permissions)

		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Insufficient permissions")
		}

		for _, p := range permissions {
			if p == permission {
				return c.Next()
			}
		}

		return helper.ErrorResponse(c, fiber.StatusForbidden, "Insufficient permissions")
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleName, ok := c.Locals("roleName").(string)
		if !ok {
			return helper.ErrorResponse(c, fiber.StatusForbidden, "Unauthorized")
		}

		for _, role := range roles {
			if role == roleName {
				return c.Next()
			}
		}

		return helper.ErrorResponse(c, fiber.StatusForbidden, "Unauthorized")
	}
}
