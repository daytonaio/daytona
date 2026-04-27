// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/daytonaio/daytona/libs/sdk-go/pkg/types"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeInterpreterServiceCreation(t *testing.T) {
	ci := NewCodeInterpreterService(nil, nil)
	require.NotNil(t, ci)
}

func TestCodeInterpreterRunCode(t *testing.T) {
	wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		var req map[string]interface{}
		if err := conn.ReadJSON(&req); err != nil {
			return
		}

		assert.Equal(t, "print('hello')", req["code"])

		_ = conn.WriteJSON(types.OutputMessage{Type: "stdout", Text: "hello\n"})

		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}))
	defer wsServer.Close()

	wsURL := strings.Replace(wsServer.URL, "http://", "http://", 1)

	client := createTestToolboxClient(httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	cfg := client.GetConfig()
	cfg.Servers[0].URL = wsURL

	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	channels, err := ci.RunCode(ctx, "print('hello')")
	require.NoError(t, err)
	require.NotNil(t, channels)

	result := <-channels.Done
	require.NotNil(t, result)
	assert.Contains(t, result.Stdout, "hello")
}

func TestCodeInterpreterRunCodeConnectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	channels, err := ci.RunCode(ctx, "print('hello')")
	require.NoError(t, err)

	result := <-channels.Done
	require.NotNil(t, result)
	assert.NotNil(t, result.Error)
	assert.Equal(t, "ConnectionError", result.Error.Name)
}

func TestCodeInterpreterCreateContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":        "ctx-1",
			"cwd":       "/home/user",
			"language":  "python",
			"active":    true,
			"createdAt": "2025-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	result, err := ci.CreateContext(ctx, nil)
	require.NoError(t, err)
	assert.Equal(t, "ctx-1", result["id"])
	assert.Equal(t, "python", result["language"])
	assert.Equal(t, true, result["active"])
}

func TestCodeInterpreterCreateContextWithCwd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"id":        "ctx-2",
			"cwd":       "/app",
			"language":  "python",
			"active":    true,
			"createdAt": "2025-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	cwd := "/app"
	ctx := context.Background()
	result, err := ci.CreateContext(ctx, &cwd)
	require.NoError(t, err)
	assert.Equal(t, "/app", result["cwd"])
}

func TestCodeInterpreterListContexts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]interface{}{
			"contexts": []map[string]interface{}{
				{"id": "ctx-1", "cwd": "/home", "language": "python", "active": true, "createdAt": "2025-01-01T00:00:00Z"},
				{"id": "ctx-2", "cwd": "/app", "language": "python", "active": false, "createdAt": "2025-01-02T00:00:00Z"},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	contexts, err := ci.ListContexts(ctx)
	require.NoError(t, err)
	require.Len(t, contexts, 2)
	assert.Equal(t, "ctx-1", contexts[0]["id"])
	assert.Equal(t, "ctx-2", contexts[1]["id"])
}

func TestCodeInterpreterDeleteContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	err := ci.DeleteContext(ctx, "ctx-1")
	assert.NoError(t, err)
}

func TestCodeInterpreterBuildWebSocketURL(t *testing.T) {
	ci := &CodeInterpreterService{}

	tests := []struct {
		name     string
		baseURL  string
		path     string
		expected string
		hasError bool
	}{
		{
			name:     "http to ws",
			baseURL:  "http://localhost:8080",
			path:     "/process/interpreter/execute",
			expected: "ws://localhost:8080/process/interpreter/execute",
		},
		{
			name:     "https to wss",
			baseURL:  "https://api.example.com",
			path:     "/process/interpreter/execute",
			expected: "wss://api.example.com/process/interpreter/execute",
		},
		{
			name:     "unsupported scheme",
			baseURL:  "ftp://example.com",
			path:     "/test",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ci.buildWebSocketURL(tt.baseURL, tt.path)
			if tt.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestCodeInterpreterBuildHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()

	client := createTestToolboxClient(server)
	cfg := client.GetConfig()
	cfg.AddDefaultHeader("Authorization", "Bearer test-token")
	cfg.AddDefaultHeader("X-Daytona-Source", "sdk-go")

	ci := NewCodeInterpreterService(client, nil)
	headers := ci.buildHeaders(client)

	assert.Contains(t, headers["Authorization"], "Bearer test-token")
	assert.Contains(t, headers["X-Daytona-Source"], "sdk-go")
}

func TestCodeInterpreterErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"message": "server error"})
	}))
	defer server.Close()

	client := createTestToolboxClient(server)
	ci := NewCodeInterpreterService(client, nil)

	ctx := context.Background()
	_, err := ci.CreateContext(ctx, nil)
	require.Error(t, err)
}

func TestCodeInterpreterRunCodeStreamBehaviors(t *testing.T) {
	t.Run("captures stderr and execution error", func(t *testing.T) {
		wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			var req map[string]any
			if conn.ReadJSON(&req) != nil {
				return
			}
			_ = conn.WriteJSON(types.OutputMessage{Type: "stdout", Text: "partial out\n"})
			_ = conn.WriteJSON(types.OutputMessage{Type: "stderr", Text: "partial err\n"})
			_ = conn.WriteJSON(types.OutputMessage{Type: "error", Name: "ValueError", Value: "boom", Traceback: "trace"})
			_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}))
		defer wsServer.Close()

		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer httpServer.Close()
		client := createTestToolboxClient(httpServer)
		client.GetConfig().Servers[0].URL = wsServer.URL
		ci := NewCodeInterpreterService(client, nil)
		channels, err := ci.RunCode(context.Background(), "raise ValueError('boom')")
		require.NoError(t, err)
		result := <-channels.Done
		require.NotNil(t, result.Error)
		assert.Equal(t, "ValueError", result.Error.Name)
		assert.Equal(t, "partial out\n", result.Stdout)
		assert.Equal(t, "partial err\n", result.Stderr)
		require.NotNil(t, result.Error.Traceback)
		assert.Equal(t, "trace", *result.Error.Traceback)
	})

	t.Run("maps websocket timeout close to timeout error", func(t *testing.T) {
		wsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				return
			}
			defer conn.Close()
			var req map[string]any
			if conn.ReadJSON(&req) != nil {
				return
			}
			_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(4001, "timeout"), time.Now().Add(time.Second))
		}))
		defer wsServer.Close()

		httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		defer httpServer.Close()
		client := createTestToolboxClient(httpServer)
		client.GetConfig().Servers[0].URL = wsServer.URL
		ci := NewCodeInterpreterService(client, nil)
		channels, err := ci.RunCode(context.Background(), "while True: pass")
		require.NoError(t, err)
		result := <-channels.Done
		require.NotNil(t, result.Error)
		assert.Equal(t, "TimeoutError", result.Error.Name)
	})
}

func TestCodeInterpreterContextOperationsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"id": "ctx-10", "cwd": "/tmp", "language": "python", "active": true, "createdAt": "2025-01-01T00:00:00Z"})
		case http.MethodGet:
			writeJSONResponse(t, w, http.StatusOK, map[string]any{"contexts": []map[string]any{{"id": "ctx-10", "cwd": "/tmp", "language": "python", "active": true, "createdAt": "2025-01-01T00:00:00Z"}}})
		case http.MethodDelete:
			writeJSONResponse(t, w, http.StatusOK, map[string]any{})
		}
	}))
	defer server.Close()

	ci := NewCodeInterpreterService(createTestToolboxClient(server), nil)
	ctxInfo, err := ci.CreateContext(context.Background(), strPtr("/tmp"))
	require.NoError(t, err)
	assert.Equal(t, "/tmp", ctxInfo["cwd"])
	contexts, err := ci.ListContexts(context.Background())
	require.NoError(t, err)
	assert.Len(t, contexts, 1)
	require.NoError(t, ci.DeleteContext(context.Background(), "ctx-10"))
}
