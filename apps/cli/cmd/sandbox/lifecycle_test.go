// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
	"gopkg.in/yaml.v2"
)

func TestRequireSandboxArg(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no args returns usage error", args: []string{}, wantErr: true},
		{name: "one arg accepted", args: []string{"my-sandbox"}},
		{name: "two args returns usage error", args: []string{"my-sandbox", "extra"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := requireSandboxArg(nil, tt.args)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("requireSandboxArg(%v) unexpected error: %v", tt.args, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("requireSandboxArg(%v) expected error, got nil", tt.args)
			}
			var cliErr *clierr.Error
			if !errors.As(err, &cliErr) {
				t.Fatalf("requireSandboxArg(%v) expected *clierr.Error, got %T", tt.args, err)
			}
			if cliErr.Category != clierr.CategoryUsage {
				t.Errorf("requireSandboxArg(%v) category = %q, want %q", tt.args, cliErr.Category, clierr.CategoryUsage)
			}
		})
	}
}

func TestRequireSandboxArgMissingArgMessage(t *testing.T) {
	err := requireSandboxArg(nil, nil)
	if err == nil {
		t.Fatal("requireSandboxArg(nil) expected error, got nil")
	}
	want := "missing required argument: sandbox ID or name"
	if err.Error() != want {
		t.Errorf("requireSandboxArg(nil) error = %q, want %q", err.Error(), want)
	}
}

func TestNewPreviewUrlOutput(t *testing.T) {
	previewUrl := &apiclient.SignedPortPreviewUrl{
		SandboxId: "sb-123",
		Port:      8080,
		Token:     "secret",
		Url:       "https://8080-sb-123.preview.daytona.io?token=secret",
	}

	out := newPreviewUrlOutput(previewUrl, 600, "my-sandbox")

	if out.Url != previewUrl.Url {
		t.Errorf("Url = %q, want %q", out.Url, previewUrl.Url)
	}
	if out.Port != previewUrl.Port {
		t.Errorf("Port = %d, want %d", out.Port, previewUrl.Port)
	}
	if out.ExpiresInSeconds != 600 {
		t.Errorf("ExpiresInSeconds = %d, want 600", out.ExpiresInSeconds)
	}
	if out.Sandbox != "my-sandbox" {
		t.Errorf("Sandbox = %q, want %q (the user-supplied argument)", out.Sandbox, "my-sandbox")
	}
}

func TestPreviewUrlOutputFieldNames(t *testing.T) {
	out := newPreviewUrlOutput(&apiclient.SignedPortPreviewUrl{Port: 3000, Url: "https://example"}, 3600, "my-sandbox")
	wantKeys := []string{"url", "port", "expiresInSeconds", "sandbox"}

	jsonBytes, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	var jsonMap map[string]any
	if err := json.Unmarshal(jsonBytes, &jsonMap); err != nil {
		t.Fatalf("json.Unmarshal failed: %v", err)
	}

	yamlBytes, err := yaml.Marshal(out)
	if err != nil {
		t.Fatalf("yaml.Marshal failed: %v", err)
	}
	yamlMap := map[string]any{}
	if err := yaml.Unmarshal(yamlBytes, &yamlMap); err != nil {
		t.Fatalf("yaml.Unmarshal failed: %v", err)
	}

	for _, key := range wantKeys {
		if _, ok := jsonMap[key]; !ok {
			t.Errorf("json output missing key %q (got %v)", key, jsonMap)
		}
		if _, ok := yamlMap[key]; !ok {
			t.Errorf("yaml output missing key %q (got %v)", key, yamlMap)
		}
	}
	if len(jsonMap) != len(wantKeys) {
		t.Errorf("json output has %d keys, want %d: %v", len(jsonMap), len(wantKeys), jsonMap)
	}
}

// newTestAPIServer starts an httptest server backed by mux and points the
// CLI at it through the env-based profile (DAYTONA_API_KEY/DAYTONA_API_URL),
// sandboxing the profile/toolbox-proxy cache in a temp DAYTONA_CONFIG_DIR.
func newTestAPIServer(t *testing.T, mux *http.ServeMux) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	configDir := t.TempDir()
	// Pre-create config.json so config.GetConfig does not run the first-time
	// autocompletion setup, which writes to the user's shell profile.
	if err := os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{"activeProfile":"","profiles":[]}`), 0o600); err != nil {
		t.Fatalf("writing config.json: %v", err)
	}
	t.Setenv("DAYTONA_CONFIG_DIR", configDir)
	t.Setenv("DAYTONA_API_KEY", "test-api-key")
	t.Setenv("DAYTONA_API_URL", server.URL)
	return server
}

// testSandboxJSON renders a minimal Sandbox payload containing every
// property the generated client requires plus the given state.
func testSandboxJSON(state string) string {
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
		"state": %q
	}`, state)
}

func writeSandboxJSON(t *testing.T, w http.ResponseWriter, state string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if _, err := fmt.Fprint(w, testSandboxJSON(state)); err != nil {
		t.Errorf("writing sandbox payload: %v", err)
	}
}

// captureStdout runs fn with os.Stdout redirected to a pipe and returns
// everything written along with fn's error.
func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	runErr := fn()

	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("reading captured stdout: %v", err)
	}
	return string(out), runErr
}

func TestStopAlreadyStoppedSkipsStopRequest(t *testing.T) {
	var stopCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sandbox/my-sandbox", func(w http.ResponseWriter, r *http.Request) {
		writeSandboxJSON(t, w, "stopped")
	})
	mux.HandleFunc("POST /sandbox/my-sandbox/stop", func(w http.ResponseWriter, r *http.Request) {
		stopCalls.Add(1)
	})
	newTestAPIServer(t, mux)

	out, err := captureStdout(t, func() error {
		return StopCmd.RunE(StopCmd, []string{"my-sandbox"})
	})
	if err != nil {
		t.Fatalf("StopCmd.RunE() unexpected error: %v", err)
	}
	if got := stopCalls.Load(); got != 0 {
		t.Errorf("stop endpoint called %d times, want 0", got)
	}
	if !strings.Contains(out, "already stopped") {
		t.Errorf("output %q does not mention the sandbox is already stopped", out)
	}
}

func TestStartAlreadyStartedPassthrough(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sandbox/my-sandbox/start", func(w http.ResponseWriter, r *http.Request) {
		// The server treats starting an already-started sandbox as success.
		writeSandboxJSON(t, w, "started")
	})
	newTestAPIServer(t, mux)

	_, err := captureStdout(t, func() error {
		return StartCmd.RunE(StartCmd, []string{"my-sandbox"})
	})
	if err != nil {
		t.Fatalf("StartCmd.RunE() unexpected error: %v", err)
	}
}

func TestStartWithoutWaitDoesNotPollSandbox(t *testing.T) {
	var getCalls atomic.Int32
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sandbox/my-sandbox/start", func(w http.ResponseWriter, r *http.Request) {
		writeSandboxJSON(t, w, "starting")
	})
	mux.HandleFunc("GET /sandbox/my-sandbox", func(w http.ResponseWriter, r *http.Request) {
		getCalls.Add(1)
		writeSandboxJSON(t, w, "starting")
	})
	newTestAPIServer(t, mux)

	oldWait := startWaitFlag
	startWaitFlag = false
	t.Cleanup(func() { startWaitFlag = oldWait })

	_, err := captureStdout(t, func() error {
		return StartCmd.RunE(StartCmd, []string{"my-sandbox"})
	})
	if err != nil {
		t.Fatalf("StartCmd.RunE() unexpected error: %v", err)
	}
	if got := getCalls.Load(); got != 0 {
		t.Errorf("GetSandbox polled %d times without --wait, want 0", got)
	}
}

func TestCreateIfExistsReuse(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /sandbox", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		if _, err := fmt.Fprint(w, `{"error":"Sandbox with name my-sandbox already exists"}`); err != nil {
			t.Errorf("writing conflict payload: %v", err)
		}
	})
	mux.HandleFunc("GET /sandbox/my-sandbox", func(w http.ResponseWriter, r *http.Request) {
		writeSandboxJSON(t, w, "started")
	})
	mux.HandleFunc("GET /sandbox/sbx-1/ports/22222/preview-url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, `{"sandboxId":"sbx-1","url":"https://preview.example.com","token":"tok"}`); err != nil {
			t.Errorf("writing preview-url payload: %v", err)
		}
	})
	newTestAPIServer(t, mux)

	oldName, oldIfExists := nameFlag, createIfExistsFlag
	nameFlag, createIfExistsFlag = "my-sandbox", "reuse"
	t.Cleanup(func() { nameFlag, createIfExistsFlag = oldName, oldIfExists })

	out, err := captureStdout(t, func() error {
		return CreateCmd.RunE(CreateCmd, nil)
	})
	if err != nil {
		t.Fatalf("CreateCmd.RunE() unexpected error: %v", err)
	}
	if !strings.Contains(out, "already exists, reusing it") {
		t.Errorf("output %q does not mention the sandbox is being reused", out)
	}
}
