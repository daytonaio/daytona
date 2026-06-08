// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package organization

import (
	"testing"

	apiclient "github.com/daytonaio/daytona/libs/api-client-go"
)

func TestResolveRegion(t *testing.T) {
	regions := []apiclient.Region{
		{Id: "abc-123", Name: "us"},
		{Id: "def-456", Name: "eu"},
	}

	tests := []struct {
		name       string
		identifier string
		wantId     string
		wantErr    bool
	}{
		{name: "match by id", identifier: "abc-123", wantId: "abc-123"},
		{name: "match by name", identifier: "eu", wantId: "def-456"},
		{name: "no match returns error", identifier: "nonexistent", wantErr: true},
		{name: "empty identifier returns error", identifier: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRegion(regions, tt.identifier)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveRegion(%q) expected error, got nil", tt.identifier)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveRegion(%q) unexpected error: %v", tt.identifier, err)
			}
			if got == nil {
				t.Fatalf("resolveRegion(%q) returned nil region", tt.identifier)
			}
			if got.Id != tt.wantId {
				t.Errorf("resolveRegion(%q) returned id %q, want %q", tt.identifier, got.Id, tt.wantId)
			}
		})
	}
}
