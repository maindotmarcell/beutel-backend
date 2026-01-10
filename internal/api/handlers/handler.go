package handlers

import "github.com/maindotmarcell/beutel-backend/internal/chain"

// Handler holds dependencies for HTTP handlers
type Handler struct {
	provider chain.Provider
}

// NewHandler creates a new Handler with the given chain provider
func NewHandler(provider chain.Provider) *Handler {
	return &Handler{provider: provider}
}
