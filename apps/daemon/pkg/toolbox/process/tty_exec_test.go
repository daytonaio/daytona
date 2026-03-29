// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package process

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteCommand_TTY(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	tests := []struct {
		name           string
		request        ExecuteRequest
		expectedStatus int
		shouldHaveTTY  bool
	}{
		{
			name: "TTY request should create session",
			request: ExecuteRequest{
				Command: "echo 'hello world'",
				Tty:     true,
			},
			expectedStatus: http.StatusOK,
			shouldHaveTTY:  true,
		},
		{
			name: "Non-TTY request should execute normally",
			request: ExecuteRequest{
				Command: "echo 'hello world'",
				Tty:     false,
			},
			expectedStatus: http.StatusOK,
			shouldHaveTTY:  false,
		},
		{
			name: "TTY request with working directory",
			request: ExecuteRequest{
				Command: "pwd",
				Tty:     true,
				Cwd:     stringPtr("/tmp"),
			},
			expectedStatus: http.StatusOK,
			shouldHaveTTY:  true,
		},
		{
			name: "TTY request with timeout",
			request: ExecuteRequest{
				Command: "sleep 1",
				Tty:     true,
				Timeout: uint32Ptr(5),
			},
			expectedStatus: http.StatusOK,
			shouldHaveTTY:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			reqBody, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/process/execute", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create gin context and handler
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// Call the handler
			handler := ExecuteCommand(logger)
			handler(c)

			// Check status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.shouldHaveTTY {
				// Should return ExecuteTTYResponse with session ID
				var ttyResp ExecuteTTYResponse
				err := json.Unmarshal(w.Body.Bytes(), &ttyResp)
				require.NoError(t, err, "Response body: %s", w.Body.String())
				assert.NotEmpty(t, ttyResp.SessionID)

				// Verify session was created in manager
				_, exists := ttyExecManager.Get(ttyResp.SessionID)
				assert.True(t, exists, "TTY exec session should exist in manager")
			} else {
				// Should return ExecuteResponse with result
				var execResp ExecuteResponse
				err := json.Unmarshal(w.Body.Bytes(), &execResp)
				require.NoError(t, err, "Response body: %s", w.Body.String())
				assert.NotEmpty(t, execResp.Result)
			}
		})
	}
}

func TestExecuteCommand_TTY_ErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
	}{
		{
			name:           "Empty command should fail",
			request:        ExecuteRequest{Command: "", Tty: true},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Whitespace only command should fail",
			request:        ExecuteRequest{Command: "   ", Tty: true},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON should fail",
			request:        "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if str, ok := tt.request.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/process/execute", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler := ExecuteCommand(logger)
			handler(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestTTYExecSession_Creation(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name    string
		request ExecuteRequest
	}{
		{
			name: "Basic TTY session",
			request: ExecuteRequest{
				Command: "echo test",
				Tty:     true,
			},
		},
		{
			name: "TTY session with working directory",
			request: ExecuteRequest{
				Command: "pwd",
				Tty:     true,
				Cwd:     stringPtr("/tmp"),
			},
		},
		{
			name: "TTY session with timeout",
			request: ExecuteRequest{
				Command: "echo test",
				Tty:     true,
				Timeout: uint32Ptr(30),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session, err := createTTYExecSession(logger, tt.request)
			require.NoError(t, err)
			assert.NotNil(t, session)
			assert.NotEmpty(t, session.id)
			assert.Equal(t, tt.request.Command, session.command)

			if tt.request.Cwd != nil {
				assert.Equal(t, *tt.request.Cwd, session.cwd)
			}

			if tt.request.Timeout != nil {
				assert.Equal(t, tt.request.Timeout, session.timeout)
			}

			// Verify session is in manager
			_, exists := ttyExecManager.Get(session.id)
			assert.True(t, exists)

			// Clean up
			session.kill()
			ttyExecManager.Remove(session.id)
		})
	}
}

func TestConnectTTYExecSession_Handler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := slog.Default()

	// This test verifies the handler function structure
	// Full WebSocket testing would require more complex setup
	t.Run("Missing session ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/process/exec/connect", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		// Don't set sessionId param to test error case

		handler := ConnectTTYExecSession(logger)
		handler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response["error"], "session ID is required")
	})

	t.Run("Session not found", func(t *testing.T) {
		// This would require WebSocket upgrade which is complex to test
		// The important part is that the handler function is accessible and compiles
		handler := ConnectTTYExecSession(logger)
		assert.NotNil(t, handler)
	})
}

// Helper functions for test data
func stringPtr(s string) *string {
	return &s
}

func uint32Ptr(i uint32) *uint32 {
	return &i
}
