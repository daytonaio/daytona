// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"context"
	"errors"
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

func TestGetActionNamesUsesCanonicalNames(t *testing.T) {
	obj := &fakeBusObject{
		properties: map[string]dbus.Variant{
			ifaceAction + ".NActions": dbus.MakeVariant(int32(2)),
		},
		call: func(method string, args ...interface{}) *dbus.Call {
			if method == ifaceAction+".GetActions" {
				t.Fatal("GetActions returns localized labels; canonical action names must come from GetName(index)")
			}
			if method != ifaceAction+".GetName" {
				return failedCall(errors.New("unexpected method: " + method))
			}
			switch args[0].(int32) {
			case 0:
				return completedCall("show-menu")
			case 1:
				return completedCall("click")
			default:
				return failedCall(errors.New("unexpected action index"))
			}
		},
	}

	got, err := getActionNames(obj)
	if err != nil {
		t.Fatalf("getActionNames() error = %v", err)
	}
	want := []string{"show-menu", "click"}
	if len(got) != len(want) {
		t.Fatalf("actions = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("actions = %v, want %v", got, want)
		}
	}
}

func TestGetActionNamesPropagatesLookupErrors(t *testing.T) {
	obj := &fakeBusObject{
		properties: map[string]dbus.Variant{
			ifaceAction + ".NActions": dbus.MakeVariant(int32(1)),
		},
		call: func(method string, args ...interface{}) *dbus.Call {
			if method != ifaceAction+".GetName" {
				return failedCall(errors.New("unexpected method: " + method))
			}
			return failedCall(dbus.NewError(
				"org.freedesktop.DBus.Error.UnknownObject",
				[]interface{}{"node disappeared"},
			))
		},
	}

	_, err := getActionNames(obj)
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("getActionNames() error = %v, want ErrNodeNotFound", err)
	}
}

func TestGetActionNamesPropagatesNActionsErrors(t *testing.T) {
	obj := &fakeBusObject{
		getProperty: func(prop string) (dbus.Variant, error) {
			if prop != ifaceAction+".NActions" {
				return dbus.Variant{}, errors.New("unexpected property: " + prop)
			}
			return dbus.Variant{}, dbus.NewError(
				"org.freedesktop.DBus.Error.NameHasNoOwner",
				[]interface{}{"provider exited"},
			)
		},
	}

	_, err := getActionNames(obj)
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("getActionNames() error = %v, want ErrNodeNotFound", err)
	}
}

func TestActionIndexUsesCanonicalNames(t *testing.T) {
	actions := []string{"show-menu", "click"}

	if got := actionIndex(actions, "show-menu"); got != 0 {
		t.Fatalf("canonical show-menu index = %d, want 0", got)
	}
	if got := actionIndex(actions, "SHOW-MENU"); got != 0 {
		t.Fatalf("case-insensitive canonical show-menu index = %d, want 0", got)
	}
	if got := actionIndex(actions, "Show menu"); got != -1 {
		t.Fatalf("localized Show menu index = %d, want -1", got)
	}
	if got := actionIndex(actions, ""); got != 0 {
		t.Fatalf("empty action index = %d, want default index 0", got)
	}
	if got := actionIndex(nil, "click"); got != -1 {
		t.Fatalf("missing actions index = %d, want -1", got)
	}
}

func TestGetConnectionUnixProcessID(t *testing.T) {
	obj := &fakeBusObject{
		call: func(method string, args ...interface{}) *dbus.Call {
			if method != "org.freedesktop.DBus.GetConnectionUnixProcessID" {
				return failedCall(errors.New("unexpected method: " + method))
			}
			if args[0] != ":1.42" {
				return failedCall(errors.New("unexpected sender"))
			}
			return completedCall(uint32(4242))
		},
	}

	got, err := getConnectionUnixProcessID(obj, ":1.42")
	if err != nil {
		t.Fatalf("getConnectionUnixProcessID() error = %v", err)
	}
	if got != 4242 {
		t.Fatalf("pid = %d, want 4242", got)
	}
}

func TestDisappearingNodeErrorIsNarrow(t *testing.T) {
	if !isDisappearingNodeError(ErrNodeNotFound) {
		t.Fatal("ErrNodeNotFound should be treated as a disappearing node")
	}
	if isDisappearingNodeError(ErrA11yUnavailable) {
		t.Fatal("ErrA11yUnavailable should propagate, not be swallowed as a disappearing node")
	}
	if isDisappearingNodeError(errors.New("random D-Bus failure")) {
		t.Fatal("unknown errors should propagate, not be swallowed as disappearing nodes")
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
		return
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

type fakeBusObject struct {
	properties  map[string]dbus.Variant
	getProperty func(prop string) (dbus.Variant, error)
	call        func(method string, args ...interface{}) *dbus.Call
}

func (f *fakeBusObject) Call(method string, _ dbus.Flags, args ...interface{}) *dbus.Call {
	if f.call == nil {
		return failedCall(errors.New("unexpected method: " + method))
	}
	return f.call(method, args...)
}

func (f *fakeBusObject) CallWithContext(_ context.Context, method string, flags dbus.Flags, args ...interface{}) *dbus.Call {
	return f.Call(method, flags, args...)
}

func (f *fakeBusObject) Go(method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	call := f.Call(method, flags, args...)
	if ch != nil {
		ch <- call
	}
	return call
}

func (f *fakeBusObject) GoWithContext(_ context.Context, method string, flags dbus.Flags, ch chan *dbus.Call, args ...interface{}) *dbus.Call {
	return f.Go(method, flags, ch, args...)
}

func (f *fakeBusObject) AddMatchSignal(string, string, ...dbus.MatchOption) *dbus.Call {
	return completedCall()
}

func (f *fakeBusObject) RemoveMatchSignal(string, string, ...dbus.MatchOption) *dbus.Call {
	return completedCall()
}

func (f *fakeBusObject) GetProperty(prop string) (dbus.Variant, error) {
	if f.getProperty != nil {
		return f.getProperty(prop)
	}
	if f.properties == nil {
		return dbus.Variant{}, errors.New("unexpected property: " + prop)
	}
	v, ok := f.properties[prop]
	if !ok {
		return dbus.Variant{}, errors.New("unexpected property: " + prop)
	}
	return v, nil
}

func (f *fakeBusObject) StoreProperty(prop string, value interface{}) error {
	v, err := f.GetProperty(prop)
	if err != nil {
		return err
	}
	return dbus.Store([]interface{}{v.Value()}, value)
}

func (f *fakeBusObject) SetProperty(string, interface{}) error { return nil }
func (f *fakeBusObject) Destination() string                   { return "fake.destination" }
func (f *fakeBusObject) Path() dbus.ObjectPath                 { return "/fake/path" }

func completedCall(body ...interface{}) *dbus.Call {
	return &dbus.Call{Body: body, Done: make(chan *dbus.Call, 1)}
}

func failedCall(err error) *dbus.Call {
	return &dbus.Call{Err: err, Done: make(chan *dbus.Call, 1)}
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
