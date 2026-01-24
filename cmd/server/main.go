package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"github.com/maindotmarcell/beutel-backend/internal/api"
	"github.com/maindotmarcell/beutel-backend/internal/api/middleware"
	"github.com/maindotmarcell/beutel-backend/internal/chain"
	"github.com/maindotmarcell/beutel-backend/internal/chain/mempool"
	"github.com/maindotmarcell/beutel-backend/internal/logging"
)

func main() {
	// Load .env file if present (development)
	_ = godotenv.Load()

	// Initialize structured logger
	logger := logging.NewLogger()

	// Parse network from environment
	network := chain.ParseNetwork(os.Getenv("NETWORK"))

	// Initialize mempool.space client for the configured network
	chainProvider := mempool.NewClient(network)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Beutel Backend",
	})

	// Middleware stack
	app.Use(recover.New())                      // Panic recovery
	app.Use(middleware.CanonicalLog(logger))    // Canonical logging (one log per request)
	app.Use(cors.New())

	// Setup API routes
	api.SetupRoutes(app, chainProvider, network)

	// Get port from environment or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	logger.Info().
		Str("port", port).
		Str("network", string(network)).
		Msg("server starting")

	if err := app.Listen(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("server failed to start")
	}
}
