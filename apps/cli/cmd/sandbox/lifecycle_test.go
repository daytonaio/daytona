// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package sandbox

import (
	"encoding/json"
	"errors"
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
