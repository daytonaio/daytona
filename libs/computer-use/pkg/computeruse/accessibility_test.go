// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"strings"
	"testing"

	"github.com/godbus/dbus/v5"
)

func TestNodeIDRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		sender string
		path   dbus.ObjectPath
	}{
		{"unique conn name", ":1.42", "/org/a11y/atspi/accessible/12"},
		{"well-known name", "org.a11y.atspi.Registry", "/org/a11y/atspi/accessible/root"},
		{"deep path", ":1.5", "/org/a11y/atspi/accessible/1/2/3"},
		{"short path", ":1.1", "/"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := makeNodeID(tc.sender, tc.path)
			sender, path, err := parseNodeID(id)
			if err != nil {
				t.Fatalf("parseNodeID(%q) returned error: %v", id, err)
			}
			if sender != tc.sender {
				t.Errorf("sender mismatch: got %q want %q", sender, tc.sender)
			}
			if path != tc.path {
				t.Errorf("path mismatch: got %q want %q", path, tc.path)
			}
		})
	}
}

func TestNodeIDParseErrors(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"empty", ""},
		{"no slash", ":1.42"},
		{"no separator", "foo/bar"},
		{"missing bus", ":/org/a11y/atspi/accessible/12"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := parseNodeID(tc.id); err == nil {
				t.Errorf("parseNodeID(%q) expected error, got nil", tc.id)
			}
		})
	}
}

// Regression test: the spec is explicit that the ':' separator is the one
// immediately before the first '/'. Unique connection names start with ':',
// which used to confuse a naive SplitN.
func TestNodeIDHandlesLeadingColonInBusName(t *testing.T) {
	const busName = ":1.123"
	const path = dbus.ObjectPath("/a/b/c")
	id := makeNodeID(busName, path)
	if !strings.HasPrefix(id, ":1.123:/a/b/c") {
		t.Fatalf("unexpected id shape: %q", id)
	}
	s, p, err := parseNodeID(id)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if s != busName || p != path {
		t.Fatalf("round-trip mismatch: got (%q,%q)", s, p)
	}
}

func TestFilterMatcher(t *testing.T) {
	button := &A11yNode{
		Role:   "push button",
		Name:   "Submit",
		States: []string{"enabled", "visible", "showing", "focusable"},
	}
	label := &A11yNode{
		Role:   "label",
		Name:   "Submit your form",
		States: []string{"visible", "showing"},
	}
	disabledBtn := &A11yNode{
		Role:   "push button",
		Name:   "Cancel",
		States: []string{"visible", "showing"},
	}

	cases := []struct {
		name    string
		filter  A11yFilter
		node    *A11yNode
		want    bool
		wantErr bool
	}{
		{
			name:   "role exact match",
			filter: A11yFilter{Role: "push button"},
			node:   button,
			want:   true,
		},
		{
			name:   "role case-insensitive",
			filter: A11yFilter{Role: "PUSH BUTTON"},
			node:   button,
			want:   true,
		},
		{
			name:   "role mismatch",
			filter: A11yFilter{Role: "label"},
			node:   button,
			want:   false,
		},
		{
			name:   "name substring default",
			filter: A11yFilter{Name: "Sub"},
			node:   button,
			want:   true,
		},
		{
			name:   "name substring matches longer",
			filter: A11yFilter{Name: "Submit"},
			node:   label,
			want:   true,
		},
		{
			name:   "name exact",
			filter: A11yFilter{Name: "Submit", NameMatch: "exact"},
			node:   button,
			want:   true,
		},
		{
			name:   "name exact fails on substring",
			filter: A11yFilter{Name: "Submit", NameMatch: "exact"},
			node:   label,
			want:   false,
		},
		{
			name:   "name regex",
			filter: A11yFilter{Name: "^Sub.*t$", NameMatch: "regex"},
			node:   button,
			want:   true,
		},
		{
			name:   "name regex no match",
			filter: A11yFilter{Name: "^Cancel$", NameMatch: "regex"},
			node:   button,
			want:   false,
		},
		{
			name:    "bad regex returns error",
			filter:  A11yFilter{Name: "[", NameMatch: "regex"},
			wantErr: true,
		},
		{
			name:    "unknown nameMatch mode",
			filter:  A11yFilter{Name: "foo", NameMatch: "fuzzy"},
			wantErr: true,
		},
		{
			name:   "states AND semantics - all present",
			filter: A11yFilter{States: []string{"enabled", "focusable"}},
			node:   button,
			want:   true,
		},
		{
			name:   "states AND semantics - one missing",
			filter: A11yFilter{States: []string{"enabled", "focusable"}},
			node:   disabledBtn,
			want:   false,
		},
		{
			name: "combined AND all satisfied",
			filter: A11yFilter{
				Role:   "push button",
				Name:   "Submit",
				States: []string{"enabled"},
			},
			node: button,
			want: true,
		},
		{
			name: "combined AND one fails",
			filter: A11yFilter{
				Role:   "push button",
				Name:   "Submit",
				States: []string{"pressed"},
			},
			node: button,
			want: false,
		},
		{
			name:   "empty filter matches everything",
			filter: A11yFilter{},
			node:   button,
			want:   true,
		},
		{
			name:   "case-sensitive substring",
			filter: A11yFilter{Name: "submit"},
			node:   button,
			want:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := buildFilterMatcher(tc.filter)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error from buildFilterMatcher, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			got := m(tc.node)
			if got != tc.want {
				t.Errorf("matcher returned %v, want %v", got, tc.want)
			}
		})
	}
}

func TestStateBitsToNames(t *testing.T) {
	// Set bits: ENABLED (8), VISIBLE (30), FOCUSED (12).
	bits := []uint32{0, 0}
	set := func(i int) {
		bits[i/32] |= 1 << uint(i%32)
	}
	set(8)
	set(30)
	set(12)

	names := stateBitsToNames(bits)

	want := map[string]bool{"enabled": true, "visible": true, "focused": true}
	if len(names) != len(want) {
		t.Fatalf("expected %d names, got %d: %v", len(want), len(names), names)
	}
	for _, n := range names {
		if !want[n] {
			t.Errorf("unexpected state name %q", n)
		}
	}
}

func TestStateBitSet(t *testing.T) {
	bits := []uint32{1 << 1, 1 << 0} // bit 1 in word 0, bit 32 in word 1
	if !stateBitSet(bits, 1) {
		t.Error("expected bit 1 set")
	}
	if !stateBitSet(bits, 32) {
		t.Error("expected bit 32 set")
	}
	if stateBitSet(bits, 0) {
		t.Error("bit 0 should not be set")
	}
	if stateBitSet(bits, 999) {
		t.Error("out-of-range bit should read as false, not panic")
	}
}

func TestParseWireScope(t *testing.T) {
	cases := []struct {
		in      string
		want    A11yScope
		wantErr bool
	}{
		{"", A11yScopeFocused, false},
		{"focused", A11yScopeFocused, false},
		{"FOCUSED", A11yScopeFocused, false},
		{" focused ", A11yScopeFocused, false},
		{"pid", A11yScopePID, false},
		{"Pid", A11yScopePID, false},
		{"all", A11yScopeAll, false},
		{"bogus", "", true},
		{"tree", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseWireScope(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.in)
				}
				if !strings.Contains(err.Error(), "invalid accessibility scope") {
					t.Errorf("error %q must contain the ErrInvalidScope message so the daemon handler can classify it", err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("parseWireScope(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestToWireNodeRecursive(t *testing.T) {
	in := &A11yNode{
		ID:     "a",
		Role:   "frame",
		Bounds: A11yBounds{X: 1, Y: 2, Width: 3, Height: 4},
		Children: []*A11yNode{
			{ID: "b", Role: "button"},
			{
				ID:   "c",
				Role: "panel",
				Children: []*A11yNode{
					{ID: "d", Role: "label"},
				},
			},
		},
	}
	out := toWireNode(in)
	if out == nil {
		t.Fatal("toWireNode returned nil for non-nil input")
	}
	if out.ID != "a" || out.Role != "frame" {
		t.Errorf("top-level fields not copied: %+v", out)
	}
	if out.Bounds.X != 1 || out.Bounds.Y != 2 || out.Bounds.Width != 3 || out.Bounds.Height != 4 {
		t.Errorf("bounds not copied: %+v", out.Bounds)
	}
	if len(out.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(out.Children))
	}
	if out.Children[1].ID != "c" {
		t.Errorf("child order changed: %+v", out.Children)
	}
	if len(out.Children[1].Children) != 1 || out.Children[1].Children[0].ID != "d" {
		t.Errorf("grandchild not recursively converted: %+v", out.Children[1])
	}

	if nilOut := toWireNode(nil); nilOut != nil {
		t.Errorf("toWireNode(nil) should return nil, got %+v", nilOut)
	}
}

func TestToWireNodesFlat(t *testing.T) {
	in := []*A11yNode{
		{ID: "a", Role: "button"},
		nil, // should be skipped, not panic
		{ID: "b", Role: "label"},
	}
	out := toWireNodes(in)
	if len(out) != 2 {
		t.Fatalf("expected 2 nodes (nil skipped), got %d", len(out))
	}
	if out[0].ID != "a" || out[1].ID != "b" {
		t.Errorf("unexpected IDs: %+v", out)
	}
}
