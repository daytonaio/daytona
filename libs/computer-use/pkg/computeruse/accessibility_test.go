// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"strings"
	"testing"

	wire "github.com/daytonaio/daemon/pkg/toolbox/computeruse"
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
			} else if !errors.Is(err, ErrInvalidRequest) {
				t.Errorf("parseNodeID(%q) error = %v, want ErrInvalidRequest", tc.id, err)
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
			name:    "unknown nameMatch mode without name",
			filter:  A11yFilter{NameMatch: "fuzzy"},
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
				if !errors.Is(err, ErrInvalidRequest) {
					t.Fatalf("error = %v, want ErrInvalidRequest", err)
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

func TestWaitAccessibilityExistsPollsUntilMatch(t *testing.T) {
	calls := 0
	c := &ComputerUse{
		findA11yNodes: func(scope A11yScope, pid int, filter A11yFilter, limit int) ([]*A11yNode, bool, error) {
			calls++
			if scope != A11yScopeAll || filter.Role != "push button" {
				t.Fatalf("unexpected find args: scope=%q filter=%+v", scope, filter)
			}
			if calls == 1 {
				return nil, false, nil
			}
			return []*A11yNode{{
				ID:     ":1.42:/org/a11y/atspi/accessible/12",
				Role:   "push button",
				Name:   "OK",
				States: []string{"enabled", "visible"},
			}}, false, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "exists",
		Query:          &wire.FindAccessibilityNodesRequest{Scope: "all", Role: "push button"},
		TimeoutMs:      100,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("find calls = %d, want 2", calls)
	}
	if !resp.Matched || resp.TimedOut || len(resp.Matches) != 1 || resp.Matches[0].Name != "OK" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestWaitAccessibilityTimeout(t *testing.T) {
	c := &ComputerUse{
		findA11yNodes: func(A11yScope, int, A11yFilter, int) ([]*A11yNode, bool, error) {
			return nil, false, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "exists",
		Query:          &wire.FindAccessibilityNodesRequest{Scope: "all", Name: "Ready"},
		TimeoutMs:      1,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if resp.Matched || !resp.TimedOut {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestWaitAccessibilityGoneMatchesEmptyResult(t *testing.T) {
	calls := 0
	c := &ComputerUse{
		findA11yNodes: func(A11yScope, int, A11yFilter, int) ([]*A11yNode, bool, error) {
			calls++
			if calls == 1 {
				return []*A11yNode{{ID: ":1.42:/node", Name: "Toast"}}, false, nil
			}
			return nil, false, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "gone",
		Query:          &wire.FindAccessibilityNodesRequest{Scope: "all", Name: "Toast"},
		TimeoutMs:      100,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if calls != 2 || !resp.Matched || len(resp.Matches) != 0 {
		t.Fatalf("calls=%d response=%+v", calls, resp)
	}
}

func TestWaitAccessibilityGoneDoesNotMatchTruncatedEmptyResult(t *testing.T) {
	c := &ComputerUse{
		findA11yNodes: func(A11yScope, int, A11yFilter, int) ([]*A11yNode, bool, error) {
			return nil, true, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "gone",
		Query:          &wire.FindAccessibilityNodesRequest{Scope: "all", Name: "Toast"},
		TimeoutMs:      1,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if resp.Matched || !resp.TimedOut {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestWaitAccessibilityValidation(t *testing.T) {
	c := &ComputerUse{}
	cases := []struct {
		name string
		req  *wire.AccessibilityWaitRequest
	}{
		{
			name: "invalid condition",
			req:  &wire.AccessibilityWaitRequest{Condition: "visible"},
		},
		{
			name: "exists missing query",
			req:  &wire.AccessibilityWaitRequest{Condition: "exists"},
		},
		{
			name: "gone missing query",
			req:  &wire.AccessibilityWaitRequest{Condition: "gone"},
		},
		{
			name: "state missing query and id",
			req:  &wire.AccessibilityWaitRequest{Condition: "state", States: []string{"focused"}},
		},
		{
			name: "state has query and id",
			req: &wire.AccessibilityWaitRequest{
				Condition: "state",
				Query:     &wire.FindAccessibilityNodesRequest{Scope: "all"},
				ID:        ":1.42:/org/a11y/atspi/accessible/12",
				States:    []string{"focused"},
			},
		},
		{
			name: "state missing states",
			req:  &wire.AccessibilityWaitRequest{Condition: "state", ID: ":1.42:/org/a11y/atspi/accessible/12"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := c.WaitAccessibility(tc.req)
			if !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("error = %v, want ErrInvalidRequest", err)
			}
		})
	}
}

func TestWaitAccessibilityStateByQuery(t *testing.T) {
	calls := 0
	c := &ComputerUse{
		findA11yNodes: func(A11yScope, int, A11yFilter, int) ([]*A11yNode, bool, error) {
			calls++
			states := []string{"visible"}
			if calls > 1 {
				states = []string{"visible", "focused", "enabled"}
			}
			return []*A11yNode{{ID: ":1.42:/node", Name: "Search", States: states}}, false, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "state",
		Query:          &wire.FindAccessibilityNodesRequest{Scope: "all", Name: "Search"},
		States:         []string{"focused", "enabled"},
		TimeoutMs:      100,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if calls != 2 || !resp.Matched || len(resp.Matches) != 1 || resp.Matches[0].Name != "Search" {
		t.Fatalf("calls=%d response=%+v", calls, resp)
	}
}

func TestWaitAccessibilityStateByID(t *testing.T) {
	calls := 0
	c := &ComputerUse{
		fetchA11yNode: func(id string) (*A11yNode, error) {
			calls++
			if id != ":1.42:/node" {
				t.Fatalf("id = %q, want :1.42:/node", id)
			}
			states := []string{"visible"}
			if calls > 1 {
				states = []string{"visible", "focused"}
			}
			return &A11yNode{ID: id, Name: "Search", States: states}, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "state",
		ID:             ":1.42:/node",
		States:         []string{"focused"},
		TimeoutMs:      100,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if calls != 2 || !resp.Matched || len(resp.Matches) != 1 || resp.Matches[0].ID != ":1.42:/node" {
		t.Fatalf("calls=%d response=%+v", calls, resp)
	}
}

func TestWaitAccessibilityNotStateRequiresNoForbiddenStates(t *testing.T) {
	calls := 0
	c := &ComputerUse{
		fetchA11yNode: func(id string) (*A11yNode, error) {
			calls++
			states := []string{"visible", "busy"}
			if calls > 1 {
				states = []string{"visible"}
			}
			return &A11yNode{ID: id, Name: "Save", States: states}, nil
		},
	}

	resp, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition:      "not_state",
		ID:             ":1.42:/node",
		States:         []string{"busy", "focused"},
		TimeoutMs:      100,
		PollIntervalMs: 1,
	})
	if err != nil {
		t.Fatalf("WaitAccessibility returned error: %v", err)
	}
	if calls != 2 || !resp.Matched || len(resp.Matches) != 1 || resp.Matches[0].Name != "Save" {
		t.Fatalf("calls=%d response=%+v", calls, resp)
	}
}

func TestWaitAccessibilityBadIDError(t *testing.T) {
	c := &ComputerUse{}
	_, err := c.WaitAccessibility(&wire.AccessibilityWaitRequest{
		Condition: "state",
		ID:        "not-a-node-id",
		States:    []string{"focused"},
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error = %v, want ErrInvalidRequest", err)
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
