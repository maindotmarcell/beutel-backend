package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maindotmarcell/beutel-backend/internal/api/middleware"
)

// BroadcastRequest is the request body for broadcasting a transaction
type BroadcastRequest struct {
	TxHex string `json:"txHex"`
}

// BroadcastTx broadcasts a signed transaction to the network
func (h *Handler) BroadcastTx(c *fiber.Ctx) error {
	logCtx := middleware.GetLogContext(c)

	var req BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		logCtx.Add("error", "invalid request body")
		logCtx.Add("error_type", "validation_error")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.TxHex == "" {
		logCtx.Add("error", "txHex is required")
		logCtx.Add("error_type", "validation_error")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "txHex is required",
		})
	}

	txid, err := h.provider.BroadcastTx(logCtx, req.TxHex)
	if err != nil {
		logCtx.Add("error", err.Error())
		logCtx.Add("error_type", "upstream_error")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"txid": txid,
	})
}
