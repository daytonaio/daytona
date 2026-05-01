// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"testing"

	common_proxy "github.com/daytonaio/common-go/pkg/proxy"
)

func TestToolboxBaseURL(t *testing.T) {
	got := toolboxBaseURL("https", "proxy.example.com", "sandbox-1")
	want := "https://proxy.example.com/toolbox/sandbox-1"
	if got != want {
		t.Fatalf("toolbox base URL = %q, want %q", got, want)
	}
}

func TestToolboxForwardedHeaders(t *testing.T) {
	headers := toolboxForwardedHeaders("https", "proxy.example.com", "sandbox-1")

	checks := map[string]string{
		common_proxy.DaytonaToolboxBaseURLHeader: "https://proxy.example.com/toolbox/sandbox-1",
		"X-Forwarded-Proto":                      "https",
		"X-Forwarded-Prefix":                     "/toolbox/sandbox-1",
	}
	for key, want := range checks {
		if got := headers[key]; got != want {
			t.Fatalf("%s = %q, want %q", key, got, want)
		}
	}
}
