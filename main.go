package main

import (
	"log"
	"projek_uas/config"
)

func main() {
	// Initialize logger
	if err := config.InitLogger(); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	config.LogInfo("Starting Student Achievement System API")

	// Load configuration
	cfg := config.Load()
	config.LogInfo("Configuration loaded successfully")

	// Setup application
	fiberApp, err := config.SetupApp(cfg)
	if err != nil {
		config.LogError("Failed to setup application: %v", err)
		log.Fatalf("Failed to setup application: %v", err)
	}
	defer config.CloseConnections()

	// Register health checks
	config.RegisterHealthChecks(fiberApp)

	// Start server
	port := ":" + cfg.Server.Port
	config.LogInfo("Server starting on port %s", port)
	if err := fiberApp.Listen(port); err != nil {
		config.LogError("Failed to start server: %v", err)
		log.Fatalf("Failed to start server: %v", err)
	}
}
