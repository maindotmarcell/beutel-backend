package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maindotmarcell/beutel-backend/internal/api/middleware"
)

// GetUTXOs returns unspent transaction outputs for an address
func (h *Handler) GetUTXOs(c *fiber.Ctx) error {
	logCtx := middleware.GetLogContext(c)
	address := c.Params("address")

	// Add request-specific fields to canonical log
	logCtx.Add("address", address)

	utxos, err := h.provider.GetUTXOs(logCtx, address)
	if err != nil {
		logCtx.Add("error", err.Error())
		logCtx.Add("error_type", "upstream_error")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(utxos)
}
