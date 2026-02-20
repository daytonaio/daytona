// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// MultiHandler implements slog.Handler and forwards logs to multiple handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, record.Level) {
			if err := h.Handle(ctx, record); err != nil {
				fmt.Fprintf(os.Stderr, "Error in log handler: %v\n", err)
			}
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: newHandlers}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: newHandlers}
}

// LevelFilterHandler wraps a handler and filters log records based on the minimum level
type LevelFilterHandler struct {
	handler         slog.Handler
	minLevelHandler slog.Handler
}

func NewLevelFilterHandler(handler slog.Handler, minLevelHandler slog.Handler) *LevelFilterHandler {
	return &LevelFilterHandler{
		handler:         handler,
		minLevelHandler: minLevelHandler,
	}
}

func (l *LevelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return l.minLevelHandler.Enabled(ctx, level) && l.handler.Enabled(ctx, level)
}

func (l *LevelFilterHandler) Handle(ctx context.Context, record slog.Record) error {
	return l.handler.Handle(ctx, record)
}

func (l *LevelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelFilterHandler{
		handler:         l.handler.WithAttrs(attrs),
		minLevelHandler: l.minLevelHandler.WithAttrs(attrs),
	}
}

func (l *LevelFilterHandler) WithGroup(name string) slog.Handler {
	return &LevelFilterHandler{
		handler:         l.handler.WithGroup(name),
		minLevelHandler: l.minLevelHandler.WithGroup(name),
	}
}
