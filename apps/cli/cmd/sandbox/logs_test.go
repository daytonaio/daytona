// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestIsSandboxBuildDone(t *testing.T) {
	tests := []struct {
		state      apiclient.SandboxState
		wantDone   bool
		wantFailed bool
	}{
		{apiclient.SANDBOXSTATE_STARTED, true, false},
		{apiclient.SANDBOXSTATE_STOPPED, true, false},
		{apiclient.SANDBOXSTATE_STOPPING, true, false},
		{apiclient.SANDBOXSTATE_ARCHIVED, true, false},
		{apiclient.SANDBOXSTATE_ARCHIVING, true, false},
		{apiclient.SANDBOXSTATE_DESTROYED, true, false},
		{apiclient.SANDBOXSTATE_DESTROYING, true, false},
		{apiclient.SANDBOXSTATE_ERROR, true, true},
		{apiclient.SANDBOXSTATE_BUILD_FAILED, true, true},
		{apiclient.SANDBOXSTATE_CREATING, false, false},
		{apiclient.SANDBOXSTATE_PENDING_BUILD, false, false},
		{apiclient.SANDBOXSTATE_BUILDING_SNAPSHOT, false, false},
		{apiclient.SANDBOXSTATE_PULLING_SNAPSHOT, false, false},
		{apiclient.SANDBOXSTATE_STARTING, false, false},
		{apiclient.SANDBOXSTATE_RESTORING, false, false},
		{apiclient.SANDBOXSTATE_UNKNOWN, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			done, failed := isSandboxBuildDone(tt.state)
			if done != tt.wantDone || failed != tt.wantFailed {
				t.Errorf("isSandboxBuildDone(%q) = (%v, %v), want (%v, %v)", tt.state, done, failed, tt.wantDone, tt.wantFailed)
			}
		})
	}
}

// logsRequestRecorder tracks the build-logs requests served by the test server.
type logsRequestRecorder struct {
	mu      sync.Mutex
	hits    int
	follows []bool
}

func (r *logsRequestRecorder) record(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hits++
	r.follows = append(r.follows, req.URL.Query().Get("follow") == "true")
}

func (r *logsRequestRecorder) snapshot() (int, []bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.hits, append([]bool(nil), r.follows...)
}

// logsTestSandboxJSON renders a minimal Sandbox payload containing every
// property the generated client requires plus the given state and error reason.
func logsTestSandboxJSON(state, errorReason string) string {
	reason := "null"
	if errorReason != "" {
		reason = fmt.Sprintf("%q", errorReason)
	}
	return fmt.Sprintf(`{
		"id": "sbx-1",
		"organizationId": "org-1",
		"name": "my-sandbox",
		"user": "daytona",
		"env": {},
		"labels": {},
		"public": false,
		"networkBlockAll": false,
		"target": "us",
		"cpu": 1,
		"gpu": 0,
		"memory": 1,
		"disk": 1,
		"toolboxProxyUrl": "",
		"state": %q,
		"errorReason": %s
	}`, state, reason)
}

// logsTestServer emulates the API routes the logs command hits: GetSandbox
// and the raw build-logs stream.
func logsTestServer(t *testing.T, state, errorReason, logBody string, rec *logsRequestRecorder) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/sandbox/sbx-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, logsTestSandboxJSON(state, errorReason))
	})
	mux.HandleFunc("/sandbox/sbx-1/build-logs", func(w http.ResponseWriter, r *http.Request) {
		rec.record(r)
		_, _ = io.WriteString(w, logBody)
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)
	return server
}

// logsTestEnv points profile resolution at the test server and sandboxes the
// config cache away from the real machine config.
func logsTestEnv(t *testing.T, serverURL string) {
	t.Helper()
	t.Setenv("DAYTONA_CONFIG_DIR", t.TempDir())
	t.Setenv("DAYTONA_API_KEY", "test-api-key")
	t.Setenv("DAYTONA_API_URL", serverURL)
}

func logsSetFollowFlag(t *testing.T, v bool) {
	t.Helper()
	orig := logsFollowFlag
	logsFollowFlag = v
	t.Cleanup(func() { logsFollowFlag = orig })
}

// logsCaptureStdout redirects os.Stdout for the duration of fn and returns
// everything written to it.
func logsCaptureStdout(t *testing.T, fn func()) string {
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

func TestLogsCmdNonFollowStreamsBody(t *testing.T) {
	body := "build log line 1\nbuild log line 2\n"
	rec := &logsRequestRecorder{}
	server := logsTestServer(t, "started", "", body, rec)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, false)

	var err error
	out := logsCaptureStdout(t, func() {
		err = LogsCmd.RunE(LogsCmd, []string{"sbx-1"})
	})

	if err != nil {
		t.Fatalf("LogsCmd.RunE() error: %v", err)
	}
	if out != body {
		t.Errorf("stdout = %q, want %q", out, body)
	}
	hits, follows := rec.snapshot()
	if hits != 1 {
		t.Fatalf("build-logs requests = %d, want 1", hits)
	}
	if follows[0] {
		t.Error("non-follow fetch sent follow=true")
	}
}

func TestLogsCmdNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"message":"sandbox not found","statusCode":404}`)
	}))
	t.Cleanup(server.Close)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, false)

	err := LogsCmd.RunE(LogsCmd, []string{"missing"})
	if err == nil {
		t.Fatal("LogsCmd.RunE() expected error, got nil")
	}
	if !clierr.HasCategory(err, clierr.CategoryNotFound) {
		t.Errorf("LogsCmd.RunE() error = %v, want not_found-category clierr", err)
	}
}

// TestLogsCmdFollowAlreadyTerminalCompleteFetch is the regression test for
// follow mode truncating output when the sandbox is already terminal at
// entry: the complete log history must arrive via exactly one plain
// (non-follow) fetch.
func TestLogsCmdFollowAlreadyTerminalCompleteFetch(t *testing.T) {
	body := strings.Repeat("a long line of build output that must not be truncated\n", 200)
	rec := &logsRequestRecorder{}
	server := logsTestServer(t, "started", "", body, rec)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, true)

	var err error
	out := logsCaptureStdout(t, func() {
		err = LogsCmd.RunE(LogsCmd, []string{"sbx-1"})
	})

	if err != nil {
		t.Fatalf("LogsCmd.RunE() error: %v", err)
	}
	if out != body {
		t.Errorf("streamed %d bytes, want the complete %d-byte log output", len(out), len(body))
	}
	hits, follows := rec.snapshot()
	if hits != 1 {
		t.Fatalf("build-logs requests = %d, want exactly 1", hits)
	}
	if follows[0] {
		t.Error("already-terminal fetch used follow=true, want a plain fetch")
	}
}

func TestLogsCmdFollowBuildFailedPropagatesReason(t *testing.T) {
	body := "build log output\n"
	rec := &logsRequestRecorder{}
	server := logsTestServer(t, "build_failed", "docker build exploded", body, rec)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, true)

	var err error
	out := logsCaptureStdout(t, func() {
		err = LogsCmd.RunE(LogsCmd, []string{"sbx-1"})
	})

	if !clierr.HasCategory(err, clierr.CategoryServer) {
		t.Fatalf("LogsCmd.RunE() error = %v, want server-category clierr", err)
	}
	if !strings.Contains(err.Error(), "docker build exploded") {
		t.Errorf("LogsCmd.RunE() error = %q, want it to mention the error reason", err.Error())
	}
	if out != body {
		t.Errorf("stdout = %q, want %q", out, body)
	}
	if hits, _ := rec.snapshot(); hits != 1 {
		t.Errorf("build-logs requests = %d, want 1", hits)
	}
}
