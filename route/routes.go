package route

import (
	"projek_uas/app/service"
	"projek_uas/config"
	"projek_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	fiberApp *fiber.App,
	cfg *config.Config,
	authService *service.AuthService,
	userService *service.UserService,
	achievementService *service.AchievementService,
) {
	api := fiberApp.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/login", authService.HandleLogin)
	auth.Post("/refresh", authService.HandleRefreshToken)

	// Protected routes
	auth.Get("/profile", middleware.AuthMiddleware(cfg), authService.HandleGetProfile)
	auth.Post("/logout", middleware.AuthMiddleware(cfg), authService.HandleLogout)

	// User management (Admin only)
	users := api.Group("/users", middleware.AuthMiddleware(cfg), middleware.RequirePermission("user:manage"))
	users.Get("/", userService.HandleGetUsers)
	users.Get("/:id", userService.HandleGetUserByID)
	users.Post("/", userService.HandleCreateUser)
	users.Put("/:id", userService.HandleUpdateUser)
	users.Delete("/:id", userService.HandleDeleteUser)

	// Achievements
	achievements := api.Group("/achievements", middleware.AuthMiddleware(cfg))
	achievements.Get("/", achievementService.HandleGetAchievements)
	achievements.Get("/:id", achievementService.HandleGetAchievementByID)
	achievements.Post("/", middleware.RequirePermission("achievement:create"), achievementService.HandleCreateAchievement)
	achievements.Put("/:id", middleware.RequirePermission("achievement:update"), achievementService.HandleUpdateAchievement)
	achievements.Delete("/:id", middleware.RequirePermission("achievement:delete"), achievementService.HandleDeleteAchievement)
	achievements.Post("/:id/submit", middleware.RequireRole("Mahasiswa"), achievementService.HandleSubmitForVerification)
	achievements.Post("/:id/verify", middleware.RequirePermission("achievement:verify"), achievementService.HandleVerifyAchievement)

	// Reports
	reports := api.Group("/reports", middleware.AuthMiddleware(cfg))
	reports.Get("/statistics", achievementService.HandleGetStatistics)
}
