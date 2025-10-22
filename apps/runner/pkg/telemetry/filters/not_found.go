// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package filters

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type NotFoundExporterFilter struct{}

// Custom exporter filter to ignore 404 errors
func (f *NotFoundExporterFilter) Apply(exporter sdktrace.SpanExporter) sdktrace.SpanExporter {
	return &filteredNotFoundExporter{
		next: exporter,
	}
}

// filteredNotFoundExporter filters out HTTP 404 errors from being marked as errors in traces.
// This is useful for optimistic error handling patterns where 404 responses are expected
// (e.g., checking if a resource exists before creating it).
type filteredNotFoundExporter struct {
	next sdktrace.SpanExporter
}

func (e *filteredNotFoundExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	// Filter out spans with 404 errors - they're expected in optimistic error handling
	filteredSpans := make([]sdktrace.ReadOnlySpan, 0, len(spans))
	for _, span := range spans {
		// Skip spans that have 404 errors - they're part of normal optimistic flow
		if span.Status().Code == codes.Error && e.is404Error(span) {
			// Don't export this span - it's an expected condition, not an error
			continue
		}
		filteredSpans = append(filteredSpans, span)
	}
	return e.next.ExportSpans(ctx, filteredSpans)
}

func (e *filteredNotFoundExporter) is404Error(s sdktrace.ReadOnlySpan) bool {
	// Check for HTTP 404 status code
	for _, attr := range s.Attributes() {
		if attr.Key == "http.status_code" {
			statusCode := attr.Value.AsInterface()
			// Check if status code is 404 (not found)
			if statusCode == int64(404) || statusCode == 404 {
				return true
			}
		}
	}
	return false
}

func (e *filteredNotFoundExporter) Shutdown(ctx context.Context) error {
	return e.next.Shutdown(ctx)
}
