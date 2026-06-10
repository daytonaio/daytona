// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/daytonaio/daytona/cli/cmd/common"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/internal/clierr"
)

// followLogsRecorder tracks the build-logs requests served by the test server.
type followLogsRecorder struct {
	mu      sync.Mutex
	hits    int
	follows []bool
}

func (r *followLogsRecorder) record(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hits++
	r.follows = append(r.follows, req.URL.Query().Get("follow") == "true")
}

func (r *followLogsRecorder) snapshot() (int, []bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hits, append([]bool(nil), r.follows...)
}

func followLogsTestServer(t *testing.T, body string, rec *followLogsRecorder) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sandbox/sbx-1/build-logs" {
			http.NotFound(w, r)
			return
		}
		rec.record(r)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(server.Close)
	return server
}

func followLogsTestParams(serverURL string) common.ReadLogParams {
	key := "test-api-key"
	follow := true
	return common.ReadLogParams{
		Id:           "sbx-1",
		ServerUrl:    serverURL,
		ServerApi:    config.ServerApi{Url: serverURL, Key: &key},
		Follow:       &follow,
		ResourceType: common.ResourceTypeSandbox,
	}
}

// followLogsCaptureStdout redirects os.Stdout for the duration of fn and
// returns everything written to it.
func followLogsCaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	collected := make(chan string, 1)
	go func() {
		data, _ := io.ReadAll(r)
		collected <- string(data)
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing stdout pipe: %v", err)
	}
	os.Stdout = orig
	return <-collected
}

// TestFollowBuildLogsAlreadyTerminal is the regression test for follow mode
// truncating output when the resource is already terminal at entry: the
// complete log history must arrive via exactly one plain (non-follow) fetch,
// and the terminal-state error semantics must still apply.
func TestFollowBuildLogsAlreadyTerminal(t *testing.T) {
	body := strings.Repeat("a long line of build output that must not be truncated\n", 200)

	tests := []struct {
		name        string
		failErr     error
		wantInError string
	}{
		{name: "terminal success", failErr: nil},
		{name: "terminal failure", failErr: clierr.New(clierr.CategoryServer, "sandbox processing failed: boom"), wantInError: "boom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &followLogsRecorder{}
			server := followLogsTestServer(t, body, rec)

			polls := 0
			var err error
			out := followLogsCaptureStdout(t, func() {
				err = common.FollowBuildLogs(context.Background(), followLogsTestParams(server.URL), func(context.Context) (bool, error) {
					polls++
					return true, tt.failErr
				})
			})

			if tt.failErr == nil {
				if err != nil {
					t.Fatalf("FollowBuildLogs() error: %v", err)
				}
			} else {
				if !clierr.HasCategory(err, clierr.CategoryServer) {
					t.Fatalf("FollowBuildLogs() error = %v, want server-category clierr", err)
				}
				if !strings.Contains(err.Error(), tt.wantInError) {
					t.Errorf("FollowBuildLogs() error = %q, want it to mention %q", err.Error(), tt.wantInError)
				}
			}
			if out != body {
				t.Errorf("streamed %d bytes, want the complete %d-byte log output", len(out), len(body))
			}
			hits, follows := rec.snapshot()
			if hits != 1 {
				t.Errorf("build-logs requests = %d, want exactly 1", hits)
			}
			if len(follows) > 0 && follows[0] {
				t.Error("already-terminal fetch used follow=true, want a plain fetch")
			}
			if polls != 1 {
				t.Errorf("pollState calls = %d, want 1", polls)
			}
		})
	}
}

func TestFollowBuildLogsPollErrorAborts(t *testing.T) {
	rec := &followLogsRecorder{}
	server := followLogsTestServer(t, "unused", rec)

	pollErr := errors.New("poll failed")
	var err error
	out := followLogsCaptureStdout(t, func() {
		err = common.FollowBuildLogs(context.Background(), followLogsTestParams(server.URL), func(context.Context) (bool, error) {
			return false, pollErr
		})
	})

	if !errors.Is(err, pollErr) {
		t.Fatalf("FollowBuildLogs() error = %v, want %v", err, pollErr)
	}
	if out != "" {
		t.Errorf("streamed output %q, want none", out)
	}
	if hits, _ := rec.snapshot(); hits != 0 {
		t.Errorf("build-logs requests = %d, want 0", hits)
	}
}

func TestFollowBuildLogsStreamsUntilTerminal(t *testing.T) {
	body := "streaming build output\n"
	rec := &followLogsRecorder{}
	server := followLogsTestServer(t, body, rec)

	polls := 0
	var err error
	out := followLogsCaptureStdout(t, func() {
		err = common.FollowBuildLogs(context.Background(), followLogsTestParams(server.URL), func(context.Context) (bool, error) {
			polls++
			return polls > 1, nil
		})
	})

	if err != nil {
		t.Fatalf("FollowBuildLogs() error: %v", err)
	}
	if out != body {
		t.Errorf("streamed output %q, want %q", out, body)
	}
	hits, follows := rec.snapshot()
	if hits != 1 {
		t.Fatalf("build-logs requests = %d, want 1", hits)
	}
	if !follows[0] {
		t.Error("streaming request missing follow=true query")
	}
	if polls < 2 {
		t.Errorf("pollState calls = %d, want at least 2", polls)
	}
}
