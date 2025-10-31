// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/daytonaio/runner/cmd/runner/config"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
)

// InitLogging optionally adds OTEL log shipping to the provided slog instance
// If OTEL logging is enabled, it sets up the global slog handler to fanout to both console and OTEL
// Returns a shutdown function (no-op if OTEL is disabled)
func InitLogging(logger *slog.Logger, cfg *config.Config) (func(), error) {
	if !cfg.OtelLoggingEnabled {
		return func() {}, nil
	}

	ctx := context.Background()

	// Create resource with service information
	res, err := getOtelResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create OTLP log exporter
	exporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP log exporter: %w", err)
	}

	// Create LoggerProvider
	lp := otellog.NewLoggerProvider(
		otellog.WithProcessor(otellog.NewBatchProcessor(exporter)),
		otellog.WithResource(res),
	)

	// Set global LoggerProvider
	global.SetLoggerProvider(lp)

	// Create OTEL slog handler
	otelHandler := otelslog.NewHandler("")

	// Wrap OTEL handler with level filter to respect the logger's configured level
	// This ensures that logs exported to OTEL respect the same LOG_LEVEL as console logs
	filteredOtelHandler := &levelFilterHandler{
		handler:  otelHandler,
		minLevel: logger.Handler(),
	}

	// Create fanout handler combining existing logger's handler and filtered OTEL handler
	fanoutHandler := &multiHandler{
		handlers: []slog.Handler{
			logger.Handler(),
			filteredOtelHandler,
		},
	}

	// Set as default logger globally
	slog.SetDefault(slog.New(fanoutHandler))

	// Return shutdown function
	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := lp.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down logger provider: %v\n", err)
		}
	}

	return shutdown, nil
}

// multiHandler implements slog.Handler and forwards logs to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, record.Level) {
			if err := h.Handle(ctx, record); err != nil {
				// Continue with other handlers even if one fails
				fmt.Fprintf(os.Stderr, "Error in log handler: %v\n", err)
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: newHandlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		newHandlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: newHandlers}
}

// levelFilterHandler wraps a handler and filters log records based on the minimum level
// determined by another handler (typically the console handler with LOG_LEVEL configuration)
type levelFilterHandler struct {
	handler  slog.Handler
	minLevel slog.Handler // Use another handler's Enabled method to determine level
}

func (l *levelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	// Delegate to the minLevel handler to determine if this level should be logged
	return l.minLevel.Enabled(ctx, level) && l.handler.Enabled(ctx, level)
}

func (l *levelFilterHandler) Handle(ctx context.Context, record slog.Record) error {
	return l.handler.Handle(ctx, record)
}

func (l *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelFilterHandler{
		handler:  l.handler.WithAttrs(attrs),
		minLevel: l.minLevel.WithAttrs(attrs),
	}
}

func (l *levelFilterHandler) WithGroup(name string) slog.Handler {
	return &levelFilterHandler{
		handler:  l.handler.WithGroup(name),
		minLevel: l.minLevel.WithGroup(name),
	}
}
