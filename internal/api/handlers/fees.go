package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// GetFees returns recommended fee rates
func (h *Handler) GetFees(c *fiber.Ctx) error {
	fees, err := h.provider.GetFeeRates()
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fees)
}
