// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package pty

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestExtractPtyEnvsSubprotocol(t *testing.T) {
	encoded := base64.RawURLEncoding.EncodeToString([]byte(`{"FOO":"bar"}`))
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "X-Daytona-SDK-Version~1.2.3, X-Daytona-Pty-Envs~"+encoded)

	envs, err := extractPtyEnvsSubprotocol(header)
	if err != nil {
		t.Fatalf("extractPtyEnvsSubprotocol() unexpected error: %v", err)
	}
	if len(envs) != 1 || envs["FOO"] != "bar" {
		t.Fatalf("extractPtyEnvsSubprotocol() = %v, want map[FOO:bar]", envs)
	}
}

func TestExtractPtyEnvsSubprotocol_Absent(t *testing.T) {
	// No Sec-WebSocket-Protocol header at all.
	envs, err := extractPtyEnvsSubprotocol(http.Header{})
	if err != nil {
		t.Fatalf("extractPtyEnvsSubprotocol(empty) unexpected error: %v", err)
	}
	if len(envs) != 0 {
		t.Fatalf("extractPtyEnvsSubprotocol(empty) = %v, want empty map", envs)
	}

	// Header present but envs token absent.
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "X-Daytona-SDK-Version~1.2.3")
	envs, err = extractPtyEnvsSubprotocol(header)
	if err != nil {
		t.Fatalf("extractPtyEnvsSubprotocol(no-token) unexpected error: %v", err)
	}
	if len(envs) != 0 {
		t.Fatalf("extractPtyEnvsSubprotocol(no-token) = %v, want empty map", envs)
	}
}

func TestExtractPtyEnvsSubprotocol_Invalid(t *testing.T) {
	// Invalid base64.
	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", "X-Daytona-Pty-Envs~not!valid!base64")
	if _, err := extractPtyEnvsSubprotocol(header); err == nil {
		t.Fatal("extractPtyEnvsSubprotocol(bad base64) expected error, got nil")
	}

	// Valid base64 but invalid JSON.
	encoded := base64.RawURLEncoding.EncodeToString([]byte(`not json`))
	header = http.Header{}
	header.Set("Sec-WebSocket-Protocol", "X-Daytona-Pty-Envs~"+encoded)
	if _, err := extractPtyEnvsSubprotocol(header); err == nil {
		t.Fatal("extractPtyEnvsSubprotocol(bad json) expected error, got nil")
	}
}
