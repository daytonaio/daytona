//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

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
	windowsA11yDefaultDepth = 10
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

		maxDepth := req.MaxDepth
		if maxDepth == 0 {
			maxDepth = windowsA11yDefaultDepth
		}
		budget := windowsA11yWalkBudget
		node, err := walkWindowsNode(walker, root, maxDepth, &budget)
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
		root, err := resolveWindowsScopeRoot(automation, scope, req.PID)
		if err != nil {
			return err
		}

		matcher, err := buildWindowsFilterMatcher(req)
		if err != nil {
			return err
		}

		condition, err := automation.CreateTrueCondition()
		if err != nil {
			return fmt.Errorf("%w: CreateTrueCondition: %v", errA11yUnavailable, err)
		}
		defer condition.Release()

		found, err := root.FindAll(uia.TreeScopeSubtree, condition)
		if err != nil {
			return fmt.Errorf("%w: FindAll: %v", errA11yUnavailable, err)
		}
		defer found.Release()

		limit := req.Limit
		if limit <= 0 {
			limit = windowsFindDefaultLimit
		}
		if limit > windowsFindCeilingLimit {
			limit = windowsFindCeilingLimit
		}

		length, err := found.Length()
		if err != nil {
			return fmt.Errorf("%w: FindAll length: %v", errA11yUnavailable, err)
		}

		matches := make([]computeruse.AccessibilityNode, 0, limit)
		budget := windowsA11yWalkBudget
		truncated := false
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
				truncated = i < length-1
				break
			}
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
		return elt, nil
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

func findWindowsRootByPID(automation *uia.UIAutomation, root *uia.Element, pid int) (*uia.Element, error) {
	condition, err := automation.CreateTrueCondition()
	if err != nil {
		return nil, fmt.Errorf("%w: CreateTrueCondition: %v", errA11yUnavailable, err)
	}
	defer condition.Release()

	found, err := root.FindAll(uia.TreeScopeSubtree, condition)
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll: %v", errA11yUnavailable, err)
	}
	defer found.Release()

	length, err := found.Length()
	if err != nil {
		return nil, fmt.Errorf("%w: FindAll length: %v", errA11yUnavailable, err)
	}

	var fallback *uia.Element
	for i := int32(0); i < length && i < windowsA11yWalkBudget; i++ {
		elt, err := found.GetElement(i)
		if err != nil || elt == nil {
			continue
		}
		processID, err := elt.CurrentProcessId()
		if err != nil || int(processID) != pid {
			releaseElement(elt)
			continue
		}

		controlType, err := elt.CurrentControlType()
		if err == nil && controlType == uia.WindowControlTypeId {
			if fallback != nil {
				releaseElement(fallback)
			}
			return elt, nil
		}
		if fallback == nil {
			fallback = elt
			continue
		}
		releaseElement(elt)
	}
	return fallback, nil
}

func walkWindowsNode(walker *uia.TreeWalker, elt *uia.Element, maxDepth int, budget *int) (*computeruse.AccessibilityNode, error) {
	if *budget <= 0 || elt == nil {
		return nil, nil
	}
	*budget--

	node := windowsNodeFromElement(elt, true)
	if maxDepth == 0 {
		return node, nil
	}

	nextDepth := maxDepth
	if maxDepth > 0 {
		nextDepth = maxDepth - 1
	}

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

func windowsElementRole(elt *uia.Element) string {
	localized := strings.TrimSpace(windowsStringProperty(elt.CurrentLocalizedControlType))
	if localized != "" {
		return localized
	}
	if controlType, err := elt.CurrentControlType(); err == nil {
		if name := strings.TrimSpace(uia.ControlTypeNames[controlType]); name != "" {
			return normalizeWindowsControlTypeName(name)
		}
		return fmt.Sprintf("control_type_%d", controlType)
	}
	return "unknown"
}

func windowsElementDescription(elt *uia.Element) string {
	if help := strings.TrimSpace(windowsStringProperty(elt.CurrentHelpText)); help != "" {
		return help
	}
	return strings.TrimSpace(windowsStringProperty(elt.CurrentLocalizedControlType))
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

func normalizeWindowsControlTypeName(name string) string {
	name = strings.TrimSuffix(name, "Control")
	var out strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			out.WriteByte(' ')
		}
		out.WriteRune(r)
	}
	return strings.ToLower(out.String())
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
