// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestParseCpEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		wantSandbox string
		wantPath    string
		wantRemote  bool
	}{
		{name: "plain local path", arg: "foo/bar.txt", wantPath: "foo/bar.txt"},
		{name: "sandbox with absolute path", arg: "box:/tmp/a", wantSandbox: "box", wantPath: "/tmp/a", wantRemote: true},
		{name: "sandbox with relative path", arg: "my-sandbox:data/x.csv", wantSandbox: "my-sandbox", wantPath: "data/x.csv", wantRemote: true},
		{name: "windows drive letter absolute", arg: `C:\foo\bar`, wantPath: `C:\foo\bar`},
		{name: "windows drive letter relative", arg: "C:foo", wantPath: "C:foo"},
		{name: "dot relative path without colon", arg: "./x", wantPath: "./x"},
		{name: "dot prefix with colon", arg: ".:y", wantPath: ".:y"},
		{name: "dot-dot prefix with colon", arg: "..:y", wantPath: "..:y"},
		{name: "empty prefix", arg: ":/tmp/a", wantPath: ":/tmp/a"},
		{name: "colon in remote path", arg: "box:/a:b", wantSandbox: "box", wantPath: "/a:b", wantRemote: true},
		{name: "remote with empty path", arg: "box:", wantSandbox: "box", wantPath: "", wantRemote: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sandbox, filePath, remote := parseCpEndpoint(tt.arg)
			if remote != tt.wantRemote {
				t.Fatalf("parseCpEndpoint(%q) remote = %v, want %v", tt.arg, remote, tt.wantRemote)
			}
			if sandbox != tt.wantSandbox {
				t.Errorf("parseCpEndpoint(%q) sandbox = %q, want %q", tt.arg, sandbox, tt.wantSandbox)
			}
			if filePath != tt.wantPath {
				t.Errorf("parseCpEndpoint(%q) path = %q, want %q", tt.arg, filePath, tt.wantPath)
			}
		})
	}
}

func TestParseCpArgs(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		destination    string
		wantErr        bool
		wantUpload     bool
		wantSandboxRef string
		wantRemotePath string
		wantLocalPath  string
	}{
		{
			name:           "upload direction",
			source:         "./local.txt",
			destination:    "box:/tmp/remote.txt",
			wantUpload:     true,
			wantSandboxRef: "box",
			wantRemotePath: "/tmp/remote.txt",
			wantLocalPath:  "./local.txt",
		},
		{
			name:           "download direction",
			source:         "my-sandbox:/var/log/app.log",
			destination:    "out/app.log",
			wantUpload:     false,
			wantSandboxRef: "my-sandbox",
			wantRemotePath: "/var/log/app.log",
			wantLocalPath:  "out/app.log",
		},
		{
			name:           "empty remote path defaults to working directory",
			source:         "box:",
			destination:    "out",
			wantUpload:     false,
			wantSandboxRef: "box",
			wantRemotePath: ".",
			wantLocalPath:  "out",
		},
		{
			name:           "windows source stays local",
			source:         `C:\data\file.bin`,
			destination:    "box:/tmp/file.bin",
			wantUpload:     true,
			wantSandboxRef: "box",
			wantRemotePath: "/tmp/file.bin",
			wantLocalPath:  `C:\data\file.bin`,
		},
		{name: "both remote rejected", source: "box-a:/tmp/x", destination: "box-b:/tmp/y", wantErr: true},
		{name: "both local rejected", source: "a.txt", destination: "b.txt", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseCpArgs(tt.source, tt.destination)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseCpArgs(%q, %q) expected error, got nil", tt.source, tt.destination)
				}
				var cliErr *clierr.Error
				if !errors.As(err, &cliErr) {
					t.Fatalf("parseCpArgs(%q, %q) error type = %T, want *clierr.Error", tt.source, tt.destination, err)
				}
				if cliErr.Category != clierr.CategoryUsage {
					t.Errorf("parseCpArgs(%q, %q) error category = %q, want %q", tt.source, tt.destination, cliErr.Category, clierr.CategoryUsage)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseCpArgs(%q, %q) unexpected error: %v", tt.source, tt.destination, err)
			}
			if req.upload != tt.wantUpload {
				t.Errorf("parseCpArgs(%q, %q) upload = %v, want %v", tt.source, tt.destination, req.upload, tt.wantUpload)
			}
			if req.sandboxRef != tt.wantSandboxRef {
				t.Errorf("parseCpArgs(%q, %q) sandboxRef = %q, want %q", tt.source, tt.destination, req.sandboxRef, tt.wantSandboxRef)
			}
			if req.remotePath != tt.wantRemotePath {
				t.Errorf("parseCpArgs(%q, %q) remotePath = %q, want %q", tt.source, tt.destination, req.remotePath, tt.wantRemotePath)
			}
			if req.localPath != tt.wantLocalPath {
				t.Errorf("parseCpArgs(%q, %q) localPath = %q, want %q", tt.source, tt.destination, req.localPath, tt.wantLocalPath)
			}
		})
	}
}

func TestCpRequireSourceAndDestination(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{name: "no args", args: nil, wantErr: "missing required arguments: SOURCE and DESTINATION"},
		{name: "one arg", args: []string{"a.txt"}, wantErr: "missing required arguments: SOURCE and DESTINATION"},
		{name: "two args", args: []string{"a.txt", "box:/tmp/a.txt"}},
		{name: "three args", args: []string{"a", "b", "c"}, wantErr: "expected exactly 2 arguments (SOURCE and DESTINATION), received 3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cpRequireSourceAndDestination(nil, tt.args)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("cpRequireSourceAndDestination(%v) unexpected error: %v", tt.args, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("cpRequireSourceAndDestination(%v) expected error, got nil", tt.args)
			}
			if !clierr.HasCategory(err, clierr.CategoryUsage) {
				t.Errorf("cpRequireSourceAndDestination(%v) error %v is not usage-category", tt.args, err)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("cpRequireSourceAndDestination(%v) error = %q, want %q", tt.args, err.Error(), tt.wantErr)
			}
			if code := clierr.ExitCode(err); code != 2 {
				t.Errorf("cpRequireSourceAndDestination(%v) exit code = %d, want 2", tt.args, code)
			}
		})
	}
}

const cpTestAuthHeader = "Bearer cp-test-key"

// cpTestApiClient starts an httptest server around mux and returns a
// generated API client pointed at it, with a static Authorization header so
// handlers can assert auth propagation.
func cpTestApiClient(t *testing.T, mux *http.ServeMux) *apiclient.APIClient {
	t.Helper()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	clientConfig := apiclient.NewConfiguration()
	clientConfig.Servers = apiclient.ServerConfigurations{{URL: server.URL}}
	clientConfig.AddDefaultHeader("Authorization", cpTestAuthHeader)
	return apiclient.NewAPIClient(clientConfig)
}

// cpTestFileInfoJSON renders a FileInfo payload with every property the
// generated client requires.
func cpTestFileInfoJSON(name string, isDir bool, size int) string {
	return fmt.Sprintf(`{
		"name": %q,
		"isDir": %t,
		"size": %d,
		"modTime": "2026-01-01T00:00:00Z",
		"mode": "-rw-r--r--",
		"permissions": "644",
		"owner": "daytona",
		"group": "daytona"
	}`, name, isDir, size)
}

func TestCpUploadFileSuccess(t *testing.T) {
	const content = "hello from local"
	localPath := filepath.Join(t.TempDir(), "hello.txt")
	if err := os.WriteFile(localPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	var uploadHits int
	var uploadAuth, uploadPath, uploadBody string

	mux := http.NewServeMux()
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error":"file not found"}`)
	})
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/folder", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadHits++
		uploadAuth = r.Header.Get("Authorization")
		uploadPath = r.URL.Query().Get("path")
		file, _, err := r.FormFile("file")
		if err != nil {
			t.Errorf("upload request has no multipart %q field: %v", "file", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer func() { _ = file.Close() }()
		body, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("reading multipart file: %v", err)
		}
		uploadBody = string(body)
		w.WriteHeader(http.StatusOK)
	})

	apiClient := cpTestApiClient(t, mux)

	n, err := cpUpload(context.Background(), apiClient, "sbx-1", "box", localPath, "/workspace/hello.txt")
	if err != nil {
		t.Fatalf("cpUpload() unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("cpUpload() transferred = %d, want 1", n)
	}
	if uploadHits != 1 {
		t.Fatalf("upload endpoint hit %d times, want 1", uploadHits)
	}
	if uploadAuth != cpTestAuthHeader {
		t.Errorf("upload Authorization header = %q, want %q", uploadAuth, cpTestAuthHeader)
	}
	if uploadPath != "/workspace/hello.txt" {
		t.Errorf("upload path query = %q, want %q", uploadPath, "/workspace/hello.txt")
	}
	if uploadBody != content {
		t.Errorf("uploaded body = %q, want %q", uploadBody, content)
	}
}

func TestCpDownloadFileSuccess(t *testing.T) {
	const content = "hello from sandbox"

	mux := http.NewServeMux()
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, cpTestFileInfoJSON("out.txt", false, len(content)))
	})
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/download", func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != cpTestAuthHeader {
			t.Errorf("download Authorization header = %q, want %q", auth, cpTestAuthHeader)
		}
		_, _ = fmt.Fprint(w, content)
	})

	apiClient := cpTestApiClient(t, mux)

	localDst := filepath.Join(t.TempDir(), "nested", "out.txt")
	n, err := cpDownload(context.Background(), apiClient, "sbx-1", "box", "/workspace/out.txt", localDst)
	if err != nil {
		t.Fatalf("cpDownload() unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("cpDownload() transferred = %d, want 1", n)
	}
	got, err := os.ReadFile(localDst)
	if err != nil {
		t.Fatalf("reading downloaded file: %v", err)
	}
	if string(got) != content {
		t.Errorf("downloaded content = %q, want %q", got, content)
	}
}

// Regression test: the generated client returns a nil *os.File for empty
// response bodies (decode short-circuits on len(b)==0), so an empty remote
// file must still produce an empty local file.
func TestCpDownloadEmptyFile(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, cpTestFileInfoJSON("empty.txt", false, 0))
	})
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/download", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	apiClient := cpTestApiClient(t, mux)

	t.Run("creates missing file and parent directories", func(t *testing.T) {
		localDst := filepath.Join(t.TempDir(), "nested", "empty.txt")
		n, err := cpDownload(context.Background(), apiClient, "sbx-1", "box", "/workspace/empty.txt", localDst)
		if err != nil {
			t.Fatalf("cpDownload() unexpected error: %v", err)
		}
		if n != 1 {
			t.Errorf("cpDownload() transferred = %d, want 1", n)
		}
		info, err := os.Stat(localDst)
		if err != nil {
			t.Fatalf("empty local file was not created: %v", err)
		}
		if info.Size() != 0 {
			t.Errorf("local file size = %d, want 0", info.Size())
		}
	})

	t.Run("truncates existing file", func(t *testing.T) {
		localDst := filepath.Join(t.TempDir(), "stale.txt")
		if err := os.WriteFile(localDst, []byte("stale content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := cpDownload(context.Background(), apiClient, "sbx-1", "box", "/workspace/empty.txt", localDst); err != nil {
			t.Fatalf("cpDownload() unexpected error: %v", err)
		}
		info, err := os.Stat(localDst)
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() != 0 {
			t.Errorf("local file size after empty download = %d, want 0 (truncated)", info.Size())
		}
	})
}

func TestCpDownloadRemoteNotFound(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/toolbox/sbx-1/toolbox/files/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"error":"file not found"}`)
	})

	apiClient := cpTestApiClient(t, mux)

	localDst := filepath.Join(t.TempDir(), "out.txt")
	_, err := cpDownload(context.Background(), apiClient, "sbx-1", "box", "/workspace/missing.txt", localDst)
	if err == nil {
		t.Fatal("cpDownload() expected error for missing remote file, got nil")
	}
	if !clierr.HasCategory(err, clierr.CategoryNotFound) {
		t.Errorf("cpDownload() error %v is not not_found-category", err)
	}
	if !strings.Contains(err.Error(), "file not found") {
		t.Errorf("cpDownload() error = %q, want it to preserve the server message", err.Error())
	}
	if _, statErr := os.Stat(localDst); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("local destination %q should not exist after failed download, stat err = %v", localDst, statErr)
	}
}
