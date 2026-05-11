// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build e2e

package e2e_test

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionDisplayPandasHtml verifies a pandas DataFrame as the last expression yields a
// display chunk containing text/html with a <table> in the data.
func TestSessionDisplayPandasHtml(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-rich-outputs")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	wsURL, _ := body["wsUrl"].(string)
	contextID, _ := body["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, contextID) })

	ws, closer := dialSessionWebSocket(t, wsURL)
	defer closer()

	sendExec(t, ws, "import pandas as pd\npd.DataFrame({'a':[1,2]})\n", nil)
	frames := collectFrames(t, ws, 30*time.Second)

	display := findFrame(frames, "display")
	require.NotNil(t, display, "must emit a display chunk for the trailing pd.DataFrame")

	formats, _ := display["formats"].([]interface{})
	hasHTML := false
	for _, f := range formats {
		if s, _ := f.(string); s == "text/html" {
			hasHTML = true
			break
		}
	}
	assert.True(t, hasHTML, "display formats must include text/html")

	data, _ := display["data"].(map[string]interface{})
	html, _ := data["text/html"].(string)
	assert.Contains(t, html, "<table", "html data must contain a <table tag")
}

// TestSessionDisplayMatplotlibPng verifies plt.show() emits a display chunk with a base64 PNG
// whose first bytes are the PNG magic header.
func TestSessionDisplayMatplotlibPng(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-rich-outputs")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	body, status := ic.Connect(t, map[string]interface{}{
		"template": "python-default",
		"language": "python",
	})
	require.Equal(t, http.StatusOK, status)
	wsURL, _ := body["wsUrl"].(string)
	contextID, _ := body["sessionId"].(string)
	t.Cleanup(func() { _ = ic.DeleteSession(t, contextID) })

	ws, closer := dialSessionWebSocket(t, wsURL)
	defer closer()

	sendExec(t, ws, "import matplotlib.pyplot as plt\nplt.plot([1,2,3])\nplt.show()\n", nil)
	frames := collectFrames(t, ws, 60*time.Second)

	display := findFrame(frames, "display")
	require.NotNil(t, display, "matplotlib show must emit a display chunk")

	data, _ := display["data"].(map[string]interface{})
	pngB64, _ := data["image/png"].(string)
	require.NotEmpty(t, pngB64, "image/png data must be present and non-empty")

	bytes, err := base64.StdEncoding.DecodeString(pngB64)
	require.NoError(t, err, "image/png must be valid base64")
	require.GreaterOrEqual(t, len(bytes), 8, "PNG must have header bytes")
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngMagic {
		assert.Equal(t, b, bytes[i], "PNG magic byte mismatch at offset %d", i)
	}
}

// TestSessionDisplayJSON verifies both Python dict and TS object as last-expression yield
// application/json display.
func TestSessionDisplayJSON(t *testing.T) {
	t.Skipf("not yet implemented: session-daemon-rich-outputs")

	cfg := LoadConfig(t)
	api := NewAPIClient(cfg)
	ic := NewSessionClient(api)

	t.Run("Python", func(t *testing.T) {
		body, status := ic.Connect(t, map[string]interface{}{
			"template": "python-default", "language": "python",
		})
		require.Equal(t, http.StatusOK, status)
		wsURL, _ := body["wsUrl"].(string)
		ctxID, _ := body["sessionId"].(string)
		t.Cleanup(func() { _ = ic.DeleteSession(t, ctxID) })

		ws, closer := dialSessionWebSocket(t, wsURL)
		defer closer()
		sendExec(t, ws, `{"a":1}`, nil)
		frames := collectFrames(t, ws, 15*time.Second)
		display := findFrame(frames, "display")
		require.NotNil(t, display)
		data, _ := display["data"].(map[string]interface{})
		jsonStr, _ := data["application/json"].(string)
		assert.True(t, strings.Contains(jsonStr, `"a"`) && strings.Contains(jsonStr, "1"),
			"json data should contain {a:1}")
	})

	t.Run("TypeScript", func(t *testing.T) {
		body, status := ic.Connect(t, map[string]interface{}{
			"template": "python-default", "language": "typescript",
		})
		require.Equal(t, http.StatusOK, status)
		wsURL, _ := body["wsUrl"].(string)
		ctxID, _ := body["sessionId"].(string)
		t.Cleanup(func() { _ = ic.DeleteSession(t, ctxID) })

		ws, closer := dialSessionWebSocket(t, wsURL)
		defer closer()
		sendExec(t, ws, `({ a: 1 })`, nil)
		frames := collectFrames(t, ws, 15*time.Second)
		display := findFrame(frames, "display")
		require.NotNil(t, display)
		data, _ := display["data"].(map[string]interface{})
		jsonStr, _ := data["application/json"].(string)
		assert.True(t, strings.Contains(jsonStr, `"a"`) && strings.Contains(jsonStr, "1"))
	})
}
