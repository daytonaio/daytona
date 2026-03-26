// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package daytona

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"bytes"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TestProcessWebsocketStream tests the WebSocket stream demultiplexing
func TestProcessWebsocketStream(t *testing.T) {
	tests := []struct {
		name           string
		serverMessages [][]byte
		expectedStdout string
		expectedStderr string
	}{
		{
			name: "simple stdout only",
			serverMessages: [][]byte{
				append(stdoutPrefixBytes, []byte("hello world\n")...),
			},
			expectedStdout: "hello world\n",
			expectedStderr: "",
		},
		{
			name: "simple stderr only",
			serverMessages: [][]byte{
				append(stderrPrefixBytes, []byte("error message\n")...),
			},
			expectedStdout: "",
			expectedStderr: "error message\n",
		},
		{
			name: "interleaved stdout and stderr",
			serverMessages: [][]byte{
				append(stdoutPrefixBytes, []byte("stdout line 1\n")...),
				append(stderrPrefixBytes, []byte("stderr line 1\n")...),
				append(stdoutPrefixBytes, []byte("stdout line 2\n")...),
			},
			expectedStdout: "stdout line 1\nstdout line 2\n",
			expectedStderr: "stderr line 1\n",
		},
		{
			name: "multiple lines in single message",
			serverMessages: [][]byte{
				bytes.Join([][]byte{
					stdoutPrefixBytes, []byte("line 1\n"),
					stderrPrefixBytes, []byte("error 1\n"),
					stdoutPrefixBytes, []byte("line 2\n"),
				}, nil),
			},
			expectedStdout: "line 1\nline 2\n",
			expectedStderr: "error 1\n",
		},
		{
			name: "marker split across messages",
			serverMessages: [][]byte{
				append(stdoutPrefixBytes, []byte("start ")...),
				[]byte("middle "),
				[]byte("end\n"),
			},
			expectedStdout: "start middle end\n",
			expectedStderr: "",
		},
		{
			name: "empty messages ignored",
			serverMessages: [][]byte{
				append(stdoutPrefixBytes, []byte("before\n")...),
				{},
				append(stdoutPrefixBytes, []byte("after\n")...),
			},
			expectedStdout: "before\nafter\n",
			expectedStderr: "",
		},
		{
			name: "rapid switching between streams",
			serverMessages: [][]byte{
				append(stdoutPrefixBytes, []byte("s1")...),
				append(stderrPrefixBytes, []byte("e1")...),
				append(stdoutPrefixBytes, []byte("s2")...),
				append(stderrPrefixBytes, []byte("e2")...),
				append(stdoutPrefixBytes, []byte("s3\n")...),
			},
			expectedStdout: "s1s2s3\n",
			expectedStderr: "e1e2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create WebSocket test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					t.Fatalf("Failed to upgrade connection: %v", err)
					return
				}
				defer conn.Close()

				// Send all test messages
				for _, msg := range tt.serverMessages {
					if len(msg) > 0 {
						err := conn.WriteMessage(websocket.BinaryMessage, msg)
						if err != nil {
							return
						}
					}
				}

				// Close the connection gracefully
				err = conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					t.Fatalf("Failed to write close message: %v", err)
					return
				}
				err = conn.Close()
				if err != nil {
					t.Fatalf("Failed to close connection: %v", err)
					return
				}
			}))
			defer server.Close()

			// Convert HTTP URL to WebSocket URL
			wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

			// Connect to the WebSocket server
			conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			require.NoError(t, err)
			defer conn.Close()

			// Create channels and run the stream processor
			stdout := make(chan string, 100)
			stderr := make(chan string, 100)

			ctx := context.Background()
			err = processWebsocketStream(ctx, conn, stdout, stderr)
			assert.NoError(t, err)

			// Close channels after processing (normally done by GetSessionCommandLogsStream)
			close(stdout)
			close(stderr)

			// Collect and concatenate results (chunks may vary due to buffering)
			var stdoutBuilder, stderrBuilder strings.Builder
			for chunk := range stdout {
				stdoutBuilder.WriteString(chunk)
			}
			for chunk := range stderr {
				stderrBuilder.WriteString(chunk)
			}

			// Verify combined results
			assert.Equal(t, tt.expectedStdout, stdoutBuilder.String(), "stdout mismatch")
			assert.Equal(t, tt.expectedStderr, stderrBuilder.String(), "stderr mismatch")
		})
	}
}

// TestProcessWebsocketStreamContextCancellation tests context cancellation
func TestProcessWebsocketStreamContextCancellation(t *testing.T) {
	// Create a server that streams indefinitely
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Stream messages slowly
		for i := 0; i < 100; i++ {
			msg := append(stdoutPrefixBytes, []byte("message\n")...)
			if err := conn.WriteMessage(websocket.BinaryMessage, msg); err != nil {
				return
			}
			time.Sleep(50 * time.Millisecond)
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	// Create a context that cancels after 100ms
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = processWebsocketStream(ctx, conn, stdout, stderr)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestProcessWebsocketStreamPartialMarker tests handling of markers split across messages
func TestProcessWebsocketStreamPartialMarker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send stdout marker + data
		err = conn.WriteMessage(websocket.BinaryMessage, append(stdoutPrefixBytes, []byte("hello ")...))
		if err != nil {
			return
		}

		// Send more data without marker (continues stdout)
		err = conn.WriteMessage(websocket.BinaryMessage, []byte("world"))
		if err != nil {
			return
		}

		// Switch to stderr
		err = conn.WriteMessage(websocket.BinaryMessage, append(stderrPrefixBytes, []byte("error\n")...))
		if err != nil {
			return
		}

		err = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return
		}
		err = conn.Close()
		if err != nil {
			return
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	ctx := context.Background()
	err = processWebsocketStream(ctx, conn, stdout, stderr)
	assert.NoError(t, err)

	close(stdout)
	close(stderr)

	// Collect results
	var gotStdout, gotStderr []string
	for chunk := range stdout {
		gotStdout = append(gotStdout, chunk)
	}
	for chunk := range stderr {
		gotStderr = append(gotStderr, chunk)
	}

	// Verify stdout received "hello world"
	assert.Equal(t, "hello world", strings.Join(gotStdout, ""))
	assert.Equal(t, "error\n", strings.Join(gotStderr, ""))
}

// TestProcessWebsocketStreamNoMarkerAtStart tests data without initial marker
func TestProcessWebsocketStreamNoMarkerAtStart(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send data without marker first (should be dropped)
		err = conn.WriteMessage(websocket.BinaryMessage, []byte("dropped data"))
		if err != nil {
			return
		}

		// Then send proper stdout
		err = conn.WriteMessage(websocket.BinaryMessage, append(stdoutPrefixBytes, []byte("kept data\n")...))
		if err != nil {
			return
		}

		err = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return
		}
		err = conn.Close()
		if err != nil {
			return
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	stdout := make(chan string, 100)
	stderr := make(chan string, 100)

	ctx := context.Background()
	err = processWebsocketStream(ctx, conn, stdout, stderr)
	assert.NoError(t, err)

	close(stdout)
	close(stderr)

	// Collect results - only "kept data" should be received
	var stdoutBuilder strings.Builder
	for chunk := range stdout {
		stdoutBuilder.WriteString(chunk)
	}

	assert.Equal(t, "kept data\n", stdoutBuilder.String())
}
