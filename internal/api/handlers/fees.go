package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maindotmarcell/beutel-backend/internal/api/middleware"
)

// GetFees returns recommended fee rates
func (h *Handler) GetFees(c *fiber.Ctx) error {
	logCtx := middleware.GetLogContext(c)

	fees, err := h.provider.GetFeeRates(logCtx)
	if err != nil {
		logCtx.Add("error", err.Error())
		logCtx.Add("error_type", "upstream_error")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fees)
}
