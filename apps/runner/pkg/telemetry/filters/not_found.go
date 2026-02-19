// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package filters

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type NotFoundExporterFilter struct{}

// Apply wraps the given exporter so that HTTP 404 spans recorded as errors
// have their status downgraded to Unset before export. The spans themselves
// are still exported, preserving trace continuity.
func (f *NotFoundExporterFilter) Apply(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return &notFoundStatusExporter{next: exporter}
}

// notFoundStatusExporter adjusts HTTP 404 spans that were classified as errors
// so they are exported with status Unset instead of Error. This handles optimistic
// error handling patterns where 404 responses are expected (e.g., checking if a
// resource exists before creating it).
type notFoundStatusExporter struct {
	next sdktrace.SpanExporter
}

func (e *notFoundStatusExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	adjusted := make([]sdktrace.ReadOnlySpan, len(spans))
	for i, span := range spans {
		if span.Status().Code == codes.Error && isHTTP404(span) {
			adjusted[i] = &statusOverrideSpan{
				ReadOnlySpan: span,
				status:       sdktrace.Status{Code: codes.Unset},
			}
		} else {
			adjusted[i] = span
		}
	}
	return e.next.ExportSpans(ctx, adjusted)
}

func (e *notFoundStatusExporter) Shutdown(ctx context.Context) error {
	return e.next.Shutdown(ctx)
}

// statusOverrideSpan wraps a ReadOnlySpan and replaces its Status.
type statusOverrideSpan struct {
	sdktrace.ReadOnlySpan
	status sdktrace.Status
}

func (s *statusOverrideSpan) Status() sdktrace.Status {
	return s.status
}

// isHTTP404 reports whether a span represents an HTTP 404 response.
// It checks both the current semconv key (http.response.status_code, semconv v1.20+)
// and the legacy key (http.status_code) for compatibility with older instrumentation.
func isHTTP404(s sdktrace.ReadOnlySpan) bool {
	for _, attr := range s.Attributes() {
		if attr.Key == semconv.HTTPResponseStatusCodeKey || attr.Key == attribute.Key("http.status_code") {
			return attr.Value.AsInt64() == 404
		}
	}
	return false
}
