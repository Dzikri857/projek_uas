package route

import (
	"projek_uas/app/repository"
	"projek_uas/app/service"
	"projek_uas/middleware"

	"github.com/gofiber/fiber/v2"
)

func Setup(
	fiberApp *fiber.App,
	jwtSecret string,
	authService *service.AuthService,
	userRepo *repository.UserRepository,
	achievementRepo *repository.AchievementRepository,
	studentRepo *repository.StudentRepository,
	lecturerRepo *repository.LecturerRepository,
) {
	api := fiberApp.Group("/api/v1")

	// Public routes
	auth := api.Group("/auth")
	auth.Post("/login", authService.HandleLoginHTTP)
	auth.Post("/refresh", authService.HandleRefreshTokenHTTP)

	// Protected routes
	auth.Get("/profile", middleware.AuthMiddleware(jwtSecret), authService.HandleGetProfileHTTP)
	auth.Post("/logout", middleware.AuthMiddleware(jwtSecret), authService.HandleLogoutHTTP)

	// User management (Admin only)
	users := api.Group("/users", middleware.AuthMiddleware(jwtSecret), middleware.RequirePermission("user:manage"))
	users.Get("/", userRepo.HandleGetAllHTTP)
	users.Get("/:id", userRepo.HandleGetByIDHTTP)
	users.Post("/", func(c *fiber.Ctx) error {
		return userRepo.HandleCreateHTTP(c, studentRepo, lecturerRepo)
	})
	users.Put("/:id", userRepo.HandleUpdateHTTP)
	users.Delete("/:id", userRepo.HandleDeleteHTTP)

	// Achievements
	achievements := api.Group("/achievements", middleware.AuthMiddleware(jwtSecret))
	achievements.Get("/", func(c *fiber.Ctx) error {
		return achievementRepo.HandleGetAllHTTP(c, studentRepo, lecturerRepo)
	})
	achievements.Get("/:id", func(c *fiber.Ctx) error {
		return achievementRepo.HandleGetByIDHTTP(c, studentRepo, lecturerRepo)
	})
	achievements.Post("/", middleware.RequirePermission("achievement:create"), func(c *fiber.Ctx) error {
		return achievementRepo.HandleCreateHTTP(c, studentRepo)
	})
	achievements.Put("/:id", middleware.RequirePermission("achievement:update"), func(c *fiber.Ctx) error {
		return achievementRepo.HandleUpdateHTTP(c, studentRepo)
	})
	achievements.Delete("/:id", middleware.RequirePermission("achievement:delete"), func(c *fiber.Ctx) error {
		return achievementRepo.HandleDeleteHTTP(c, studentRepo)
	})
	achievements.Post("/:id/submit", middleware.RequireRole("Mahasiswa"), func(c *fiber.Ctx) error {
		return achievementRepo.HandleSubmitHTTP(c, studentRepo)
	})
	achievements.Post("/:id/verify", middleware.RequirePermission("achievement:verify"), achievementRepo.HandleVerifyHTTP)

	// Reports
	reports := api.Group("/reports", middleware.AuthMiddleware(jwtSecret))
	reports.Get("/statistics", func(c *fiber.Ctx) error {
		return achievementRepo.HandleStatisticsHTTP(c, studentRepo, lecturerRepo)
	})
}
