package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maindotmarcell/beutel-backend/internal/api/handlers"
	"github.com/maindotmarcell/beutel-backend/internal/chain"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, provider chain.Provider, network chain.Network) {
	h := handlers.NewHandler(provider)

	// Health check endpoint (includes network info)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"network": network,
		})
	})

	// API v1 group
	v1 := app.Group("/v1")

	// Address endpoints
	v1.Get("/address/:address/balance", h.GetBalance)
	v1.Get("/address/:address/utxos", h.GetUTXOs)
	v1.Get("/address/:address/transactions", h.GetTransactions)

	// Fee rates
	v1.Get("/fees", h.GetFees)

	// Transaction broadcast
	v1.Post("/tx/broadcast", h.BroadcastTx)
}
