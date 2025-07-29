package api

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"sync"
	"testing"
)

type LogEntry struct {
	Level   slog.Level
	Message string
	Attrs   []slog.Attr
}

type TestLogHandler struct {
	mu      sync.Mutex
	entries []LogEntry
}

func NewTestLog() *TestLogHandler {
	return &TestLogHandler{}
}

func (h *TestLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return true
}

func (h *TestLogHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var attrs []slog.Attr
	r.Attrs(func(a slog.Attr) bool {
		attrs = append(attrs, a)
		return true
	})

	h.entries = append(h.entries, LogEntry{
		Level:   r.Level,
		Message: r.Message,
		Attrs:   attrs,
	})
	return nil
}

func (h *TestLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *TestLogHandler) WithGroup(name string) slog.Handler {
	return h
}

func (h *TestLogHandler) Entries() []LogEntry {
	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]LogEntry(nil), h.entries...)
}

func (h *TestLogHandler) assertLog(t *testing.T, expected string) {
	entries := h.Entries()
	found := false
	for _, entry := range entries {
		if entry.Message == expected {
			found = true
			break
		}
	}
	assert.True(t, found, "expected log to contain %s", expected)
}
