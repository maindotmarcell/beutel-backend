package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// BroadcastRequest is the request body for broadcasting a transaction
type BroadcastRequest struct {
	TxHex string `json:"txHex"`
}

// BroadcastTx broadcasts a signed transaction to the network
func (h *Handler) BroadcastTx(c *fiber.Ctx) error {
	var req BroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.TxHex == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "txHex is required",
		})
	}

	txid, err := h.provider.BroadcastTx(req.TxHex)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"txid": txid,
	})
}
