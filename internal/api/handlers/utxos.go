package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetUTXOs returns unspent transaction outputs for an address
func (h *Handler) GetUTXOs(c *fiber.Ctx) error {
	address := c.Params("address")

	utxos, err := h.provider.GetUTXOs(address)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(utxos)
}
