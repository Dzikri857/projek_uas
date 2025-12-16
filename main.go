package main

import (
	"log"
	"projek_uas/config"
	"projek_uas/database"
	"projek_uas/app/repository"
	"projek_uas/app/service"
	"projek_uas/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to databases
	if err := database.ConnectPostgres(cfg); err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer database.ClosePostgres()

	if err := database.ConnectMongoDB(cfg); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer database.CloseMongoDB()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	studentRepo := repository.NewStudentRepository()
	lecturerRepo := repository.NewLecturerRepository()
	achievementRepo := repository.NewAchievementRepository()

	authService := service.NewAuthService(userRepo, cfg)
	userService := service.NewUserService(userRepo, studentRepo, lecturerRepo)
	achievementService := service.NewAchievementService(achievementRepo, studentRepo, lecturerRepo)

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// Middleware
	fiberApp.Use(cors.New())
	fiberApp.Use(logger.New())
	fiberApp.Use(recover.New())

	// Health check
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Student Achievement System API",
			"version": "1.0",
		})
	})

	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "success",
			"postgres": "connected",
			"mongodb":  "connected",
		})
	})

	route.Setup(fiberApp, cfg, authService, userService, achievementService)

	// Start server
	port := ":" + cfg.Server.Port
	log.Printf("Server starting on port %s", port)
	if err := fiberApp.Listen(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
