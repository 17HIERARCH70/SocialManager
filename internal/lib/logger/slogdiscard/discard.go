package slogdiscard

import (
	"context"
	"golang.org/x/exp/slog"
)

// NewDiscardLogger creates a logger that discards all logs.
func NewDiscardLogger() *slog.Logger {
	return slog.New(NewDiscardHandler())
}

// DiscardHandler is an empty log handler that discards all logs.
type DiscardHandler struct{}

// NewDiscardHandler instantiates a new DiscardHandler.
func NewDiscardHandler() slog.Handler {
	return &DiscardHandler{}
}

// Handle implements the slog.Handler interface but does nothing.
func (h *DiscardHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

// WithAttrs returns the same handler as it doesn't utilize attributes.
func (h *DiscardHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup returns the same handler as it doesn't utilize grouping.
func (h *DiscardHandler) WithGroup(_ string) slog.Handler {
	return h
}

// Enabled always returns false as all logs are discarded.
func (h *DiscardHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return false
}
