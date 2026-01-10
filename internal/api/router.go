package api

import (
	"github.com/maindotmarcell/beutel-backend/internal/api/handlers"
	"github.com/maindotmarcell/beutel-backend/internal/chain"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all API routes
func SetupRoutes(app *fiber.App, provider chain.Provider) {
	h := handlers.NewHandler(provider)

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
