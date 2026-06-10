// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	"github.com/daytonaio/daytona/cli/toolbox"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestExecFailure(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		wantCategory clierr.Category
		wantMessage  string
		wantHint     string
	}{
		{
			name:         "clierr passthrough keeps category and hint",
			err:          clierr.New(clierr.CategoryNotFound, "sandbox not found").WithHint("check the sandbox ID"),
			wantCategory: clierr.CategoryNotFound,
			wantMessage:  "sandbox not found",
			wantHint:     "check the sandbox ID",
		},
		{
			name:         "wrapped clierr is unwrapped and keeps category",
			err:          fmt.Errorf("toolbox: %w", clierr.New(clierr.CategoryAuth, "unauthorized")),
			wantCategory: clierr.CategoryAuth,
			wantMessage:  "unauthorized",
		},
		{
			name:         "plain error becomes server-category clierr",
			err:          errors.New("connection reset"),
			wantCategory: clierr.CategoryServer,
			wantMessage:  "connection reset",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := execFailure(tt.err)

			var cliErr *clierr.Error
			if !errors.As(got, &cliErr) {
				t.Fatalf("execFailure(%v) = %T, want *clierr.Error", tt.err, got)
			}
			if cliErr.Code != 255 {
				t.Errorf("Code = %d, want 255", cliErr.Code)
			}
			if cliErr.Category != tt.wantCategory {
				t.Errorf("Category = %q, want %q", cliErr.Category, tt.wantCategory)
			}
			if cliErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", cliErr.Message, tt.wantMessage)
			}
			if cliErr.Hint != tt.wantHint {
				t.Errorf("Hint = %q, want %q", cliErr.Hint, tt.wantHint)
			}
			if exitCode := clierr.ExitCode(got); exitCode != 255 {
				t.Errorf("ExitCode = %d, want 255", exitCode)
			}
		})
	}
}

func TestExecResultTags(t *testing.T) {
	typ := reflect.TypeOf(execResult{})

	tests := []struct {
		field string
		tag   string
	}{
		{field: "Result", tag: "result"},
		{field: "ExitCode", tag: "exitCode"},
	}

	for _, tt := range tests {
		t.Run(tt.field, func(t *testing.T) {
			f, ok := typ.FieldByName(tt.field)
			if !ok {
				t.Fatalf("execResult has no field %s", tt.field)
			}
			if got := f.Tag.Get("json"); got != tt.tag {
				t.Errorf("json tag = %q, want %q", got, tt.tag)
			}
			if got := f.Tag.Get("yaml"); got != tt.tag {
				t.Errorf("yaml tag = %q, want %q", got, tt.tag)
			}
		})
	}
}

func TestExecArgsValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "no arguments", args: nil, wantErr: true},
		{name: "sandbox without command", args: []string{"my-sandbox"}, wantErr: true},
		{name: "sandbox and command", args: []string{"my-sandbox", "ls"}, wantErr: false},
		{name: "sandbox and command with args", args: []string{"my-sandbox", "ls", "-la"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ExecCmd.Args(ExecCmd, tt.args)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Args(%v) unexpected error: %v", tt.args, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Args(%v) expected error, got nil", tt.args)
			}
			if !clierr.HasCategory(err, clierr.CategoryUsage) {
				t.Errorf("HasCategory(err, usage) = false, err = %v", err)
			}
			if want := "missing required arguments: sandbox and command"; err.Error() != want {
				t.Errorf("error = %q, want %q", err.Error(), want)
			}
			if got := clierr.ExitCode(err); got != 2 {
				t.Errorf("ExitCode = %d, want 2", got)
			}
		})
	}
}

func TestExecSuccessPrintsResult(t *testing.T) {
	mux := http.NewServeMux()
	server := newTestAPIServer(t, mux)

	var gotCommand string
	mux.HandleFunc("GET /sandbox/my-sandbox", func(w http.ResponseWriter, r *http.Request) {
		writeSandboxJSON(t, w, "started")
	})
	mux.HandleFunc("GET /sandbox/sbx-1/toolbox-proxy-url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"url":%q}`, server.URL+"/proxy"); err != nil {
			t.Errorf("writing proxy-url payload: %v", err)
		}
	})
	mux.HandleFunc("POST /proxy/sbx-1/process/execute", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Command string `json:"command"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decoding execute request: %v", err)
		}
		gotCommand = req.Command
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, `{"result":"hello\n","exitCode":0}`); err != nil {
			t.Errorf("writing execute payload: %v", err)
		}
	})

	out, err := captureStdout(t, func() error {
		return ExecCmd.RunE(ExecCmd, []string{"my-sandbox", "echo", "hello"})
	})
	if err != nil {
		t.Fatalf("ExecCmd.RunE() unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Errorf("stdout = %q, want %q", out, "hello\n")
	}
	if gotCommand != "echo hello" {
		t.Errorf("proxy received command %q, want %q", gotCommand, "echo hello")
	}
}

// TestExecRemoteFailureResultJSON exercises the toolbox-proxy path for a
// remote command exiting non-zero and asserts the JSON shape of the result
// struct. The full RunE path cannot be used here: a non-zero remote exit
// code makes the command call os.Exit, which would kill the test process.
func TestExecRemoteFailureResultJSON(t *testing.T) {
	mux := http.NewServeMux()
	server := newTestAPIServer(t, mux)

	mux.HandleFunc("GET /sandbox/sbx-1/toolbox-proxy-url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprintf(w, `{"url":%q}`, server.URL+"/proxy"); err != nil {
			t.Errorf("writing proxy-url payload: %v", err)
		}
	})
	mux.HandleFunc("POST /proxy/sbx-1/process/execute", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := fmt.Fprint(w, `{"result":"boom","exitCode":3}`); err != nil {
			t.Errorf("writing execute payload: %v", err)
		}
	})

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{{URL: server.URL}}
	client := apiclient.NewAPIClient(clientConfig)

	sandbox := apiclient.Sandbox{Id: "sbx-1", Target: "us"}
	response, err := toolbox.NewClient(client).ExecuteCommand(context.Background(), &sandbox, toolbox.ExecuteRequest{Command: "exit 3"})
	if err != nil {
		t.Fatalf("ExecuteCommand() unexpected error: %v", err)
	}

	data, err := json.Marshal(execResult{Result: response.Result, ExitCode: int(response.ExitCode)})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if want := `{"result":"boom","exitCode":3}`; string(data) != want {
		t.Errorf("execResult JSON = %s, want %s", data, want)
	}
}

func TestExecGetSandboxNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /sandbox/missing", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if _, err := fmt.Fprint(w, `{"error":"Sandbox not found"}`); err != nil {
			t.Errorf("writing not-found payload: %v", err)
		}
	})
	newTestAPIServer(t, mux)

	err := ExecCmd.RunE(ExecCmd, []string{"missing", "ls"})
	if err == nil {
		t.Fatal("expected error for missing sandbox, got nil")
	}
	if !clierr.HasCategory(err, clierr.CategoryNotFound) {
		t.Errorf("HasCategory(err, not_found) = false, err = %v", err)
	}
	var cliErr *clierr.Error
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *clierr.Error, got %T: %v", err, err)
	}
	if cliErr.Code != 255 {
		t.Errorf("Code = %d, want 255", cliErr.Code)
	}
	if got := clierr.ExitCode(err); got != 255 {
		t.Errorf("ExitCode = %d, want 255", got)
	}
}
