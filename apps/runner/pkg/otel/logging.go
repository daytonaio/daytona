// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-isatty"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
)

// newConsoleHandler always returns a tint handler writing to stdout. Colors
// are suppressed when stdout is not a terminal (i.e. in production / pods),
// which keeps the on-the-wire format byte-identical to the pre-pkg/otel
// runner. This is intentional: external log pipelines (Promtail, Fluent Bit,
// etc.) are configured against tint's "2006-01-02T15:04:05Z INF msg key=v"
// shape. Switching to slog.NewTextHandler would emit "time=… level=… msg=…"
// which silently breaks those parsers.
func newConsoleHandler(level slog.Level) slog.Handler {
	return tint.NewHandler(os.Stdout, &tint.Options{
		Level:      level,
		TimeFormat: time.RFC3339,
		NoColor:    !isatty.IsTerminal(os.Stdout.Fd()) && !isatty.IsCygwinTerminal(os.Stdout.Fd()),
	})
}

func initConsoleLogging(level slog.Level) {
	slog.SetDefault(slog.New(newConsoleHandler(level)))
}

// InitConsoleLogging installs a console-only slog handler as the default logger.
// Useful for emitting log lines before the full Init() call (e.g. while loading
// the runner config). The full Init() will reinstall the same handler and
// optionally fan out to OTel.
func InitConsoleLogging(level slog.Level) {
	initConsoleLogging(level)
}

func initOTelLogging(ctx context.Context, res *resource.Resource, serviceName string, level slog.Level) (func(context.Context) error, error) {
	exp, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return noop, fmt.Errorf("otel: create log exporter: %w", err)
	}

	lp := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exp)),
		sdklog.WithResource(res),
	)
	global.SetLoggerProvider(lp)

	consoleHandler := newConsoleHandler(level)
	// otelslog's handler does not honour LOG_LEVEL on its own, so wrap it to
	// drop records below the configured level. Without this, OTLP export would
	// receive sub-LOG_LEVEL records (e.g. debug logs when LOG_LEVEL=info) that
	// the console handler already suppresses.
	otelHandler := &leveledHandler{
		handler: otelslog.NewHandler(serviceName, otelslog.WithLoggerProvider(lp)),
		level:   level,
	}

	slog.SetDefault(slog.New(&fanoutHandler{handlers: []slog.Handler{consoleHandler, otelHandler}}))

	return lp.Shutdown, nil
}

// leveledHandler gates a delegate slog.Handler by a minimum level. It exists
// because some handlers (e.g. otelslog) don't apply slog level configuration
// themselves. The fanoutHandler checks Enabled before dispatching, so the
// level gate here keeps below-threshold records out of the delegate entirely.
type leveledHandler struct {
	handler slog.Handler
	level   slog.Level
}

func (h *leveledHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return l >= h.level && h.handler.Enabled(ctx, l)
}

func (h *leveledHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *leveledHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &leveledHandler{handler: h.handler.WithAttrs(attrs), level: h.level}
}

func (h *leveledHandler) WithGroup(name string) slog.Handler {
	return &leveledHandler{handler: h.handler.WithGroup(name), level: h.level}
}

type fanoutHandler struct {
	handlers []slog.Handler
}

func (f *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range f.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (f *fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for _, h := range f.handlers {
		if h.Enabled(ctx, r.Level) {
			errs = append(errs, h.Handle(ctx, r.Clone()))
		}
	}
	return errors.Join(errs...)
}

func (f *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: hs}
}

func (f *fanoutHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(f.handlers))
	for i, h := range f.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &fanoutHandler{handlers: hs}
}
