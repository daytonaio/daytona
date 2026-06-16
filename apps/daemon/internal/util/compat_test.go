// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package util

import (
	"net/http"
	"testing"
)

func TestClientRejectsUnknownResponseFields(t *testing.T) {
	// Set so keys canonicalize as net/http does for incoming requests.
	header := func(kv ...string) http.Header {
		h := http.Header{}
		for i := 0; i+1 < len(kv); i += 2 {
			h.Set(kv[i], kv[i+1])
		}
		return h
	}

	tests := []struct {
		name   string
		header http.Header
		want   bool
	}{
		{
			name:   "no source header",
			header: header("X-Daytona-SDK-Version", "0.1.0"),
			want:   false,
		},
		{
			name:   "non-go source",
			header: header("X-Daytona-Source", "sdk-typescript", "X-Daytona-SDK-Version", "0.1.0"),
			want:   false,
		},
		{
			name:   "go source without version",
			header: header("X-Daytona-Source", "sdk-go"),
			want:   false,
		},
		{
			name:   "go source dev build",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "v0.0.0-dev"),
			want:   false,
		},
		{
			name:   "go source below threshold",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "0.187.0"),
			want:   true,
		},
		{
			name:   "go source one patch below threshold",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "0.187.9"),
			want:   true,
		},
		{
			name:   "go source at threshold",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "0.188.0"),
			want:   false,
		},
		{
			name:   "go source above threshold",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "0.189.1"),
			want:   false,
		},
		{
			name:   "go source unparseable version",
			header: header("X-Daytona-Source", "sdk-go", "X-Daytona-SDK-Version", "not-a-version"),
			want:   false,
		},
		{
			name:   "go source version via websocket subprotocol",
			header: header("X-Daytona-Source", "sdk-go", "Sec-WebSocket-Protocol", "X-Daytona-SDK-Version~0.187.0"),
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClientRejectsUnknownResponseFields(tt.header); got != tt.want {
				t.Errorf("ClientRejectsUnknownResponseFields() = %v, want %v", got, tt.want)
			}
		})
	}
}
