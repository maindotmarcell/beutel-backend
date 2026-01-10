package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetBalance returns the balance for an address
func (h *Handler) GetBalance(c *fiber.Ctx) error {
	address := c.Params("address")

	balance, err := h.provider.GetBalance(address)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(balance)
}
