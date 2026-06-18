// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package snapshot

import (
	"errors"
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

func TestIsSnapshotBuildDone(t *testing.T) {
	tests := []struct {
		state      apiclient.SnapshotState
		wantDone   bool
		wantFailed bool
	}{
		{apiclient.SNAPSHOTSTATE_ACTIVE, true, false},
		{apiclient.SNAPSHOTSTATE_INACTIVE, true, false},
		{apiclient.SNAPSHOTSTATE_ERROR, true, true},
		{apiclient.SNAPSHOTSTATE_BUILD_FAILED, true, true},
		{apiclient.SNAPSHOTSTATE_BUILDING, false, false},
		{apiclient.SNAPSHOTSTATE_PENDING, false, false},
		{apiclient.SNAPSHOTSTATE_PULLING, false, false},
		{apiclient.SNAPSHOTSTATE_REMOVING, false, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			done, failed := isSnapshotBuildDone(tt.state)
			if done != tt.wantDone || failed != tt.wantFailed {
				t.Errorf("isSnapshotBuildDone(%q) = (%v, %v), want (%v, %v)", tt.state, done, failed, tt.wantDone, tt.wantFailed)
			}
		})
	}
}

func TestRequireSnapshotArg(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args", args: nil, wantErr: true},
		{name: "one arg", args: []string{"my-snapshot"}, wantErr: false},
		{name: "two args", args: []string{"a", "b"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireSnapshotArg(nil, tt.args)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("requireSnapshotArg(%v) unexpected error: %v", tt.args, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("requireSnapshotArg(%v) expected error, got nil", tt.args)
			}
			var cliErr *clierr.Error
			if !errors.As(err, &cliErr) {
				t.Fatalf("requireSnapshotArg(%v) expected *clierr.Error, got %T", tt.args, err)
			}
			if cliErr.Category != clierr.CategoryUsage {
				t.Errorf("requireSnapshotArg(%v) category = %q, want %q", tt.args, cliErr.Category, clierr.CategoryUsage)
			}
		})
	}
}

func TestRequireSnapshotArgMissingArgMessage(t *testing.T) {
	err := requireSnapshotArg(nil, nil)
	if err == nil {
		t.Fatal("requireSnapshotArg(nil) expected error, got nil")
	}
	want := "missing required argument: snapshot ID or name"
	if err.Error() != want {
		t.Errorf("requireSnapshotArg(nil) error = %q, want %q", err.Error(), want)
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

// logsTestSnapshotJSON renders a minimal SnapshotDto payload containing every
// property the generated client requires plus the given state and error reason.
func logsTestSnapshotJSON(state, errorReason string) string {
	reason := "null"
	if errorReason != "" {
		reason = fmt.Sprintf("%q", errorReason)
	}
	return fmt.Sprintf(`{
		"id": "snap-1",
		"general": false,
		"name": "my-snapshot",
		"state": %q,
		"size": null,
		"entrypoint": [],
		"cpu": 1,
		"gpu": 0,
		"mem": 1,
		"disk": 1,
		"errorReason": %s,
		"createdAt": "2025-01-01T00:00:00Z",
		"updatedAt": "2025-01-01T00:00:00Z",
		"lastUsedAt": null
	}`, state, reason)
}

// logsTestServer emulates the API routes the logs command hits: GetSnapshot
// and the raw build-logs stream.
func logsTestServer(t *testing.T, state, errorReason, logBody string, rec *logsRequestRecorder) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/snapshots/snap-1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, logsTestSnapshotJSON(state, errorReason))
	})
	mux.HandleFunc("/snapshots/snap-1/build-logs", func(w http.ResponseWriter, r *http.Request) {
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
	server := logsTestServer(t, "active", "", body, rec)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, false)

	var err error
	out := logsCaptureStdout(t, func() {
		err = LogsCmd.RunE(LogsCmd, []string{"snap-1"})
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
		_, _ = fmt.Fprint(w, `{"message":"snapshot not found","statusCode":404}`)
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
// follow mode truncating output when the snapshot is already terminal at
// entry: the complete log history must arrive via exactly one plain
// (non-follow) fetch.
func TestLogsCmdFollowAlreadyTerminalCompleteFetch(t *testing.T) {
	body := strings.Repeat("a long line of build output that must not be truncated\n", 200)
	rec := &logsRequestRecorder{}
	server := logsTestServer(t, "active", "", body, rec)
	logsTestEnv(t, server.URL)
	logsSetFollowFlag(t, true)

	var err error
	out := logsCaptureStdout(t, func() {
		err = LogsCmd.RunE(LogsCmd, []string{"snap-1"})
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
		err = LogsCmd.RunE(LogsCmd, []string{"snap-1"})
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
