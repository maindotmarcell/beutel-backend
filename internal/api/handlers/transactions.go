package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetTransactions returns transaction history for an address
func (h *Handler) GetTransactions(c *fiber.Ctx) error {
	address := c.Params("address")

	txs, err := h.provider.GetTransactions(address)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(txs)
}
