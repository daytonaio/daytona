// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestToSnakeCase(t *testing.T) {
	assert.Equal(t, "sandbox_wait_for_start", toSnakeCase("Sandbox.WaitForStart"))
	assert.Equal(t, "httpserver", toSnakeCase("HTTPServer"))
}

func TestShutdownOtelNil(t *testing.T) {
	assert.NoError(t, shutdownOtel(context.Background(), nil))
}

func TestWithInstrumentationNilState(t *testing.T) {
	value, err := withInstrumentation(context.Background(), nil, "Comp", "Method", func(ctx context.Context) (string, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", value)
}

func TestOtelTransportInjectsHeaders(t *testing.T) {
	tp := trace.NewTracerProvider()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	ctx, span := tp.Tracer("test").Start(context.Background(), "span")
	defer span.End()

	transport := &otelTransport{base: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.NotEmpty(t, req.Header.Get("Traceparent"))
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody, Header: http.Header{}}, nil
	})}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)
	_, err = transport.RoundTrip(req)
	require.NoError(t, err)
}
