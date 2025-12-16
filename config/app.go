package config

import (
	"projek_uas/app/repository"
	"projek_uas/app/service"
	"projek_uas/database"
	"projek_uas/middleware"
	"projek_uas/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupApp creates and configures the Fiber application
func SetupApp(cfg *Config) (*fiber.App, error) {
	// Connect to databases
	if err := database.ConnectPostgres(cfg); err != nil {
		LogError("Failed to connect to PostgreSQL: %v", err)
		return nil, err
	}

	if err := database.ConnectMongoDB(cfg); err != nil {
		LogError("Failed to connect to MongoDB: %v", err)
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	studentRepo := repository.NewStudentRepository()
	lecturerRepo := repository.NewLecturerRepository()
	achievementRepo := repository.NewAchievementRepository()

	// Initialize services
	authService := service.NewAuthService(userRepo, cfg)

	// Create Fiber app
	fiberApp := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			LogError("Request error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// Register middlewares
	RegisterMiddleware(fiberApp)

	// Register routes
	route.Setup(fiberApp, cfg, authService, userRepo, achievementRepo, studentRepo, lecturerRepo)

	LogInfo("Application setup completed successfully")
	return fiberApp, nil
}

// RegisterMiddleware registers all middleware for the Fiber app
func RegisterMiddleware(fiberApp *fiber.App) {
	// CORS middleware
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	// Helmet middleware for security headers
	fiberApp.Use(helmet.New())

	// Logger middleware
	fiberApp.Use(logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	// Recover middleware for panic recovery
	fiberApp.Use(recover.New())

	// CSRF protection (optional, can be enabled if needed)
	// fiberApp.Use(csrf.New())

	LogInfo("Middleware registered successfully")
}

// RegisterHealthChecks registers health check endpoints
func RegisterHealthChecks(fiberApp *fiber.App) {
	// Root health check
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Student Achievement System API",
			"version": "1.0",
		})
	})

	// Detailed health check
	fiberApp.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":   "success",
			"postgres": "connected",
			"mongodb":  "connected",
		})
	})

	LogInfo("Health check routes registered")
}

// CloseConnections closes all database connections
func CloseConnections() {
	database.ClosePostgres()
	database.CloseMongoDB()
	LogInfo("Database connections closed")
}
