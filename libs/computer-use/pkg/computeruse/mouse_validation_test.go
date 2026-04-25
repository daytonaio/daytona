// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import "testing"

func TestNormalizeMouseButton(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "defaults empty to left", input: "", want: defaultMouseButton},
		{name: "trims whitespace", input: " right ", want: "right"},
		{name: "keeps left", input: "left", want: defaultMouseButton},
		{name: "keeps middle", input: middleMouseButton, want: middleMouseButton},
		{name: "normalizes casing", input: "Middle", want: middleMouseButton},
		{name: "rejects center", input: "center", wantErr: true},
		{name: "rejects unsupported button", input: "wheel", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeMouseButton(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeMouseButton(%q) returned error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("normalizeMouseButton(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeScrollDirection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "accepts up", input: scrollDirectionUp, want: scrollDirectionUp},
		{name: "accepts down", input: scrollDirectionDown, want: scrollDirectionDown},
		{name: "normalizes casing", input: "Up", want: scrollDirectionUp},
		{name: "trims whitespace", input: " down ", want: scrollDirectionDown},
		{name: "rejects empty", input: "", wantErr: true},
		{name: "rejects unsupported direction", input: "left", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeScrollDirection(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeScrollDirection(%q) returned error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("normalizeScrollDirection(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeScrollAmount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   int
		want    int
		wantErr bool
	}{
		{name: "defaults zero to one", input: 0, want: defaultScrollAmount},
		{name: "keeps positive amount", input: 3, want: 3},
		{name: "rejects negative amount", input: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeScrollAmount(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %d", tt.input)
				}
				return
			}

			if err != nil {
				t.Fatalf("normalizeScrollAmount(%d) returned error: %v", tt.input, err)
			}

			if got != tt.want {
				t.Fatalf("normalizeScrollAmount(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}
