//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	uia "github.com/uandersonricardo/uiautomation"
)

func TestWindowsRoleNameCanonicalVocabulary(t *testing.T) {
	cases := map[uia.ControlTypeId]string{
		uia.ButtonControlTypeId:      "push button",
		uia.SplitButtonControlTypeId: "push button",
		uia.EditControlTypeId:        "entry",
		uia.TextControlTypeId:        "label",
		uia.HyperlinkControlTypeId:   "link",
		uia.WindowControlTypeId:      "frame",
		uia.PaneControlTypeId:        "panel",
		uia.GroupControlTypeId:       "panel",
		uia.TabControlTypeId:         "page tab list",
		uia.TabItemControlTypeId:     "page tab",
		uia.SpinnerControlTypeId:     "spin button",
		uia.DataGridControlTypeId:    "table",
		uia.HeaderItemControlTypeId:  "column header",
		uia.CustomControlTypeId:      "unknown",
	}
	for controlType, want := range cases {
		if got := windowsRoleName(controlType); got != want {
			t.Errorf("windowsRoleName(%d) = %q, want %q", controlType, got, want)
		}
	}
}

func TestWindowsRoleNameCoversAllStandardControlTypes(t *testing.T) {
	for controlType := uia.ButtonControlTypeId; controlType <= uia.AppBarControlTypeId; controlType++ {
		role := windowsRoleName(controlType)
		if strings.HasPrefix(role, windowsRoleRawPrefix) {
			t.Errorf("standard control type %d has no canonical role", controlType)
		}
		if role != strings.ToLower(role) {
			t.Errorf("role %q for control type %d is not lowercase", role, controlType)
		}
	}
}

func TestWindowsRoleNameUnknownControlTypeFallsBack(t *testing.T) {
	if got, want := windowsRoleName(uia.ControlTypeId(99999)), "control_type_99999"; got != want {
		t.Errorf("windowsRoleName(99999) = %q, want %q", got, want)
	}
}

func TestWindowsControlTypesForRoleRoundTrip(t *testing.T) {
	for controlType, role := range windowsControlTypeRoles {
		ids, ok := windowsControlTypesForRole(role)
		if !ok {
			t.Errorf("role %q is emitted but not reverse-mappable", role)
			continue
		}
		found := false
		for _, id := range ids {
			if id == controlType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("reverse mapping for %q misses control type %d", role, controlType)
		}
	}
}

func TestWindowsControlTypesForRole(t *testing.T) {
	tests := []struct {
		role string
		want []uia.ControlTypeId
		ok   bool
	}{
		{"push button", []uia.ControlTypeId{uia.ButtonControlTypeId, uia.SplitButtonControlTypeId}, true},
		{" Push Button ", []uia.ControlTypeId{uia.ButtonControlTypeId, uia.SplitButtonControlTypeId}, true},
		{"panel", []uia.ControlTypeId{uia.GroupControlTypeId, uia.PaneControlTypeId, uia.SemanticZoomControlTypeId}, true},
		{"table", []uia.ControlTypeId{uia.DataGridControlTypeId, uia.TableControlTypeId}, true},
		{"unknown", []uia.ControlTypeId{uia.CustomControlTypeId, uia.ThumbControlTypeId}, true},
		{"control_type_50000", []uia.ControlTypeId{uia.ButtonControlTypeId}, true},
		{"control_type_99999", []uia.ControlTypeId{uia.ControlTypeId(99999)}, true},
		{"control_type_12x", nil, false},
		{"control_type_", nil, false},
		{"schaltfläche", nil, false},
		{"", nil, false},
	}
	for _, tc := range tests {
		ids, ok := windowsControlTypesForRole(tc.role)
		if ok != tc.ok {
			t.Errorf("windowsControlTypesForRole(%q) ok = %v, want %v", tc.role, ok, tc.ok)
			continue
		}
		if tc.ok && !reflect.DeepEqual(ids, tc.want) {
			t.Errorf("windowsControlTypesForRole(%q) = %v, want %v", tc.role, ids, tc.want)
		}
	}
}

func TestPlanWindowsFindConditions(t *testing.T) {
	t.Run("canonical role narrows natively", func(t *testing.T) {
		plan := planWindowsFindConditions("push button", "", "")
		if plan.impossible {
			t.Fatal("push button must not be impossible")
		}
		want := []uia.ControlTypeId{uia.ButtonControlTypeId, uia.SplitButtonControlTypeId}
		if !reflect.DeepEqual(plan.roleControlTypes, want) {
			t.Errorf("roleControlTypes = %v, want %v", plan.roleControlTypes, want)
		}
		if plan.nameNative {
			t.Error("empty name must not be native")
		}
	})

	t.Run("vocabulary is closed", func(t *testing.T) {
		plan := planWindowsFindConditions("zorblax", "", "")
		if !plan.impossible {
			t.Error("unmappable role must be impossible")
		}
	})

	t.Run("unknown role cannot narrow", func(t *testing.T) {
		plan := planWindowsFindConditions("unknown", "", "")
		if plan.impossible {
			t.Error("unknown role is emitted and must not be impossible")
		}
		if len(plan.roleControlTypes) != 0 {
			t.Errorf("unknown role must not narrow natively, got %v", plan.roleControlTypes)
		}
	})

	t.Run("raw control type role narrows natively", func(t *testing.T) {
		plan := planWindowsFindConditions("control_type_50099", "", "")
		if plan.impossible {
			t.Fatal("raw control type role must not be impossible")
		}
		want := []uia.ControlTypeId{uia.ControlTypeId(50099)}
		if !reflect.DeepEqual(plan.roleControlTypes, want) {
			t.Errorf("roleControlTypes = %v, want %v", plan.roleControlTypes, want)
		}
	})

	t.Run("substring name is native with widening flags", func(t *testing.T) {
		for _, mode := range []string{"", "substring"} {
			plan := planWindowsFindConditions("", "OK", mode)
			if !plan.nameNative {
				t.Fatalf("nameMatch %q must be native", mode)
			}
			want := uia.PropertyConditionFlagsIgnoreCase | uia.PropertyConditionFlagsMatchSubstring
			if plan.nameFlags != want {
				t.Errorf("nameMatch %q flags = %v, want %v", mode, plan.nameFlags, want)
			}
		}
	})

	t.Run("exact name is native without substring flag", func(t *testing.T) {
		plan := planWindowsFindConditions("", "OK", "exact")
		if !plan.nameNative {
			t.Fatal("exact name must be native")
		}
		if plan.nameFlags != uia.PropertyConditionFlagsIgnoreCase {
			t.Errorf("flags = %v, want IgnoreCase", plan.nameFlags)
		}
	})

	t.Run("regex name stays go-side", func(t *testing.T) {
		plan := planWindowsFindConditions("frame", "^Save", "regex")
		if plan.nameNative {
			t.Error("regex name must not be native")
		}
		if len(plan.roleControlTypes) == 0 {
			t.Error("role must still narrow natively alongside a regex name")
		}
	})

	t.Run("interior NUL name never narrows natively", func(t *testing.T) {
		for _, mode := range []string{"", "substring", "exact"} {
			plan := planWindowsFindConditions("", "a\x00b", mode)
			if plan.nameNative {
				t.Errorf("nameMatch %q with interior NUL must stay go-side", mode)
			}
			if plan.impossible {
				t.Errorf("nameMatch %q with interior NUL must not be impossible", mode)
			}
		}
	})

	t.Run("interior NUL name keeps role narrowing", func(t *testing.T) {
		plan := planWindowsFindConditions("push button", "a\x00b", "substring")
		if plan.nameNative {
			t.Error("interior NUL name must stay go-side")
		}
		if len(plan.roleControlTypes) == 0 {
			t.Error("role must still narrow natively alongside a NUL name")
		}
	})
}

func TestWindowsDepthSemantics(t *testing.T) {
	if windowsDepthAllowsDescent(0) {
		t.Error("maxDepth 0 must visit the root only")
	}
	for _, depth := range []int{-1, -5, 1, 3} {
		if !windowsDepthAllowsDescent(depth) {
			t.Errorf("maxDepth %d must allow descent", depth)
		}
	}
	next := map[int]int{-5: -5, -1: -1, 1: 0, 3: 2}
	for depth, want := range next {
		if got := windowsNextDepth(depth); got != want {
			t.Errorf("windowsNextDepth(%d) = %d, want %d", depth, got, want)
		}
	}
}

func TestNormalizeWindowsFindLimit(t *testing.T) {
	tests := map[int]int{
		-1:                          windowsFindDefaultLimit,
		0:                           windowsFindDefaultLimit,
		1:                           1,
		windowsFindDefaultLimit:     windowsFindDefaultLimit,
		windowsFindCeilingLimit:     windowsFindCeilingLimit,
		windowsFindCeilingLimit + 1: windowsFindCeilingLimit,
	}
	for in, want := range tests {
		if got := normalizeWindowsFindLimit(in); got != want {
			t.Errorf("normalizeWindowsFindLimit(%d) = %d, want %d", in, got, want)
		}
	}
}

func TestParseWindowsWireScope(t *testing.T) {
	tests := []struct {
		in   string
		want windowsA11yScope
	}{
		{"", windowsA11yScopeFocused},
		{"focused", windowsA11yScopeFocused},
		{" FOCUSED ", windowsA11yScopeFocused},
		{"all", windowsA11yScopeAll},
		{" All", windowsA11yScopeAll},
		{"pid", windowsA11yScopePID},
		{"PID", windowsA11yScopePID},
	}
	for _, tc := range tests {
		got, err := parseWindowsWireScope(tc.in)
		if err != nil {
			t.Errorf("parseWindowsWireScope(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseWindowsWireScope(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}

	for _, in := range []string{"windows", "focus", "0"} {
		if _, err := parseWindowsWireScope(in); !errors.Is(err, errA11yInvalidScope) {
			t.Errorf("parseWindowsWireScope(%q) = %v, want errA11yInvalidScope", in, err)
		}
	}
}

func TestBuildWindowsFilterMatcher(t *testing.T) {
	node := &computeruse.AccessibilityNode{
		Role:   "push button",
		Name:   "Save As",
		States: []string{"enabled", "visible"},
	}

	match := func(t *testing.T, req *computeruse.FindAccessibilityNodesRequest) bool {
		t.Helper()
		matcher, err := buildWindowsFilterMatcher(req)
		if err != nil {
			t.Fatalf("buildWindowsFilterMatcher: %v", err)
		}
		return matcher(node)
	}

	if !match(t, &computeruse.FindAccessibilityNodesRequest{}) {
		t.Error("empty filter must match")
	}
	if !match(t, &computeruse.FindAccessibilityNodesRequest{Role: "PUSH Button"}) {
		t.Error("role must compare case-insensitively")
	}
	if match(t, &computeruse.FindAccessibilityNodesRequest{Role: "frame"}) {
		t.Error("mismatched role must not match")
	}
	if !match(t, &computeruse.FindAccessibilityNodesRequest{Name: "Save"}) {
		t.Error("substring name must match by default")
	}
	if match(t, &computeruse.FindAccessibilityNodesRequest{Name: "save"}) {
		t.Error("substring name must stay case-sensitive like Linux")
	}
	if !match(t, &computeruse.FindAccessibilityNodesRequest{Name: "Save As", NameMatch: "exact"}) {
		t.Error("exact name must match")
	}
	if match(t, &computeruse.FindAccessibilityNodesRequest{Name: "Save", NameMatch: "exact"}) {
		t.Error("partial exact name must not match")
	}
	if !match(t, &computeruse.FindAccessibilityNodesRequest{Name: "^Save", NameMatch: "regex"}) {
		t.Error("regex name must match")
	}
	if !match(t, &computeruse.FindAccessibilityNodesRequest{States: []string{"enabled"}}) {
		t.Error("state subset must match")
	}
	if match(t, &computeruse.FindAccessibilityNodesRequest{States: []string{"enabled", "focused"}}) {
		t.Error("missing state must not match")
	}

	if _, err := buildWindowsFilterMatcher(&computeruse.FindAccessibilityNodesRequest{Name: "(", NameMatch: "regex"}); !errors.Is(err, errA11yInvalidRequest) {
		t.Errorf("invalid regex error = %v, want errA11yInvalidRequest", err)
	}
	if _, err := buildWindowsFilterMatcher(&computeruse.FindAccessibilityNodesRequest{NameMatch: "fuzzy"}); !errors.Is(err, errA11yInvalidRequest) {
		t.Errorf("invalid nameMatch error = %v, want errA11yInvalidRequest", err)
	}
}

func TestWindowsRoleControlTypesSorted(t *testing.T) {
	for role, ids := range windowsRoleControlTypes {
		for i := 1; i < len(ids); i++ {
			if ids[i-1] >= ids[i] {
				t.Errorf("control types for %q not strictly sorted: %v", role, ids)
				break
			}
		}
	}
}
