//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-ole/go-ole"
	uia "github.com/uandersonricardo/uiautomation"
)

// Sentinel errors mirroring the Linux AT-SPI implementation. The daemon maps
// these by exact leading text across the plugin RPC boundary.
var (
	errA11yUnavailable        = fmt.Errorf("accessibility bus not reachable")
	errA11yNoAccessibleRoot   = fmt.Errorf("no accessible root for focused window")
	errA11yNodeNotFound       = fmt.Errorf("accessibility node not found")
	errA11yActionNotSupported = fmt.Errorf("action not supported by node")
	errA11yInvalidScope       = fmt.Errorf("invalid accessibility scope")
	errA11yInvalidRequest     = fmt.Errorf("invalid accessibility request")
)

// COM/UI Automation requires calls from a Single-Threaded Apartment. Go
// goroutines move between OS threads, so all UIA work is serialized through one
// locked thread initialized as an STA.
var (
	staOnce    sync.Once
	staCh      chan func()
	staInitErr error
)

func startSTA() {
	staOnce.Do(func() {
		staCh = make(chan func(), 16)
		go func() {
			runtime.LockOSThread()
			staInitErr = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
			if staInitErr == nil {
				defer ole.CoUninitialize()
			}
			for fn := range staCh {
				fn()
			}
		}()
	})
}

func runOnSTA(fn func() error) error {
	startSTA()
	done := make(chan error, 1)
	staCh <- func() { done <- fn() }
	return <-done
}

// The daemon treats AccessibilityNode.ID as an opaque handle. UIA runtime IDs
// are not stable enough for later HTTP calls, so Windows keeps a short-lived
// session map from generated UUID handles to COM element pointers.
type elementCache struct {
	mu   sync.Mutex
	elts map[string]*cachedElement
}

type cachedElement struct {
	elt    *uia.Element
	expiry time.Time
}

const elementCacheTTL = 5 * time.Minute

var elementMap = &elementCache{elts: map[string]*cachedElement{}}

func (c *elementCache) put(handle string, elt *uia.Element) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.elts[handle] = &cachedElement{elt: elt, expiry: time.Now().Add(elementCacheTTL)}
	c.gcLocked()
}

func (c *elementCache) get(handle string) (*uia.Element, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ce, ok := c.elts[handle]
	if !ok || time.Now().After(ce.expiry) {
		if ok {
			delete(c.elts, handle)
			releaseElement(ce.elt)
		}
		return nil, false
	}
	ce.expiry = time.Now().Add(elementCacheTTL)
	return ce.elt, true
}

func (c *elementCache) gcLocked() {
	now := time.Now()
	for k, v := range c.elts {
		if now.After(v.expiry) {
			delete(c.elts, k)
			releaseElement(v.elt)
		}
	}
}

type windowsA11yScope string

const (
	windowsA11yScopeFocused windowsA11yScope = "focused"
	windowsA11yScopePID     windowsA11yScope = "pid"
	windowsA11yScopeAll     windowsA11yScope = "all"

	windowsA11yWalkBudget   = 20000
	windowsFindDefaultLimit = 500
	windowsFindCeilingLimit = 5000
	actionInvoke            = "invoke"
	actionSelect            = "select"
	actionSetValue          = "set_value"
)

func (c *ComputerUse) GetAccessibilityTree(req *computeruse.GetAccessibilityTreeRequest) (*computeruse.AccessibilityTreeResponse, error) {
	if req == nil {
		req = &computeruse.GetAccessibilityTreeRequest{}
	}

	var resp *computeruse.AccessibilityTreeResponse
	err := runOnSTA(func() error {
		automation, err := newWindowsAutomation()
		if err != nil {
			return err
		}
		defer automation.Release()

		scope, err := parseWindowsWireScope(req.Scope)
		if err != nil {
			return err
		}
		root, err := resolveWindowsScopeRoot(automation, scope, req.PID)
		if err != nil {
			return err
		}

		walker, err := automation.ControlViewWalker()
		if err != nil {
			return fmt.Errorf("%w: ControlViewWalker: %v", errA11yUnavailable, err)
		}
		defer walker.Release()

		// MaxDepth follows the Linux walker contract: 0 visits only the
		// root, negative values are unbounded (the daemon defaults an
		// absent maxDepth query parameter to -1).
		budget := windowsA11yWalkBudget
		node, err := walkWindowsNode(walker, root, req.MaxDepth, &budget)
		if err != nil {
			return err
		}
		if node == nil {
			return errA11yNoAccessibleRoot
		}
		resp = &computeruse.AccessibilityTreeResponse{Root: *node, Truncated: budget <= 0}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *ComputerUse) FindAccessibilityNodes(req *computeruse.FindAccessibilityNodesRequest) (*computeruse.AccessibilityNodesResponse, error) {
	if req == nil {
		req = &computeruse.FindAccessibilityNodesRequest{}
	}

	var resp *computeruse.AccessibilityNodesResponse
	err := runOnSTA(func() error {
		automation, err := newWindowsAutomation()
		if err != nil {
			return err
		}
		defer automation.Release()

		scope, err := parseWindowsWireScope(req.Scope)
		if err != nil {
			return err
		}
		matcher, err := buildWindowsFilterMatcher(req)
		if err != nil {
			return err
		}
		limit := normalizeWindowsFindLimit(req.Limit)

		root, err := resolveWindowsScopeRoot(automation, scope, req.PID)
		if err != nil {
			return err
		}

		plan := planWindowsFindConditions(req.Role, req.Name, req.NameMatch)
		if plan.impossible {
			// The role vocabulary is closed (control-type table plus the
			// control_type_N fallback), so no node can ever satisfy this
			// filter; skip the walk entirely.
			releaseElement(root)
			resp = &computeruse.AccessibilityNodesResponse{Matches: []computeruse.AccessibilityNode{}}
			return nil
		}

		matches := make([]computeruse.AccessibilityNode, 0, 16)
		budget := windowsA11yWalkBudget
		truncated := false

		if condition := buildWindowsFindCondition(automation, plan, req.Name); condition != nil {
			// Role/name pushed into a native UIA condition: the provider
			// filters server-side and only candidates are materialized
			// cross-process. The Go matcher re-checks every candidate (the
			// native condition is a superset) and applies what UIA cannot
			// express: regex name matching and state filters.
			defer condition.Release()
			defer releaseElement(root)

			found, err := root.FindAll(uia.TreeScopeSubtree, condition)
			if err != nil {
				return fmt.Errorf("%w: FindAll: %v", errA11yUnavailable, err)
			}
			defer found.Release()

			length, err := found.Length()
			if err != nil {
				return fmt.Errorf("%w: FindAll length: %v", errA11yUnavailable, err)
			}
			for i := int32(0); i < length; i++ {
				if budget <= 0 {
					truncated = true
					break
				}
				budget--

				elt, err := found.GetElement(i)
				if err != nil || elt == nil {
					continue
				}
				node := windowsNodeFromElement(elt, false)
				if !matcher(node) {
					releaseElement(elt)
					continue
				}

				node.ID = cacheWindowsElement(elt)
				matches = append(matches, *node)
				if len(matches) >= limit {
					truncated = true
					break
				}
			}
		} else {
			// Nothing natively expressible (regex name or state-only
			// filters): walk lazily and stop as soon as the limit or node
			// budget is reached instead of materializing the whole tree.
			walker, err := automation.ControlViewWalker()
			if err != nil {
				releaseElement(root)
				return fmt.Errorf("%w: ControlViewWalker: %v", errA11yUnavailable, err)
			}
			defer walker.Release()
			truncated = findWindowsWalk(walker, root, matcher, &matches, limit, &budget)
		}

		resp = &computeruse.AccessibilityNodesResponse{Matches: matches, Truncated: truncated}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *ComputerUse) FocusAccessibilityNode(req *computeruse.AccessibilityNodeRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", errA11yInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", errA11yInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return errA11yNodeNotFound
		}
		if err := elt.SetFocus(); err != nil {
			return fmt.Errorf("%w: SetFocus: %v", errA11yActionNotSupported, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

func (c *ComputerUse) InvokeAccessibilityNode(req *computeruse.AccessibilityInvokeRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", errA11yInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", errA11yInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return errA11yNodeNotFound
		}
		return invokeWindowsElement(elt, req.Action)
	})
	if err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

func (c *ComputerUse) SetAccessibilityNodeValue(req *computeruse.AccessibilitySetValueRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", errA11yInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", errA11yInvalidRequest)
	}
	if strings.ContainsRune(req.Value, 0) {
		// ValuePattern.SetValue marshals through ole.SysAllocString, which
		// panics on interior NUL — on the locked STA thread that would
		// kill the plugin process (same class as the find name filter).
		return nil, fmt.Errorf("%w: value must not contain NUL", errA11yInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return errA11yNodeNotFound
		}
		pattern, err := elt.GetValuePattern()
		if err != nil || pattern == nil {
			return fmt.Errorf("%w: ValuePattern", errA11yActionNotSupported)
		}
		defer pattern.Release()
		if readonly, err := pattern.CurrentIsReadonly(); err == nil && readonly {
			return fmt.Errorf("%w: value is read-only", errA11yActionNotSupported)
		}
		if err := pattern.SetValue(req.Value); err != nil {
			return fmt.Errorf("%w: SetValue: %v", errA11yActionNotSupported, err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

func newWindowsAutomation() (*uia.UIAutomation, error) {
	if staInitErr != nil {
		return nil, fmt.Errorf("%w: CoInitializeEx: %v", errA11yUnavailable, staInitErr)
	}
	automation, err := uia.NewUIAutomation()
	if err != nil {
		return nil, fmt.Errorf("%w: NewUIAutomation: %v", errA11yUnavailable, err)
	}
	return automation, nil
}

func parseWindowsWireScope(s string) (windowsA11yScope, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "focused":
		return windowsA11yScopeFocused, nil
	case "pid":
		return windowsA11yScopePID, nil
	case "all":
		return windowsA11yScopeAll, nil
	default:
		return "", fmt.Errorf("%w: got %q, expected focused|pid|all", errA11yInvalidScope, s)
	}
}

func resolveWindowsScopeRoot(automation *uia.UIAutomation, scope windowsA11yScope, pid int) (*uia.Element, error) {
	switch scope {
	case windowsA11yScopeFocused:
		elt, err := automation.GetFocusedElement()
		if err != nil || elt == nil {
			return nil, fmt.Errorf("%w: GetFocusedElement: %v", errA11yNoAccessibleRoot, err)
		}
		// Linux resolves scope=focused to the focused application's root,
		// not the focused control; hoist to the top-level window so tree
		// and find cover the whole foreground app.
		return windowsTopLevelAncestor(automation, elt), nil
	case windowsA11yScopeAll:
		root, err := automation.GetRootElement()
		if err != nil || root == nil {
			return nil, fmt.Errorf("%w: GetRootElement: %v", errA11yNoAccessibleRoot, err)
		}
		return root, nil
	case windowsA11yScopePID:
		if pid <= 0 {
			return nil, fmt.Errorf("%w: pid must be positive", errA11yInvalidRequest)
		}
		root, err := automation.GetRootElement()
		if err != nil || root == nil {
			return nil, fmt.Errorf("%w: GetRootElement: %v", errA11yNoAccessibleRoot, err)
		}
		elt, err := findWindowsRootByPID(automation, root, pid)
		releaseElement(root) // only needed as the FindAll anchor
		if err != nil {
			return nil, err
		}
		if elt == nil {
			return nil, errA11yNoAccessibleRoot
		}
		return elt, nil
	default:
		return nil, fmt.Errorf("%w: unknown scope %q", errA11yInvalidScope, scope)
	}
}

// findWindowsRootByPID resolves scope=pid through a native ProcessId
// property condition so UIA filters server-side instead of materializing the
// entire desktop tree cross-process. Top-level windows (desktop children)
// are preferred — a Window-typed match over the first match — with a deep,
// still condition-narrowed probe as fallback for processes whose UI is
// hosted inside another process's window. Returns nil when the pid owns no
// accessible element.
func findWindowsRootByPID(automation *uia.UIAutomation, root *uia.Element, pid int) (*uia.Element, error) {
	value := ole.NewVariant(ole.VT_I4, int64(pid))
	condition, err := automation.CreatePropertyCondition(uia.ProcessIdPropertyId, &value)
	ole.VariantClear(&value)
	if err != nil || condition == nil {
		return nil, fmt.Errorf("%w: CreatePropertyCondition: %v", errA11yUnavailable, err)
	}
	defer condition.Release()

	found, err := root.FindAll(uia.TreeScopeChildren, condition)
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll: %v", errA11yUnavailable, err)
	}
	defer found.Release()

	length, err := found.Length()
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll length: %v", errA11yUnavailable, err)
	}

	var fallback *uia.Element
	for i := int32(0); i < length; i++ {
		elt, err := found.GetElement(i)
		if err != nil || elt == nil {
			continue
		}
		controlType, err := elt.CurrentControlType()
		if err == nil && controlType == uia.WindowControlTypeId {
			releaseElement(fallback)
			return elt, nil
		}
		if fallback == nil {
			fallback = elt
			continue
		}
		releaseElement(elt)
	}
	if fallback != nil {
		return fallback, nil
	}

	// No top-level window owned by the pid; probe the subtree with the same
	// native condition (FindFirst short-circuits at the first match).
	elt, err := root.FindFirst(uia.TreeScopeDescendants, condition)
	if err != nil {
		return nil, fmt.Errorf("%w: FindFirst: %v", errA11yUnavailable, err)
	}
	return elt, nil
}

// windowsTopLevelAncestor climbs from elt to its top-level window — the
// element whose parent is the desktop root. Returns elt unchanged when the
// parent chain cannot be resolved. Ownership of the returned element passes
// to the caller; intermediate elements are released.
func windowsTopLevelAncestor(automation *uia.UIAutomation, elt *uia.Element) *uia.Element {
	walker, err := automation.ControlViewWalker()
	if err != nil {
		return elt
	}
	defer walker.Release()

	current := elt
	parent, err := walker.GetParentElement(current)
	if err != nil || parent == nil {
		return current // already the desktop root (or detached)
	}
	// Cap the climb defensively; real window chains are a handful of hops.
	for hops := 0; hops < 64; hops++ {
		grand, err := walker.GetParentElement(parent)
		if err != nil || grand == nil {
			// parent is the desktop root: current is the top-level window.
			releaseElement(parent)
			return current
		}
		releaseElement(current)
		current = parent
		parent = grand
	}
	releaseElement(parent)
	return current
}

func walkWindowsNode(walker *uia.TreeWalker, elt *uia.Element, maxDepth int, budget *int) (*computeruse.AccessibilityNode, error) {
	if *budget <= 0 || elt == nil {
		return nil, nil
	}
	*budget--

	node := windowsNodeFromElement(elt, true)
	if !windowsDepthAllowsDescent(maxDepth) {
		return node, nil
	}
	nextDepth := windowsNextDepth(maxDepth)

	child, err := walker.GetFirstChildElement(elt)
	if err != nil || child == nil {
		return node, nil
	}
	for child != nil {
		if *budget <= 0 {
			break
		}
		current := child
		next, _ := walker.GetNextSiblingElement(current)
		childNode, err := walkWindowsNode(walker, current, nextDepth, budget)
		if err != nil {
			releaseElement(current)
			child = next
			continue
		}
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
		child = next
	}
	return node, nil
}

func windowsNodeFromElement(elt *uia.Element, cache bool) *computeruse.AccessibilityNode {
	node := &computeruse.AccessibilityNode{
		Role:        windowsElementRole(elt),
		Name:        windowsStringProperty(elt.CurrentName),
		Description: windowsElementDescription(elt),
		Bounds:      windowsElementBounds(elt),
		States:      windowsElementStates(elt),
		Actions:     windowsElementActions(elt),
	}
	if cache {
		node.ID = cacheWindowsElement(elt)
	}
	return node
}

// windowsElementRole maps the element's control type to the canonical role
// vocabulary. LocalizedControlType is intentionally never consulted: it is
// translated by the OS language pack, which would break cross-platform role
// filters.
func windowsElementRole(elt *uia.Element) string {
	controlType, err := elt.CurrentControlType()
	if err != nil {
		return windowsRoleUnknown
	}
	return windowsRoleName(controlType)
}

func windowsElementDescription(elt *uia.Element) string {
	return strings.TrimSpace(windowsStringProperty(elt.CurrentHelpText))
}

func windowsElementBounds(elt *uia.Element) computeruse.AccessibilityBounds {
	rect, err := elt.CurrentBoundingRectangle()
	if err != nil {
		return computeruse.AccessibilityBounds{}
	}
	left := int(int32(rect.Left))
	top := int(int32(rect.Top))
	right := int(int32(rect.Right))
	bottom := int(int32(rect.Bottom))
	return computeruse.AccessibilityBounds{
		X:      left,
		Y:      top,
		Width:  right - left,
		Height: bottom - top,
	}
}

func windowsElementStates(elt *uia.Element) []string {
	states := make([]string, 0, 8)
	if windowsBoolProperty(elt.CurrentIsEnabled) {
		states = append(states, "enabled", "sensitive")
	}
	if windowsBoolProperty(elt.CurrentIsKeyboardFocusable) {
		states = append(states, "focusable")
	}
	if windowsBoolProperty(elt.CurrentHasKeyboardFocus) {
		states = append(states, "focused", "active")
	}
	if offscreen, err := elt.CurrentIsOffscreen(); err == nil {
		if offscreen {
			states = append(states, "offscreen")
		} else {
			states = append(states, "visible", "showing")
		}
	}
	if windowsBoolProperty(elt.CurrentIsPassword) {
		states = append(states, "password")
	}
	if pattern, err := elt.GetValuePattern(); err == nil && pattern != nil {
		if readonly, err := pattern.CurrentIsReadonly(); err == nil && readonly {
			states = append(states, "read_only")
		}
		pattern.Release()
	}
	return states
}

func windowsElementActions(elt *uia.Element) []string {
	actions := make([]string, 0, 3)
	if pattern, err := elt.GetInvokePattern(); err == nil && pattern != nil {
		actions = append(actions, actionInvoke)
		pattern.Release()
	}
	if pattern, err := elt.GetSelectionItemPattern(); err == nil && pattern != nil {
		actions = append(actions, actionSelect)
		pattern.Release()
	}
	if pattern, err := elt.GetValuePattern(); err == nil && pattern != nil {
		actions = append(actions, actionSetValue)
		pattern.Release()
	}
	return actions
}

func buildWindowsFilterMatcher(req *computeruse.FindAccessibilityNodesRequest) (func(*computeruse.AccessibilityNode) bool, error) {
	nameMatch := req.NameMatch
	if nameMatch == "" {
		nameMatch = "substring"
	}
	if nameMatch != "exact" && nameMatch != "substring" && nameMatch != "regex" {
		return nil, fmt.Errorf("%w: unknown nameMatch mode %q, want exact|substring|regex", errA11yInvalidRequest, nameMatch)
	}

	var nameRe *regexp.Regexp
	if req.Name != "" && nameMatch == "regex" {
		re, err := regexp.Compile(req.Name)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid regex for name filter: %v", errA11yInvalidRequest, err)
		}
		nameRe = re
	}

	role := strings.ToLower(req.Role)
	states := append([]string(nil), req.States...)
	return func(n *computeruse.AccessibilityNode) bool {
		if role != "" && strings.ToLower(n.Role) != role {
			return false
		}
		if req.Name != "" {
			switch nameMatch {
			case "exact":
				if n.Name != req.Name {
					return false
				}
			case "substring":
				if !strings.Contains(n.Name, req.Name) {
					return false
				}
			case "regex":
				if nameRe == nil || !nameRe.MatchString(n.Name) {
					return false
				}
			}
		}
		for _, want := range states {
			if !containsWindowsString(n.States, want) {
				return false
			}
		}
		return true
	}, nil
}

// windowsFindPlan describes how much of a find filter can be pushed into
// native UIA search conditions. Native conditions only ever narrow the
// candidate set to a superset of the matcher's accept set —
// buildWindowsFilterMatcher remains the source of truth and re-checks every
// candidate — so a missing or wider condition costs time, never correctness.
type windowsFindPlan struct {
	// roleControlTypes lists the control types OR-ed into a native role
	// condition. Empty means the role cannot narrow natively.
	roleControlTypes []uia.ControlTypeId
	// nameFlags holds the property-condition flags for a native name
	// condition; meaningful only when nameNative is true.
	nameFlags  uia.PropertyConditionFlags
	nameNative bool
	// impossible marks filters no emitted node can ever satisfy (the role
	// vocabulary is closed), allowing an immediate empty result.
	impossible bool
}

// planWindowsFindConditions decides which filter parts are natively
// expressible. Pure logic, unit-tested; assumes nameMatch was already
// validated by buildWindowsFilterMatcher.
func planWindowsFindConditions(role, name, nameMatch string) windowsFindPlan {
	var plan windowsFindPlan
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "" {
		switch ids, ok := windowsControlTypesForRole(role); {
		case role == windowsRoleUnknown:
			// "unknown" doubles as the fallback for unreadable control
			// types, so narrowing it to ControlType==Custom could drop
			// candidates the matcher would accept. Walk unnarrowed.
		case ok:
			plan.roleControlTypes = ids
		default:
			plan.impossible = true
			return plan
		}
	}
	// An interior NUL cannot round-trip through a BSTR: go-ole's
	// SysAllocString panics on it (syscall.StringToUTF16Ptr), killing the
	// locked STA thread and with it the plugin process. Such names stay
	// Go-side, where the matcher returns the same (empty) match set the
	// Linux implementation produces.
	if name != "" && !strings.ContainsRune(name, 0) {
		switch nameMatch {
		case "", "substring":
			// IgnoreCase only widens the candidate set (safe); the
			// matcher restores exact case semantics. MatchSubstring needs
			// Windows 10 1809+ — the builder degrades to a Go-side-only
			// name filter when the OS rejects it.
			plan.nameFlags = uia.PropertyConditionFlagsIgnoreCase | uia.PropertyConditionFlagsMatchSubstring
			plan.nameNative = true
		case "exact":
			plan.nameFlags = uia.PropertyConditionFlagsIgnoreCase
			plan.nameNative = true
		}
		// regex stays Go-side.
	}
	return plan
}

// buildWindowsFindCondition turns a plan into a UIA condition, best-effort:
// any COM failure degrades to a wider (or nil) condition rather than an
// error, because a nil condition falls back to the walker path and the
// Go-side matcher re-filters every candidate anyway. The filter is AND-ed
// with the control-view condition so native FindAll searches the same
// element universe as the walker fallback and the tree endpoint
// (ControlViewWalker); a bare property condition would traverse the raw
// view and surface elements those paths never emit.
func buildWindowsFindCondition(automation *uia.UIAutomation, plan windowsFindPlan, name string) *uia.Condition {
	condition := windowsRoleCondition(automation, plan.roleControlTypes)
	if plan.nameNative && name != "" {
		nameCondition := windowsNameCondition(automation, name, plan.nameFlags)
		switch {
		case condition == nil:
			condition = nameCondition
		case nameCondition != nil:
			combined, err := automation.CreateAndCondition(condition, nameCondition)
			condition.Release()
			nameCondition.Release()
			if err != nil || combined == nil {
				return nil
			}
			condition = combined
		}
	}
	if condition == nil {
		return nil
	}
	controlView, err := automation.ControlViewCondition()
	if err != nil || controlView == nil {
		condition.Release()
		return nil
	}
	combined, err := automation.CreateAndCondition(condition, controlView)
	condition.Release()
	controlView.Release()
	if err != nil || combined == nil {
		return nil
	}
	return combined
}

// windowsRoleCondition ORs a ControlType property condition for every
// control type mapped to the requested role. Partial failures abandon the
// whole role condition: an incomplete OR would narrow to a subset of the
// matcher's accept set and silently drop matches.
func windowsRoleCondition(automation *uia.UIAutomation, controlTypes []uia.ControlTypeId) *uia.Condition {
	var combined *uia.Condition
	for _, controlType := range controlTypes {
		value := ole.NewVariant(ole.VT_I4, int64(controlType))
		condition, err := automation.CreatePropertyCondition(uia.ControlTypePropertyId, &value)
		ole.VariantClear(&value)
		if err != nil || condition == nil {
			releaseCondition(combined)
			return nil
		}
		if combined == nil {
			combined = condition
			continue
		}
		next, err := automation.CreateOrCondition(combined, condition)
		combined.Release()
		condition.Release()
		if err != nil || next == nil {
			return nil
		}
		combined = next
	}
	return combined
}

// windowsNameCondition builds a native Name property condition. Returns nil
// when the name cannot be marshaled to a BSTR (interior NUL) or the OS
// rejects the flags (MatchSubstring needs Windows 10 1809+); the caller
// then relies on the Go-side matcher alone.
func windowsNameCondition(automation *uia.UIAutomation, name string, flags uia.PropertyConditionFlags) *uia.Condition {
	if strings.ContainsRune(name, 0) {
		// ole.SysAllocString panics on interior NUL.
		return nil
	}
	bstr := ole.SysAllocString(name)
	if bstr == nil {
		return nil
	}
	value := ole.NewVariant(ole.VT_BSTR, int64(uintptr(unsafe.Pointer(bstr))))
	condition, err := automation.CreatePropertyConditionEx(uia.NamePropertyId, &value, flags)
	ole.VariantClear(&value) // frees the BSTR; the condition keeps its own copy
	if err != nil || condition == nil {
		return nil
	}
	return condition
}

func releaseCondition(condition *uia.Condition) {
	if condition != nil {
		condition.Release()
	}
}

// findWindowsWalk mirrors the Linux findWalk: a pre-order depth-first walk
// that stops as soon as the result limit or the node budget is reached, so
// an unnarrowed find never materializes more of the tree than it returns.
// It takes ownership of elt and releases every element it visits unless
// ownership is handed to the element cache for a returned match. Reports
// whether the walk was truncated.
func findWindowsWalk(walker *uia.TreeWalker, elt *uia.Element, match func(*computeruse.AccessibilityNode) bool, out *[]computeruse.AccessibilityNode, limit int, budget *int) bool {
	if elt == nil {
		return false
	}
	if *budget <= 0 {
		releaseElement(elt)
		return true
	}
	*budget--

	node := windowsNodeFromElement(elt, false)
	matched := match(node)

	// Fetch the first child while elt is still alive; children of returned
	// nodes stay nil per the find contract.
	child, _ := walker.GetFirstChildElement(elt)
	if matched {
		node.ID = cacheWindowsElement(elt) // cache takes ownership of elt
		*out = append(*out, *node)
		if len(*out) >= limit {
			releaseElement(child)
			return true
		}
	} else {
		releaseElement(elt)
	}

	for child != nil {
		if *budget <= 0 {
			releaseElement(child)
			return true
		}
		next, _ := walker.GetNextSiblingElement(child)
		if findWindowsWalk(walker, child, match, out, limit, budget) {
			releaseElement(next)
			return true
		}
		child = next
	}
	return *budget <= 0
}

func invokeWindowsElement(elt *uia.Element, action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" || action == actionInvoke || action == "click" || action == "press" {
		pattern, err := elt.GetInvokePattern()
		if err == nil && pattern != nil {
			defer pattern.Release()
			if err := pattern.Invoke(); err != nil {
				return fmt.Errorf("%w: Invoke: %v", errA11yActionNotSupported, err)
			}
			return nil
		}
		if action != "" {
			return fmt.Errorf("%w: InvokePattern", errA11yActionNotSupported)
		}
	}

	if action == "" || action == actionSelect {
		pattern, err := elt.GetSelectionItemPattern()
		if err == nil && pattern != nil {
			defer pattern.Release()
			if err := pattern.Select(); err != nil {
				return fmt.Errorf("%w: Select: %v", errA11yActionNotSupported, err)
			}
			return nil
		}
		if action == actionSelect {
			return fmt.Errorf("%w: SelectionItemPattern", errA11yActionNotSupported)
		}
	}

	return errA11yActionNotSupported
}

func cacheWindowsElement(elt *uia.Element) string {
	handle := newWindowsElementHandle()
	elementMap.put(handle, elt)
	return handle
}

func newWindowsElementHandle() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func windowsStringProperty(fn func() (string, error)) string {
	value, err := fn()
	if err != nil {
		return ""
	}
	return value
}

func windowsBoolProperty(fn func() (bool, error)) bool {
	value, err := fn()
	return err == nil && value
}

// windowsControlTypeRoles maps UIA control type IDs to the canonical,
// non-localized role vocabulary the Linux AT-SPI implementation emits
// (at-spi2-core role names), keeping role filters portable across
// platforms and OS display languages.
var windowsControlTypeRoles = map[uia.ControlTypeId]string{
	uia.ButtonControlTypeId:       "push button",
	uia.CalendarControlTypeId:     "calendar",
	uia.CheckBoxControlTypeId:     "check box",
	uia.ComboBoxControlTypeId:     "combo box",
	uia.EditControlTypeId:         "entry",
	uia.HyperlinkControlTypeId:    "link",
	uia.ImageControlTypeId:        "image",
	uia.ListItemControlTypeId:     "list item",
	uia.ListControlTypeId:         "list",
	uia.MenuControlTypeId:         "menu",
	uia.MenuBarControlTypeId:      "menu bar",
	uia.MenuItemControlTypeId:     "menu item",
	uia.ProgressBarControlTypeId:  "progress bar",
	uia.RadioButtonControlTypeId:  "radio button",
	uia.ScrollBarControlTypeId:    "scroll bar",
	uia.SliderControlTypeId:       "slider",
	uia.SpinnerControlTypeId:      "spin button",
	uia.StatusBarControlTypeId:    "status bar",
	uia.TabControlTypeId:          "page tab list",
	uia.TabItemControlTypeId:      "page tab",
	uia.TextControlTypeId:         "label",
	uia.ToolBarControlTypeId:      "tool bar",
	uia.ToolTipControlTypeId:      "tool tip",
	uia.TreeControlTypeId:         "tree",
	uia.TreeItemControlTypeId:     "tree item",
	uia.CustomControlTypeId:       "unknown",
	uia.GroupControlTypeId:        "panel",
	uia.ThumbControlTypeId:        "unknown",
	uia.DataGridControlTypeId:     "table",
	uia.DataItemControlTypeId:     "table cell",
	uia.DocumentControlTypeId:     "document frame",
	uia.SplitButtonControlTypeId:  "push button",
	uia.WindowControlTypeId:       "frame",
	uia.PaneControlTypeId:         "panel",
	uia.HeaderControlTypeId:       "header",
	uia.HeaderItemControlTypeId:   "column header",
	uia.TableControlTypeId:        "table",
	uia.TitleBarControlTypeId:     "title bar",
	uia.SeparatorControlTypeId:    "separator",
	uia.SemanticZoomControlTypeId: "panel",
	uia.AppBarControlTypeId:       "tool bar",
}

const (
	// windowsRoleUnknown is emitted for the Custom/Thumb control types and
	// whenever the control type cannot be read.
	windowsRoleUnknown = "unknown"
	// windowsRoleRawPrefix prefixes the numeric fallback role for control
	// type IDs outside the standard UIA set.
	windowsRoleRawPrefix = "control_type_"
)

// windowsRoleControlTypes is the inverse of windowsControlTypeRoles: role
// name -> every control type that emits it, sorted for deterministic native
// condition construction.
var windowsRoleControlTypes = buildWindowsRoleControlTypes()

func buildWindowsRoleControlTypes() map[string][]uia.ControlTypeId {
	m := make(map[string][]uia.ControlTypeId, len(windowsControlTypeRoles))
	for id, role := range windowsControlTypeRoles {
		m[role] = append(m[role], id)
	}
	for _, ids := range m {
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	}
	return m
}

// windowsRoleName resolves a control type ID to its canonical role string.
// IDs outside the standard UIA set fall back to a stable numeric form so
// they remain filterable.
func windowsRoleName(controlType uia.ControlTypeId) string {
	if role, ok := windowsControlTypeRoles[controlType]; ok {
		return role
	}
	return fmt.Sprintf("%s%d", windowsRoleRawPrefix, controlType)
}

// windowsControlTypesForRole inverts windowsRoleName for native filter
// narrowing. The role is matched case-insensitively, mirroring the Go-side
// matcher semantics.
func windowsControlTypesForRole(role string) ([]uia.ControlTypeId, bool) {
	role = strings.ToLower(strings.TrimSpace(role))
	if ids, ok := windowsRoleControlTypes[role]; ok {
		return ids, true
	}
	if raw, ok := strings.CutPrefix(role, windowsRoleRawPrefix); ok {
		if v, err := strconv.Atoi(raw); err == nil {
			return []uia.ControlTypeId{uia.ControlTypeId(v)}, true
		}
	}
	return nil, false
}

// Depth semantics mirror the Linux walker contract: 0 visits only the
// current node, positive values bound descent, negative values are
// unbounded.
func windowsDepthAllowsDescent(maxDepth int) bool {
	return maxDepth != 0
}

func windowsNextDepth(maxDepth int) int {
	if maxDepth > 0 {
		return maxDepth - 1
	}
	return maxDepth
}

// normalizeWindowsFindLimit applies the Linux find limit defaults: non
// positive limits fall back to the default, oversized limits are capped.
func normalizeWindowsFindLimit(limit int) int {
	if limit <= 0 {
		return windowsFindDefaultLimit
	}
	if limit > windowsFindCeilingLimit {
		return windowsFindCeilingLimit
	}
	return limit
}

func containsWindowsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func releaseElement(elt *uia.Element) {
	if elt != nil {
		elt.Release()
	}
}
