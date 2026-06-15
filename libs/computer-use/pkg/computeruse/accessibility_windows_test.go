//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-ole/go-ole"
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

// Depth, find-limit, and wire-scope semantics moved to
// accessibility_common.go and are pinned by the Linux test suite
// (accessibility_test.go), which runs in CI.

func TestInvokeWindowsElementSetValueRedirectsToValueEndpoint(t *testing.T) {
	// The set_value branch must short-circuit before any COM call (nil
	// element) since the invoke request cannot carry a value.
	for _, action := range []string{"set_value", " SET_VALUE "} {
		err := invokeWindowsElement("handle", nil, action)
		if !errors.Is(err, ErrActionNotSupported) {
			t.Fatalf("invokeWindowsElement(%q) = %v, want ErrActionNotSupported", action, err)
		}
		if !strings.HasPrefix(err.Error(), ErrActionNotSupported.Error()+":") {
			t.Errorf("error %q must keep the sentinel prefix the daemon matches on", err.Error())
		}
		if !strings.Contains(err.Error(), "value endpoint") {
			t.Errorf("error %q must point the caller at the value endpoint", err.Error())
		}
	}
}

func TestClassifyWindowsActionError(t *testing.T) {
	dead := ole.NewError(uiaErrElementNotAvailable)
	elementMap.put("stale-handle", nil)

	err := classifyWindowsActionError("stale-handle", "SetFocus", dead)
	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("element-not-available error = %v, want ErrNodeNotFound", err)
	}
	if !strings.HasPrefix(err.Error(), ErrNodeNotFound.Error()+":") {
		t.Errorf("error %q must keep the sentinel prefix the daemon matches on", err.Error())
	}
	if _, ok := elementMap.get("stale-handle"); ok {
		t.Error("dead handle must be evicted from the element cache")
	}

	// HRESULTs that unambiguously mean "the element refuses this action"
	// map to the 400 sentinel.
	for _, code := range []uintptr{
		uiaErrElementNotEnabled,
		uiaErrNotSupported,
		uiaErrInvalidOperation,
		hresultNoInterface,
		hresultNotImplemented,
	} {
		err := classifyWindowsActionError("h", "Invoke", ole.NewError(code))
		if !errors.Is(err, ErrActionNotSupported) {
			t.Errorf("refusal HRESULT %#x = %v, want ErrActionNotSupported", code, err)
		}
		if !strings.HasPrefix(err.Error(), ErrActionNotSupported.Error()+":") {
			t.Errorf("error %q must keep the sentinel prefix the daemon matches on", err.Error())
		}
	}

	// Everything else — transient COM faults and non-COM errors — passes
	// through untranslated so the daemon reports a retryable internal
	// error instead of a permanent refusal (Linux classifyDbusError
	// contract: unknown errors are returned as-is).
	transient := ole.NewError(0x80004005) // E_FAIL
	err = classifyWindowsActionError("h", "Invoke", transient)
	if errors.Is(err, ErrActionNotSupported) || errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("generic COM error = %v, must not map to a sentinel", err)
	}
	if !errors.Is(err, transient) {
		t.Errorf("generic COM error %v must wrap the original error", err)
	}

	plain := errors.New("rpc fault")
	err = classifyWindowsActionError("h", "SetValue", plain)
	if errors.Is(err, ErrActionNotSupported) || errors.Is(err, ErrNodeNotFound) {
		t.Fatalf("non-COM error = %v, must not map to a sentinel", err)
	}
	if !errors.Is(err, plain) {
		t.Errorf("non-COM error %v must wrap the original error", err)
	}
}

func TestAccessibilityValidationRunsBeforeSTA(t *testing.T) {
	// Malformed requests must be rejected by pure validation before any
	// COM/STA work: the 400 must win even on hosts where CoInitializeEx
	// fails, and must never queue behind in-flight UIA walks.
	staStarted := staCh != nil
	c := &ComputerUse{}

	if _, err := c.GetAccessibilityTree(&computeruse.GetAccessibilityTreeRequest{Scope: "bogus"}); !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("GetAccessibilityTree(bogus scope) = %v, want ErrInvalidScope", err)
	}
	if _, err := c.FindAccessibilityNodes(&computeruse.FindAccessibilityNodesRequest{Scope: "bogus"}); !errors.Is(err, ErrInvalidScope) {
		t.Fatalf("FindAccessibilityNodes(bogus scope) = %v, want ErrInvalidScope", err)
	}
	if _, err := c.FindAccessibilityNodes(&computeruse.FindAccessibilityNodesRequest{Name: "(", NameMatch: "regex"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("FindAccessibilityNodes(invalid regex) = %v, want ErrInvalidRequest", err)
	}

	if !staStarted && staCh != nil {
		t.Error("pure request validation must not start the STA thread")
	}
}

func TestRunOnSTAReentrantCallPanics(t *testing.T) {
	// A nested runOnSTA call from the STA thread can never be serviced —
	// the single consumer goroutine is busy executing the outer closure —
	// so it must panic loudly instead of hanging every a11y request. If
	// the guard regresses this test deadlocks and fails via the go test
	// timeout rather than asserting.
	var recovered any
	err := runOnSTA(func() error {
		defer func() { recovered = recover() }()
		_ = runOnSTA(func() error { return nil })
		return nil
	})
	if err != nil {
		t.Fatalf("outer runOnSTA returned %v, want nil", err)
	}
	if recovered == nil {
		t.Fatal("re-entrant runOnSTA on the STA thread must panic, not deadlock")
	}
}

func TestElementCachePutAmortizesGC(t *testing.T) {
	c := &elementCache{elts: map[string]*cachedElement{
		"expired": {expiry: time.Now().Add(-time.Minute)},
	}}

	c.nextGC = time.Now().Add(time.Hour) // sweep not due yet
	c.put("fresh", nil)
	if _, ok := c.elts["expired"]; !ok {
		t.Fatal("expired entry must survive puts until the GC interval elapses")
	}

	c.nextGC = time.Time{} // sweep due
	c.put("fresh2", nil)
	if _, ok := c.elts["expired"]; ok {
		t.Fatal("due sweep must drop expired entries")
	}
	if len(c.elts) != 2 {
		t.Fatalf("cache has %d entries, want the 2 fresh ones", len(c.elts))
	}
}

func TestElementCachePutEvictsOldestOnOverflow(t *testing.T) {
	c := &elementCache{
		elts:   make(map[string]*cachedElement, elementCacheMaxEntries),
		nextGC: time.Now().Add(time.Hour),
	}
	base := time.Now().Add(time.Hour) // nothing expired; only the cap acts
	for i := 0; i < elementCacheMaxEntries; i++ {
		c.elts[fmt.Sprintf("h%05d", i)] = &cachedElement{expiry: base.Add(time.Duration(i) * time.Second)}
	}

	c.put("newest", nil)

	if len(c.elts) > elementCacheMaxEntries {
		t.Fatalf("cache size %d exceeds cap %d", len(c.elts), elementCacheMaxEntries)
	}
	if _, ok := c.elts["newest"]; !ok {
		t.Fatal("triggering insert must be present")
	}
	batch := elementCacheMaxEntries / 10
	if _, ok := c.elts[fmt.Sprintf("h%05d", batch-1)]; ok {
		t.Errorf("entry inside the oldest-expiry batch (%d) must be evicted", batch-1)
	}
	if _, ok := c.elts[fmt.Sprintf("h%05d", batch)]; !ok {
		t.Errorf("entry just outside the oldest-expiry batch (%d) must survive", batch)
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

	if _, err := buildWindowsFilterMatcher(&computeruse.FindAccessibilityNodesRequest{Name: "(", NameMatch: "regex"}); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("invalid regex error = %v, want ErrInvalidRequest", err)
	}
	if _, err := buildWindowsFilterMatcher(&computeruse.FindAccessibilityNodesRequest{NameMatch: "fuzzy"}); !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("invalid nameMatch error = %v, want ErrInvalidRequest", err)
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
