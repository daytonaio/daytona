// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errorEnvelope mirrors the shared response shape stamped by Daytona services
// when an HTTP request fails. `source` and `code` are optional on the wire
// (`code` is only set when the error needs to be distinguished beyond its
// HTTP status; `source` is always set by the new middleware but may be absent
// on legacy paths until they're migrated).
type errorEnvelope struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Source     string `json:"source,omitempty"`
	Code       string `json:"code,omitempty"`
	Timestamp  string `json:"timestamp,omitempty"`
	Path       string `json:"path,omitempty"`
	Method     string `json:"method,omitempty"`
	// Legacy field preserved for backward compatibility (NestJS API only).
	Error string `json:"error,omitempty"`
}

// TestErrorEnvelopeAPI404 verifies that a 404 from the API surfaces the
// shared error envelope: source = DAYTONA_API, status code matches, and the
// legacy backward-compat fields are still present.
func TestErrorEnvelopeAPI404(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	// 00000000-... is a syntactically valid UUID that never matches a real
	// sandbox, guaranteeing a 404 path through AllExceptionsFilter.
	resp, body := client.DoRequest(t, http.MethodGet, "/sandbox/00000000-0000-0000-0000-000000000000", nil)
	require.Equal(t, http.StatusNotFound, resp.StatusCode, "expected 404, got %d: %s", resp.StatusCode, string(body))

	var env errorEnvelope
	require.NoError(t, json.Unmarshal(body, &env), "response body is not JSON: %s", string(body))

	assert.Equal(t, http.StatusNotFound, env.StatusCode, "statusCode field must mirror HTTP status")
	assert.Equal(t, "DAYTONA_API", env.Source, "source field must be stamped by AllExceptionsFilter")
	assert.NotEmpty(t, env.Message, "message field is required")
	assert.NotEmpty(t, env.Timestamp, "timestamp field is required")

	// Legacy fields preserved for backward compatibility with older SDK consumers.
	assert.NotEmpty(t, env.Error, "legacy 'error' field must remain for backward compat")
	assert.NotEmpty(t, env.Path, "legacy 'path' field must remain for backward compat")

	t.Logf("404 envelope: source=%s code=%s message=%s", env.Source, env.Code, env.Message)
}

// TestErrorEnvelopeAPIUnauthorized verifies the unauthenticated path also
// flows through AllExceptionsFilter and stamps `source`. No `code` is set
// because the error is fully classified by the HTTP status (401).
func TestErrorEnvelopeAPIUnauthorized(t *testing.T) {
	cfg := LoadConfig(t)
	// Force an auth failure by using a bogus key.
	cfg.APIKey = "dtn_invalid"
	client := NewAPIClient(cfg)

	resp, body := client.DoRequest(t, http.MethodGet, "/sandbox/paginated", nil)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401, got %d: %s", resp.StatusCode, string(body))

	var env errorEnvelope
	require.NoError(t, json.Unmarshal(body, &env), "response body is not JSON: %s", string(body))

	assert.Equal(t, http.StatusUnauthorized, env.StatusCode)
	assert.Equal(t, "DAYTONA_API", env.Source, "source must be stamped on auth failures too")
	assert.Empty(t, env.Code, "401 without a specific code should omit the field, not emit a placeholder")

	t.Logf("401 envelope: source=%s code=%q message=%s", env.Source, env.Code, env.Message)
}

// TestErrorEnvelopeAPIBadRequest verifies that a 400 from a validation error
// flows through the new filter shape too. We hit a snapshot endpoint with a
// nonsensical body so class-validator rejects it before any runner is involved.
func TestErrorEnvelopeAPIBadRequest(t *testing.T) {
	cfg := LoadConfig(t)
	client := NewAPIClient(cfg)

	// snapshot:imageName is required and must be a valid image — sending empty
	// triggers class-validator failure before any downstream work.
	resp, body := client.DoRequest(t, http.MethodPost, "/snapshots", map[string]interface{}{})
	require.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400, got %d: %s", resp.StatusCode, string(body))

	var env errorEnvelope
	require.NoError(t, json.Unmarshal(body, &env), "response body is not JSON: %s", string(body))

	assert.Equal(t, http.StatusBadRequest, env.StatusCode)
	assert.Equal(t, "DAYTONA_API", env.Source)
	assert.NotEmpty(t, env.Message)

	t.Logf("400 envelope: source=%s code=%q message=%s", env.Source, env.Code, env.Message)
}
