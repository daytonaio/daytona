// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"errors"
	"testing"

	"github.com/daytonaio/daytona/cli/internal/clierr"
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
