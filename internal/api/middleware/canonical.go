package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/maindotmarcell/beutel-backend/internal/logging"
)

// LogContextKey is the key used to store LogContext in fiber.Locals
const LogContextKey = "logContext"

// CanonicalLog is a middleware that emits a single structured log line per request.
// It creates a LogContext at the start of the request, which handlers and services
// can use to add fields. At the end of the request, all accumulated fields are
// logged in one "canonical" log entry.
func CanonicalLog(logger zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		requestID := uuid.New().String()

		// Create log context and store in fiber locals
		logCtx := logging.NewLogContext()
		logCtx.Add("request_id", requestID)
		logCtx.Add("method", c.Method())
		logCtx.Add("path", c.Path())
		c.Locals(LogContextKey, logCtx)

		// Set request ID header for client-side tracing
		c.Set("X-Request-ID", requestID)

		// Process request
		err := c.Next()

		// Collect final fields after request processing
		logCtx.Add("status", c.Response().StatusCode())
		logCtx.Add("duration_ms", time.Since(start).Milliseconds())

		// Add route pattern if available (e.g., "/v1/address/:address/balance")
		if route := c.Route(); route != nil && route.Path != "" {
			logCtx.Add("route", route.Path)
		}

		// Add response size
		logCtx.Add("response_bytes", len(c.Response().Body()))

		// Determine log level based on status code
		status := c.Response().StatusCode()
		var event *zerolog.Event
		if status >= 500 {
			event = logger.Error()
		} else if status >= 400 {
			event = logger.Warn()
		} else {
			event = logger.Info()
		}

		// Emit single canonical log with all accumulated fields
		for k, v := range logCtx.Fields() {
			event = event.Interface(k, v)
		}
		event.Msg("request")

		return err
	}
}

// GetLogContext retrieves the LogContext from fiber locals.
// Returns a new empty LogContext if none is found (fallback for safety).
func GetLogContext(c *fiber.Ctx) *logging.LogContext {
	if lc, ok := c.Locals(LogContextKey).(*logging.LogContext); ok {
		return lc
	}
	return logging.NewLogContext()
}
