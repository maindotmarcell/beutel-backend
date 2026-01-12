package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/maindotmarcell/beutel-backend/internal/api"
	"github.com/maindotmarcell/beutel-backend/internal/chain"
	"github.com/maindotmarcell/beutel-backend/internal/chain/mempool"
)

func main() {
	// Load .env file if present (development)
	_ = godotenv.Load()

	// Parse network from environment
	network := chain.ParseNetwork(os.Getenv("NETWORK"))

	// Initialize mempool.space client for the configured network
	chainProvider := mempool.NewClient(network)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Beutel Backend",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Health check endpoint (includes network info)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"network": network,
		})
	})

	// Setup API routes
	api.SetupRoutes(app, chainProvider)

	// Get port from environment or default to 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🚀 Beutel backend starting on port %s (network: %s)", port, network)
	log.Fatal(app.Listen(":" + port))
}
