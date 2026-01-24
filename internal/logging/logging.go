package logging

import (
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// LogContext accumulates fields throughout a request lifecycle.
// It is safe for concurrent use.
type LogContext struct {
	mu     sync.RWMutex
	fields map[string]interface{}
}

// NewLogContext creates a new context for accumulating log fields
func NewLogContext() *LogContext {
	return &LogContext{
		fields: make(map[string]interface{}),
	}
}

// Add adds a field to the log context (thread-safe)
func (lc *LogContext) Add(key string, value interface{}) {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.fields[key] = value
}

// Get retrieves a field from the log context
func (lc *LogContext) Get(key string) (interface{}, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	v, ok := lc.fields[key]
	return v, ok
}

// Fields returns a copy of all accumulated fields
func (lc *LogContext) Fields() map[string]interface{} {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	copy := make(map[string]interface{}, len(lc.fields))
	for k, v := range lc.fields {
		copy[k] = v
	}
	return copy
}

// NewLogger creates the application logger configured for structured JSON output
func NewLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = time.RFC3339
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
