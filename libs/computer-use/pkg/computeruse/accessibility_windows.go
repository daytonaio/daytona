//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"crypto/rand"
	"errors"
	"fmt"
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

// The wire-contract sentinel errors (ErrA11yUnavailable and friends), scope
// parsing, walk/find limits, and filter matcher semantics are shared with
// the Linux implementation in accessibility_common.go.

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
	mu     sync.Mutex
	elts   map[string]*cachedElement
	nextGC time.Time
}

type cachedElement struct {
	elt    *uia.Element
	expiry time.Time
}

const (
	elementCacheTTL = 5 * time.Minute

	// elementCacheGCInterval amortizes expiry sweeps. gcLocked is a full
	// map scan, so running it on every put made budget-sized tree walks
	// accidentally quadratic on the serialized STA thread; expired entries
	// are also reclaimed lazily per-key in get.
	elementCacheGCInterval = time.Minute

	// elementCacheMaxEntries caps live COM element proxies. The margin
	// above a11yWalkBudget is load-bearing: overflow eviction removes
	// oldest-expiry entries first and a single request inserts at most
	// a11yWalkBudget entries (always the newest expiries), so eviction can
	// only ever reclaim earlier requests' entries — never an element the
	// in-flight walk still dereferences from its recursion stack.
	elementCacheMaxEntries = 30000
)

var elementMap = &elementCache{elts: map[string]*cachedElement{}}

func (c *elementCache) put(handle string, elt *uia.Element) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	if now.After(c.nextGC) {
		c.gcLocked(now)
		c.nextGC = now.Add(elementCacheGCInterval)
	}
	if len(c.elts) >= elementCacheMaxEntries {
		c.evictOldestLocked()
	}
	c.elts[handle] = &cachedElement{elt: elt, expiry: now.Add(elementCacheTTL)}
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

// remove evicts a single handle, releasing its element. Used when an action
// discovers the underlying UI element is gone, so later calls 404 without
// another UIA round-trip.
func (c *elementCache) remove(handle string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ce, ok := c.elts[handle]; ok {
		delete(c.elts, handle)
		releaseElement(ce.elt)
	}
}

// evictOldestLocked drops the tenth of the cache closest to expiry — expiry
// order is insertion/refresh order, so this is least-recently-used-first.
// Evicting in batches keeps overflow handling amortized instead of
// re-scanning the whole map for every insertion at the cap.
func (c *elementCache) evictOldestLocked() {
	drop := len(c.elts) / 10
	if drop < 1 {
		drop = 1
	}
	type expiringHandle struct {
		handle string
		expiry time.Time
	}
	entries := make([]expiringHandle, 0, len(c.elts))
	for handle, ce := range c.elts {
		entries = append(entries, expiringHandle{handle, ce.expiry})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].expiry.Before(entries[j].expiry) })
	if drop > len(entries) {
		drop = len(entries)
	}
	for _, e := range entries[:drop] {
		releaseElement(c.elts[e.handle].elt)
		delete(c.elts, e.handle)
	}
}

func (c *elementCache) gcLocked(now time.Time) {
	for k, v := range c.elts {
		if now.After(v.expiry) {
			delete(c.elts, k)
			releaseElement(v.elt)
		}
	}
}

const (
	actionInvoke   = "invoke"
	actionSelect   = "select"
	actionSetValue = "set_value"
)

func (c *ComputerUse) GetAccessibilityTree(req *computeruse.GetAccessibilityTreeRequest) (*computeruse.AccessibilityTreeResponse, error) {
	if req == nil {
		req = &computeruse.GetAccessibilityTreeRequest{}
	}

	// Pure request validation runs before entering the STA (Linux parity:
	// scope is parsed before connecting to the bus), so a malformed request
	// 400s even on hosts where COM init fails and never queues behind
	// in-flight UIA work on the single STA thread.
	scope, err := parseWireScope(req.Scope)
	if err != nil {
		return nil, err
	}

	var resp *computeruse.AccessibilityTreeResponse
	err = runOnSTA(func() error {
		automation, err := newWindowsAutomation()
		if err != nil {
			return err
		}
		defer automation.Release()

		root, err := resolveWindowsScopeRoot(automation, scope, req.PID)
		if err != nil {
			return err
		}

		walker, err := automation.ControlViewWalker()
		if err != nil {
			releaseElement(root)
			return fmt.Errorf("%w: ControlViewWalker: %v", ErrA11yUnavailable, err)
		}
		defer walker.Release()

		// MaxDepth follows the Linux walker contract: 0 visits only the
		// root, negative values are unbounded (the daemon defaults an
		// absent maxDepth query parameter to -1).
		budget := a11yWalkBudget
		node := walkWindowsNode(walker, root, req.MaxDepth, &budget)
		if node == nil {
			return ErrNoAccessibleRoot
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

	// Pure request validation runs before entering the STA, as in
	// GetAccessibilityTree.
	scope, err := parseWireScope(req.Scope)
	if err != nil {
		return nil, err
	}
	matcher, err := buildWindowsFilterMatcher(req)
	if err != nil {
		return nil, err
	}
	limit := normalizeFindLimit(req.Limit)

	var resp *computeruse.AccessibilityNodesResponse
	err = runOnSTA(func() error {
		automation, err := newWindowsAutomation()
		if err != nil {
			return err
		}
		defer automation.Release()

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
		budget := a11yWalkBudget
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
				return fmt.Errorf("%w: FindAll: %v", ErrA11yUnavailable, err)
			}
			defer found.Release()

			length, err := found.Length()
			if err != nil {
				return fmt.Errorf("%w: FindAll length: %v", ErrA11yUnavailable, err)
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
				return fmt.Errorf("%w: ControlViewWalker: %v", ErrA11yUnavailable, err)
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
		return nil, fmt.Errorf("%w: request is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return ErrNodeNotFound
		}
		if err := elt.SetFocus(); err != nil {
			return classifyWindowsActionError(req.ID, "SetFocus", err)
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
		return nil, fmt.Errorf("%w: request is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return ErrNodeNotFound
		}
		return invokeWindowsElement(req.ID, elt, req.Action)
	})
	if err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

func (c *ComputerUse) SetAccessibilityNodeValue(req *computeruse.AccessibilitySetValueRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("%w: request is required", ErrInvalidRequest)
	}
	if strings.TrimSpace(req.ID) == "" {
		return nil, fmt.Errorf("%w: id is required", ErrInvalidRequest)
	}
	if strings.ContainsRune(req.Value, 0) {
		// ValuePattern.SetValue marshals through ole.SysAllocString, which
		// panics on interior NUL — on the locked STA thread that would
		// kill the plugin process (same class as the find name filter).
		return nil, fmt.Errorf("%w: value must not contain NUL", ErrInvalidRequest)
	}

	err := runOnSTA(func() error {
		elt, ok := elementMap.get(req.ID)
		if !ok || elt == nil {
			return ErrNodeNotFound
		}
		pattern, err := windowsValuePattern(elt)
		if err != nil {
			return classifyWindowsActionError(req.ID, "ValuePattern", err)
		}
		if pattern == nil {
			return fmt.Errorf("%w: ValuePattern", ErrActionNotSupported)
		}
		defer pattern.Release()
		if readonly, err := pattern.CurrentIsReadonly(); err == nil && readonly {
			return fmt.Errorf("%w: value is read-only", ErrActionNotSupported)
		}
		if err := pattern.SetValue(req.Value); err != nil {
			return classifyWindowsActionError(req.ID, "SetValue", err)
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
		return nil, fmt.Errorf("%w: CoInitializeEx: %v", ErrA11yUnavailable, staInitErr)
	}
	automation, err := uia.NewUIAutomation()
	if err != nil {
		return nil, fmt.Errorf("%w: NewUIAutomation: %v", ErrA11yUnavailable, err)
	}
	return automation, nil
}

func resolveWindowsScopeRoot(automation *uia.UIAutomation, scope A11yScope, pid int) (*uia.Element, error) {
	switch scope {
	case A11yScopeFocused:
		elt, err := automation.GetFocusedElement()
		if err != nil || elt == nil {
			return nil, fmt.Errorf("%w: GetFocusedElement: %v", ErrNoAccessibleRoot, err)
		}
		// Linux resolves scope=focused to the focused application's root,
		// not the focused control; hoist to the top-level window so tree
		// and find cover the whole foreground app.
		return windowsTopLevelAncestor(automation, elt), nil
	case A11yScopeAll:
		root, err := automation.GetRootElement()
		if err != nil || root == nil {
			return nil, fmt.Errorf("%w: GetRootElement: %v", ErrNoAccessibleRoot, err)
		}
		return root, nil
	case A11yScopePID:
		if pid <= 0 {
			return nil, fmt.Errorf("%w: pid must be positive", ErrInvalidRequest)
		}
		root, err := automation.GetRootElement()
		if err != nil || root == nil {
			return nil, fmt.Errorf("%w: GetRootElement: %v", ErrNoAccessibleRoot, err)
		}
		elt, err := findWindowsRootByPID(automation, root, pid)
		releaseElement(root) // only needed as the FindAll anchor
		if err != nil {
			return nil, err
		}
		if elt == nil {
			return nil, ErrNoAccessibleRoot
		}
		return elt, nil
	default:
		return nil, fmt.Errorf("%w: unknown scope %q", ErrInvalidScope, scope)
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
		return nil, fmt.Errorf("%w: CreatePropertyCondition: %v", ErrA11yUnavailable, err)
	}
	defer condition.Release()

	found, err := root.FindAll(uia.TreeScopeChildren, condition)
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll: %v", ErrA11yUnavailable, err)
	}
	defer found.Release()

	length, err := found.Length()
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll length: %v", ErrA11yUnavailable, err)
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
		return nil, fmt.Errorf("%w: FindFirst: %v", ErrA11yUnavailable, err)
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

// walkWindowsNode recursively materialises a wire node tree rooted at elt.
// It takes ownership of elt: ownership passes to the element cache when the
// node is materialised, and elt is released on the budget-exhausted entry
// path. budget is decremented on every visited node; when it hits zero,
// descent stops and the caller infers truncation.
func walkWindowsNode(walker *uia.TreeWalker, elt *uia.Element, maxDepth int, budget *int) *computeruse.AccessibilityNode {
	if elt == nil {
		return nil
	}
	if *budget <= 0 {
		releaseElement(elt)
		return nil
	}
	*budget--

	node := windowsNodeFromElement(elt, true) // cache takes ownership of elt
	if !a11yDepthAllowsDescent(maxDepth) {
		return node
	}
	nextDepth := a11yNextDepth(maxDepth)

	child, err := walker.GetFirstChildElement(elt)
	if err != nil {
		return node
	}
	for child != nil {
		if *budget <= 0 {
			releaseElement(child)
			break
		}
		// Fetch the sibling before recursing: the recursion hands child to
		// the element cache, after which this frame must not touch it.
		next, _ := walker.GetNextSiblingElement(child)
		if childNode := walkWindowsNode(walker, child, nextDepth, budget); childNode != nil {
			node.Children = append(node.Children, childNode)
		}
		child = next
	}
	return node
}

func windowsNodeFromElement(elt *uia.Element, cache bool) *computeruse.AccessibilityNode {
	// One ValuePattern fetch feeds both the read_only state and the
	// set_value action; a node losing its provider mid-walk just reports
	// fewer capabilities.
	valuePattern, _ := windowsValuePattern(elt)
	node := &computeruse.AccessibilityNode{
		Role:        windowsElementRole(elt),
		Name:        windowsStringProperty(elt.CurrentName),
		Description: windowsElementDescription(elt),
		Bounds:      windowsElementBounds(elt),
		States:      windowsElementStates(elt, valuePattern),
		Actions:     windowsElementActions(elt, valuePattern != nil),
	}
	if valuePattern != nil {
		valuePattern.Release()
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

func windowsElementStates(elt *uia.Element, valuePattern *uia.ValuePattern) []string {
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
	if valuePattern != nil {
		if readonly, err := valuePattern.CurrentIsReadonly(); err == nil && readonly {
			states = append(states, "read_only")
		}
	}
	return states
}

func windowsElementActions(elt *uia.Element, hasValuePattern bool) []string {
	actions := make([]string, 0, 3)
	if windowsHasPattern(elt, uia.InvokePatternId) {
		actions = append(actions, actionInvoke)
	}
	if windowsHasPattern(elt, uia.SelectionItemPatternId) {
		actions = append(actions, actionSelect)
	}
	if hasValuePattern {
		actions = append(actions, actionSetValue)
	}
	return actions
}

// The binding's typed pattern getters (GetInvokePattern and friends)
// QueryInterface the IUnknown out-param of GetCurrentPattern without ever
// releasing it, permanently leaking one COM reference per successful call
// (binding replacement is deferred). The helpers below own the whole fetch:
// probe, QI, release the intermediate, hand back only the typed interface.

// windowsHasPattern reports whether the element supports a UIA pattern
// without retaining any reference — capability probing for States/Actions.
func windowsHasPattern(elt *uia.Element, patternId uia.PatternId) bool {
	obj, err := elt.GetCurrentPattern(patternId)
	if err != nil || obj == nil {
		return false
	}
	obj.Release()
	return true
}

// windowsPattern fetches a typed UIA pattern interface. Returns (nil, nil)
// when the element does not support the pattern (S_OK + null out-param) and
// the raw COM error when the call itself fails (e.g. the element is gone).
// The caller owns the returned interface and must Release it.
func windowsPattern(elt *uia.Element, patternId uia.PatternId, iid *ole.GUID) (unsafe.Pointer, error) {
	obj, err := elt.GetCurrentPattern(patternId)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	typed, err := obj.QueryInterface(iid)
	obj.Release()
	if err != nil {
		return nil, err
	}
	return unsafe.Pointer(typed), nil
}

func windowsValuePattern(elt *uia.Element) (*uia.ValuePattern, error) {
	p, err := windowsPattern(elt, uia.ValuePatternId, uia.IID_IUIAutomationValuePattern)
	return (*uia.ValuePattern)(p), err
}

func windowsInvokePattern(elt *uia.Element) (*uia.InvokePattern, error) {
	p, err := windowsPattern(elt, uia.InvokePatternId, uia.IID_IUIAutomationInvokePattern)
	return (*uia.InvokePattern)(p), err
}

func windowsSelectionItemPattern(elt *uia.Element) (*uia.SelectionItemPattern, error) {
	p, err := windowsPattern(elt, uia.SelectionItemPatternId, uia.IID_IUIAutomationSelectionItemPattern)
	return (*uia.SelectionItemPattern)(p), err
}

// buildWindowsFilterMatcher adapts the shared buildA11yMatcher semantics
// (accessibility_common.go) to the wire request/node types.
func buildWindowsFilterMatcher(req *computeruse.FindAccessibilityNodesRequest) (func(*computeruse.AccessibilityNode) bool, error) {
	match, err := buildA11yMatcher(req.Role, req.Name, req.NameMatch, req.States)
	if err != nil {
		return nil, err
	}
	return func(n *computeruse.AccessibilityNode) bool { return match(n.Role, n.Name, n.States) }, nil
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

// UIA/COM HRESULTs whose meaning on action paths is unambiguous (values
// from uiautomationcoreapi.h / winerror.h). Everything outside this set —
// RPC faults, timeouts, E_FAIL, E_OUTOFMEMORY — is deliberately left
// untranslated.
const (
	// uiaErrElementNotAvailable is UIA_E_ELEMENTNOTAVAILABLE: the provider
	// behind the element is gone (window closed, control destroyed).
	uiaErrElementNotAvailable uintptr = 0x80040201
	// uiaErrElementNotEnabled is UIA_E_ELEMENTNOTENABLED: the element is
	// disabled and refuses the action.
	uiaErrElementNotEnabled uintptr = 0x80040200
	// uiaErrNotSupported is UIA_E_NOTSUPPORTED: the provider rejects the
	// requested operation.
	uiaErrNotSupported uintptr = 0x80040204
	// uiaErrInvalidOperation is UIA_E_INVALIDOPERATION: the operation is
	// not valid in the element's current state.
	uiaErrInvalidOperation uintptr = 0x80131509
	// hresultNoInterface (E_NOINTERFACE) and hresultNotImplemented
	// (E_NOTIMPL) are permanent "this object cannot do that" refusals.
	hresultNoInterface    uintptr = 0x80004002
	hresultNotImplemented uintptr = 0x80004001
)

// classifyWindowsActionError follows the Linux classifyDbusError contract on
// action paths: only failures with an unambiguous meaning are translated to
// sentinel errors, anything else (transient RPC faults, timeouts, E_FAIL,
// non-COM errors) is returned as-is so the daemon surfaces it as a
// retryable internal error instead of a permanent refusal. A call that
// failed because the element vanished maps to ErrNodeNotFound (404 —
// "re-find the node") and evicts the dead handle so subsequent calls fail
// fast without another UIA round-trip; refusal HRESULTs map to
// ErrActionNotSupported (400 — "pick another action"). The binding
// surfaces every failing COM call as *ole.OleError holding the raw HRESULT.
func classifyWindowsActionError(handle, op string, err error) error {
	var oleErr *ole.OleError
	if errors.As(err, &oleErr) {
		switch oleErr.Code() {
		case uiaErrElementNotAvailable:
			elementMap.remove(handle)
			return fmt.Errorf("%w: %s: %v", ErrNodeNotFound, op, err)
		case uiaErrElementNotEnabled, uiaErrNotSupported, uiaErrInvalidOperation,
			hresultNoInterface, hresultNotImplemented:
			return fmt.Errorf("%w: %s: %v", ErrActionNotSupported, op, err)
		}
	}
	return fmt.Errorf("%s: %w", op, err)
}

func invokeWindowsElement(handle string, elt *uia.Element, action string) error {
	action = strings.ToLower(strings.TrimSpace(action))
	if action == actionSetValue {
		// Advertised in Actions for value-capable nodes, but the invoke
		// request cannot carry a value; keep the sentinel prefix the daemon
		// matches on and point the caller at the dedicated endpoint.
		return fmt.Errorf("%w: set_value takes a value; use the node value endpoint", ErrActionNotSupported)
	}

	if action == "" || action == actionInvoke || action == "click" || action == "press" {
		pattern, err := windowsInvokePattern(elt)
		if err != nil {
			return classifyWindowsActionError(handle, "InvokePattern", err)
		}
		if pattern != nil {
			defer pattern.Release()
			if err := pattern.Invoke(); err != nil {
				return classifyWindowsActionError(handle, "Invoke", err)
			}
			return nil
		}
		if action != "" {
			return fmt.Errorf("%w: InvokePattern", ErrActionNotSupported)
		}
	}

	if action == "" || action == actionSelect {
		pattern, err := windowsSelectionItemPattern(elt)
		if err != nil {
			return classifyWindowsActionError(handle, "SelectionItemPattern", err)
		}
		if pattern != nil {
			defer pattern.Release()
			if err := pattern.Select(); err != nil {
				return classifyWindowsActionError(handle, "Select", err)
			}
			return nil
		}
		if action == actionSelect {
			return fmt.Errorf("%w: SelectionItemPattern", ErrActionNotSupported)
		}
	}

	return ErrActionNotSupported
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

func releaseElement(elt *uia.Element) {
	if elt != nil {
		elt.Release()
	}
}
