// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import "testing"

func TestBuildExtraHeaders(t *testing.T) {
	cases := []struct {
		name                   string
		apiKey                 string
		incomingXForwardedHost string
		requestHost            string
		wantXForwardedHost     string
		wantXForwardedHostSet  bool
	}{
		{
			name:                   "no incoming X-Forwarded-Host: set from request host",
			apiKey:                 "tok-123",
			incomingXForwardedHost: "",
			requestHost:            "3000-sandbox.proxy.daytona.work",
			wantXForwardedHost:     "3000-sandbox.proxy.daytona.work",
			wantXForwardedHostSet:  true,
		},
		{
			name:                   "existing X-Forwarded-Host: not overwritten (upstream preserved)",
			apiKey:                 "tok-456",
			incomingXForwardedHost: "customer.example.com",
			requestHost:            "3000-sandbox.proxy.daytona.work",
			wantXForwardedHostSet:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := BuildExtraHeaders(tc.apiKey, tc.incomingXForwardedHost, tc.requestHost)

			if got["X-Daytona-Authorization"] != "Bearer "+tc.apiKey {
				t.Errorf("X-Daytona-Authorization = %q, want %q", got["X-Daytona-Authorization"], "Bearer "+tc.apiKey)
			}

			val, ok := got["X-Forwarded-Host"]
			if tc.wantXForwardedHostSet {
				if !ok || val != tc.wantXForwardedHost {
					t.Errorf("X-Forwarded-Host = (%q, present=%v), want (%q, present=true)", val, ok, tc.wantXForwardedHost)
				}
			} else {
				if ok {
					t.Errorf("X-Forwarded-Host should not be set when upstream already carries one, got %q", val)
				}
			}
		})
	}
}
