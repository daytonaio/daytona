// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package services

import "testing"

func TestPushInFlight(t *testing.T) {
	cases := []struct {
		name    string
		entries map[string]*degradedEntry
		want    bool
	}{
		{
			name:    "untracked sandbox is not in flight",
			entries: map[string]*degradedEntry{},
			want:    false,
		},
		{
			name:    "tracked entry without active push is not in flight",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", pushing: false}},
			want:    false,
		},
		{
			name:    "tracked entry with active push blocks the clear",
			entries: map[string]*degradedEntry{"sb": {reason: "fd exhaustion", pushing: true}},
			want:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := &SandboxDegradedService{entries: tc.entries}
			if got := s.pushInFlight("sb"); got != tc.want {
				t.Errorf("pushInFlight() = %v, want %v", got, tc.want)
			}
		})
	}
}
