// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// AT-SPI is spoken directly over D-Bus via godbus/dbus/v5. AT-SPI's public
// API is a D-Bus protocol; this file walks the registry, reads node state,
// and invokes actions through the same bus the desktop uses for a11y.

package computeruse

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/godbus/dbus/v5"
)

// ---------------------------------------------------------------------------
// Public plugin-internal types (mirror the wire shape the daemon will expose).
// ---------------------------------------------------------------------------

type A11yScope string

const (
	A11yScopeFocused A11yScope = "focused"
	A11yScopePID     A11yScope = "pid"
	A11yScopeAll     A11yScope = "all"
)

type A11yBounds struct {
	X      int
	Y      int
	Width  int
	Height int
}

type A11yNode struct {
	ID          string // "<bus-name>:<object-path>"
	Role        string
	Name        string
	Description string
	Bounds      A11yBounds
	States      []string
	Actions     []string    // names of supported Action interface actions
	Children    []*A11yNode // nil when flattened (find result)
}

type A11yFilter struct {
	Role      string
	Name      string
	NameMatch string // "exact" | "substring" | "regex" — default "substring"
	States    []string
}

// ---------------------------------------------------------------------------
// Sentinel errors (wire-translated by the daemon layer).
// ---------------------------------------------------------------------------

var (
	ErrA11yUnavailable    = errors.New("accessibility bus not reachable")
	ErrNoAccessibleRoot   = errors.New("no accessible root for focused window")
	ErrNodeNotFound       = errors.New("accessibility node not found")
	ErrActionNotSupported = errors.New("action not supported by node")
	ErrInvalidScope       = errors.New("invalid accessibility scope")
)

// ---------------------------------------------------------------------------
// AT-SPI protocol constants.
// ---------------------------------------------------------------------------

const (
	atspiRegistryBus  = "org.a11y.atspi.Registry"
	atspiRootPath     = dbus.ObjectPath("/org/a11y/atspi/accessible/root")
	atspiBusServiceBN = "org.a11y.Bus"
	atspiBusServiceOP = dbus.ObjectPath("/org/a11y/bus")

	ifaceAccessible   = "org.a11y.atspi.Accessible"
	ifaceComponent    = "org.a11y.atspi.Component"
	ifaceAction       = "org.a11y.atspi.Action"
	ifaceEditableText = "org.a11y.atspi.EditableText"
	ifaceValue        = "org.a11y.atspi.Value"
	ifaceApplication  = "org.a11y.atspi.Application"

	coordTypeScreen uint32 = 0 // AT-SPI CoordType_SCREEN

	// Hard cap on nodes visited during a single tree walk. Tuneable if real
	// workloads demand it; sized to survive a full xfce4 desktop dump.
	walkBudget = 20000

	// State bit indices of interest. Covers the full AT-SPI StateType enum
	// (AT-SPI uses 2x uint32 = 64 bit-slots; ~44 are defined today).
	stateActive = 1 // used for focus-scoped root resolution
)

// atspiStateNames is the AT-SPI StateType enum ordered by bit index. Names
// use the underscored lowercase form matching the underlying C enum (e.g.
// MULTI_LINE → "multi_line") for predictable filtering. The ordering is
// stable ABI; new states are only ever appended at the end.
//
// Source: at-spi2-core/xml/Registry.xml
// https://gitlab.gnome.org/GNOME/at-spi2-core/-/blob/master/xml/Registry.xml
var atspiStateNames = []string{
	"invalid",                 // 0
	"active",                  // 1
	"armed",                   // 2
	"busy",                    // 3
	"checked",                 // 4
	"collapsed",               // 5
	"defunct",                 // 6
	"editable",                // 7
	"enabled",                 // 8
	"expandable",              // 9
	"expanded",                // 10
	"focusable",               // 11
	"focused",                 // 12
	"has_tooltip",             // 13
	"horizontal",              // 14
	"iconified",               // 15
	"modal",                   // 16
	"multi_line",              // 17
	"multiselectable",         // 18
	"opaque",                  // 19
	"pressed",                 // 20
	"resizable",               // 21
	"selectable",              // 22
	"selected",                // 23
	"sensitive",               // 24
	"showing",                 // 25
	"single_line",             // 26
	"stale",                   // 27
	"transient",               // 28
	"vertical",                // 29
	"visible",                 // 30
	"manages_descendants",     // 31
	"indeterminate",           // 32
	"required",                // 33
	"truncated",               // 34
	"animated",                // 35
	"invalid_entry",           // 36
	"supports_autocompletion", // 37
	"selectable_text",         // 38
	"is_default",              // 39
	"visited",                 // 40
	"checkable",               // 41
	"has_popup",               // 42
	"read_only",               // 43
}

// accessibleRef matches the AT-SPI (so) struct — a (bus-name, object-path) pair.
type accessibleRef struct {
	Sender string
	Path   dbus.ObjectPath
}

// ---------------------------------------------------------------------------
// Bootstrap: lazily establish and cache a connection to the AT-SPI bus.
// ---------------------------------------------------------------------------

// connectA11y returns a live connection to the AT-SPI bus, dialling it on
// first call and caching it thereafter. The AT-SPI bus is a separate D-Bus
// instance whose address is discovered via org.a11y.Bus.GetAddress() on the
// session bus.
func (c *ComputerUse) connectA11y() (*dbus.Conn, error) {
	c.atspiMu.Lock()
	defer c.atspiMu.Unlock()

	if c.atspiConn != nil && c.atspiConn.Connected() {
		return c.atspiConn, nil
	}
	// Drop a stale reference so we don't accidentally hand it out again.
	c.atspiConn = nil

	sess, err := dbus.SessionBus()
	if err != nil {
		return nil, fmt.Errorf("%w: session bus: %v", ErrA11yUnavailable, err)
	}

	var addr string
	busObj := sess.Object(atspiBusServiceBN, atspiBusServiceOP)
	if err := busObj.Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		return nil, fmt.Errorf("%w: GetAddress: %v", ErrA11yUnavailable, err)
	}
	if addr == "" {
		return nil, fmt.Errorf("%w: GetAddress returned empty address", ErrA11yUnavailable)
	}

	conn, err := dbus.Connect(addr)
	if err != nil {
		return nil, fmt.Errorf("%w: dial %s: %v", ErrA11yUnavailable, addr, err)
	}

	c.atspiConn = conn
	return conn, nil
}

// ---------------------------------------------------------------------------
// Node ID helpers.
// ---------------------------------------------------------------------------

// makeNodeID produces the stable "<bus-name>:<object-path>" identifier the
// API returns. Object paths always start with '/' and bus names never contain
// '/', which makes parseNodeID unambiguous regardless of whether the bus name
// is a unique ":1.42" or well-known "org.foo" form.
func makeNodeID(sender string, path dbus.ObjectPath) string {
	return sender + ":" + string(path)
}

// parseNodeID is the inverse. The separator is the ':' immediately preceding
// the first '/' (the start of the object path).
func parseNodeID(id string) (string, dbus.ObjectPath, error) {
	slash := strings.Index(id, "/")
	if slash <= 1 {
		return "", "", fmt.Errorf("invalid node id %q: missing object path", id)
	}
	if id[slash-1] != ':' {
		return "", "", fmt.Errorf("invalid node id %q: missing ':' separator before object path", id)
	}
	sender := id[:slash-1]
	path := dbus.ObjectPath(id[slash:])
	if sender == "" {
		return "", "", fmt.Errorf("invalid node id %q: empty bus name", id)
	}
	if !path.IsValid() {
		return "", "", fmt.Errorf("invalid node id %q: malformed object path", id)
	}
	return sender, path, nil
}

// ---------------------------------------------------------------------------
// State bitmap decode.
// ---------------------------------------------------------------------------

// stateBitsToNames maps an AT-SPI StateSet (2x uint32) to the canonical
// lowercase state names from atspiStateNames. Unknown bits are ignored.
func stateBitsToNames(bits []uint32) []string {
	if len(bits) == 0 {
		return nil
	}
	var out []string
	for i, name := range atspiStateNames {
		word := i / 32
		if word >= len(bits) {
			break
		}
		if bits[word]&(1<<uint(i%32)) != 0 {
			out = append(out, name)
		}
	}
	return out
}

func stateBitSet(bits []uint32, bit int) bool {
	word := bit / 32
	if word >= len(bits) {
		return false
	}
	return bits[word]&(1<<uint(bit%32)) != 0
}

// ---------------------------------------------------------------------------
// D-Bus error classification.
// ---------------------------------------------------------------------------

// dbusErrorName extracts the fully-qualified D-Bus error name (e.g.
// "org.freedesktop.DBus.Error.UnknownMethod") from an error returned by a
// godbus call, walking through Unwrap chains. Returns "" for non-dbus errors.
func dbusErrorName(err error) string {
	for e := err; e != nil; {
		switch v := e.(type) {
		case *dbus.Error:
			return v.Name
		case dbus.Error:
			return v.Name
		}
		u := errors.Unwrap(e)
		if u == nil || u == e {
			return ""
		}
		e = u
	}
	return ""
}

// classifyDbusError translates a raw godbus error into one of our sentinel
// errors where the mapping is unambiguous, or returns it as-is otherwise.
func classifyDbusError(err error) error {
	if err == nil {
		return nil
	}
	switch dbusErrorName(err) {
	case "org.freedesktop.DBus.Error.UnknownMethod",
		"org.freedesktop.DBus.Error.UnknownInterface",
		"org.freedesktop.DBus.Error.UnknownProperty",
		"org.freedesktop.DBus.Error.NotSupported":
		return ErrActionNotSupported
	case "org.freedesktop.DBus.Error.ServiceUnknown",
		"org.freedesktop.DBus.Error.NameHasNoOwner",
		"org.freedesktop.DBus.Error.UnknownObject":
		return ErrNodeNotFound
	}
	return err
}

// ---------------------------------------------------------------------------
// Low-level AT-SPI accessors.
// ---------------------------------------------------------------------------

// getStringProp reads a single string property from a D-Bus object. An
// UnknownProperty error is silently squashed to "" so we don't blow up a walk
// when a node declines to implement an optional property.
func getStringProp(obj dbus.BusObject, prop string) string {
	v, err := obj.GetProperty(prop)
	if err != nil {
		return ""
	}
	if s, ok := v.Value().(string); ok {
		return s
	}
	return ""
}

func getInt32Prop(obj dbus.BusObject, prop string) (int32, bool) {
	v, err := obj.GetProperty(prop)
	if err != nil {
		return 0, false
	}
	if i, ok := v.Value().(int32); ok {
		return i, true
	}
	return 0, false
}

func getChildren(conn *dbus.Conn, sender string, path dbus.ObjectPath) ([]accessibleRef, error) {
	obj := conn.Object(sender, path)
	var refs []accessibleRef
	if err := obj.Call(ifaceAccessible+".GetChildren", 0).Store(&refs); err != nil {
		return nil, classifyDbusError(err)
	}
	return refs, nil
}

func getState(obj dbus.BusObject) ([]uint32, error) {
	var bits []uint32
	if err := obj.Call(ifaceAccessible+".GetState", 0).Store(&bits); err != nil {
		return nil, classifyDbusError(err)
	}
	return bits, nil
}

func getInterfaces(obj dbus.BusObject) []string {
	var ifaces []string
	if err := obj.Call(ifaceAccessible+".GetInterfaces", 0).Store(&ifaces); err != nil {
		return nil
	}
	return ifaces
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func getExtents(obj dbus.BusObject) (A11yBounds, bool) {
	var ext struct {
		X, Y, W, H int32
	}
	if err := obj.Call(ifaceComponent+".GetExtents", 0, coordTypeScreen).Store(&ext); err != nil {
		return A11yBounds{}, false
	}
	return A11yBounds{X: int(ext.X), Y: int(ext.Y), Width: int(ext.W), Height: int(ext.H)}, true
}

func getActionNames(obj dbus.BusObject) []string {
	var actions []struct {
		Name, Description, KeyBinding string
	}
	if err := obj.Call(ifaceAction+".GetActions", 0).Store(&actions); err != nil {
		return nil
	}
	out := make([]string, 0, len(actions))
	for _, a := range actions {
		out = append(out, a.Name)
	}
	return out
}

// fetchNodeMeta reads every per-node field other than children. interfaces is
// returned so callers can decide whether to try optional sub-interface calls.
func fetchNodeMeta(conn *dbus.Conn, sender string, path dbus.ObjectPath) (*A11yNode, []string, error) {
	obj := conn.Object(sender, path)

	var role string
	if err := obj.Call(ifaceAccessible+".GetRoleName", 0).Store(&role); err != nil {
		return nil, nil, classifyDbusError(err)
	}

	ifaces := getInterfaces(obj)

	node := &A11yNode{
		ID:          makeNodeID(sender, path),
		Role:        role,
		Name:        getStringProp(obj, ifaceAccessible+".Name"),
		Description: getStringProp(obj, ifaceAccessible+".Description"),
	}

	if bits, err := getState(obj); err == nil {
		node.States = stateBitsToNames(bits)
	}

	if containsStr(ifaces, ifaceComponent) {
		if b, ok := getExtents(obj); ok {
			node.Bounds = b
		}
	}

	if containsStr(ifaces, ifaceAction) {
		node.Actions = getActionNames(obj)
	}

	return node, ifaces, nil
}

// ---------------------------------------------------------------------------
// Scope resolution.
// ---------------------------------------------------------------------------

// resolveScopeRoot picks the starting accessibleRef for a walk based on scope.
func (c *ComputerUse) resolveScopeRoot(conn *dbus.Conn, scope A11yScope, pid int) (accessibleRef, error) {
	switch scope {
	case A11yScopeAll, "":
		return accessibleRef{Sender: atspiRegistryBus, Path: atspiRootPath}, nil
	case A11yScopeFocused:
		return c.getFocusedAppRoot(conn)
	case A11yScopePID:
		return c.getAppRootByPID(conn, pid)
	default:
		return accessibleRef{}, fmt.Errorf("unknown scope %q", scope)
	}
}

// getFocusedAppRoot walks registry children and returns the app whose
// top-level window descendant has the ACTIVE state set. Falls back to the
// first registered app when no ACTIVE window is found.
func (c *ComputerUse) getFocusedAppRoot(conn *dbus.Conn) (accessibleRef, error) {
	apps, err := getChildren(conn, atspiRegistryBus, atspiRootPath)
	if err != nil {
		return accessibleRef{}, err
	}
	if len(apps) == 0 {
		return accessibleRef{}, ErrNoAccessibleRoot
	}
	for _, app := range apps {
		if appHasActiveWindow(conn, app) {
			return app, nil
		}
	}
	return apps[0], nil
}

func appHasActiveWindow(conn *dbus.Conn, app accessibleRef) bool {
	windows, err := getChildren(conn, app.Sender, app.Path)
	if err != nil {
		return false
	}
	for _, w := range windows {
		obj := conn.Object(w.Sender, w.Path)
		bits, err := getState(obj)
		if err != nil {
			continue
		}
		if stateBitSet(bits, stateActive) {
			return true
		}
	}
	return false
}

// getAppRootByPID matches by the Application.Id property, which AT-SPI wires
// to the app's OS process id.
func (c *ComputerUse) getAppRootByPID(conn *dbus.Conn, pid int) (accessibleRef, error) {
	if pid <= 0 {
		return accessibleRef{}, fmt.Errorf("%w: pid must be positive", ErrNoAccessibleRoot)
	}
	apps, err := getChildren(conn, atspiRegistryBus, atspiRootPath)
	if err != nil {
		return accessibleRef{}, err
	}
	for _, app := range apps {
		obj := conn.Object(app.Sender, app.Path)
		if id, ok := getInt32Prop(obj, ifaceApplication+".Id"); ok {
			if int(id) == pid {
				return app, nil
			}
		}
	}
	return accessibleRef{}, ErrNoAccessibleRoot
}

// ---------------------------------------------------------------------------
// Tree walk.
// ---------------------------------------------------------------------------

// getAccessibilityTree walks the AT-SPI tree under the requested scope and
// returns the subtree as an A11yNode. Returns truncated=true when the global
// node budget is exhausted.
func (c *ComputerUse) getAccessibilityTree(scope A11yScope, pid int, maxDepth int) (*A11yNode, bool, error) {
	conn, err := c.connectA11y()
	if err != nil {
		return nil, false, err
	}
	root, err := c.resolveScopeRoot(conn, scope, pid)
	if err != nil {
		return nil, false, err
	}
	budget := walkBudget
	node, err := walkNode(conn, root.Sender, root.Path, maxDepth, &budget)
	if err != nil {
		return nil, false, err
	}
	truncated := budget <= 0
	return node, truncated, nil
}

// walkNode recursively materialises an A11yNode tree rooted at (sender, path).
// maxDepth < 0 means unbounded. budget is decremented on every visited node;
// when it hits zero, descent stops and the caller infers truncation.
func walkNode(conn *dbus.Conn, sender string, path dbus.ObjectPath, maxDepth int, budget *int) (*A11yNode, error) {
	if *budget <= 0 {
		return nil, nil
	}
	*budget--

	node, _, err := fetchNodeMeta(conn, sender, path)
	if err != nil {
		return nil, err
	}

	if maxDepth == 0 {
		return node, nil
	}

	refs, err := getChildren(conn, sender, path)
	if err != nil {
		// Missing children list isn't fatal for the walk; log implicitly by
		// returning the node without children.
		return node, nil
	}

	nextDepth := maxDepth
	if maxDepth > 0 {
		nextDepth = maxDepth - 1
	}

	for _, r := range refs {
		if *budget <= 0 {
			break
		}
		child, err := walkNode(conn, r.Sender, r.Path, nextDepth, budget)
		if err != nil {
			// A child disappearing mid-walk shouldn't abort the walk.
			if errors.Is(err, ErrNodeNotFound) {
				continue
			}
			return nil, err
		}
		if child != nil {
			node.Children = append(node.Children, child)
		}
	}
	return node, nil
}

// ---------------------------------------------------------------------------
// Find.
// ---------------------------------------------------------------------------

const (
	findDefaultLimit = 500
	findCeilingLimit = 5000
)

// findAccessibilityNodes walks the same scope as getAccessibilityTree but
// returns a flat list of matches. Children of returned nodes are always nil.
func (c *ComputerUse) findAccessibilityNodes(scope A11yScope, pid int, filter A11yFilter, limit int) ([]*A11yNode, bool, error) {
	if limit <= 0 {
		limit = findDefaultLimit
	}
	if limit > findCeilingLimit {
		limit = findCeilingLimit
	}

	matcher, err := buildFilterMatcher(filter)
	if err != nil {
		return nil, false, err
	}

	conn, err := c.connectA11y()
	if err != nil {
		return nil, false, err
	}
	root, err := c.resolveScopeRoot(conn, scope, pid)
	if err != nil {
		return nil, false, err
	}

	budget := walkBudget
	matches := make([]*A11yNode, 0, 16)
	truncated := findWalk(conn, root.Sender, root.Path, matcher, &matches, limit, &budget)
	return matches, truncated, nil
}

// findWalk returns true when the walk was truncated by either the node budget
// or the result limit.
func findWalk(conn *dbus.Conn, sender string, path dbus.ObjectPath, match func(*A11yNode) bool, out *[]*A11yNode, limit int, budget *int) bool {
	if *budget <= 0 {
		return true
	}
	*budget--

	node, _, err := fetchNodeMeta(conn, sender, path)
	if err != nil {
		// Can't read this node — treat as invisible rather than failing the walk.
		return false
	}

	if match(node) {
		// Children are dropped for find results per spec.
		flat := *node
		flat.Children = nil
		*out = append(*out, &flat)
		if len(*out) >= limit {
			return true
		}
	}

	refs, err := getChildren(conn, sender, path)
	if err != nil {
		return false
	}
	for _, r := range refs {
		if *budget <= 0 {
			return true
		}
		if findWalk(conn, r.Sender, r.Path, match, out, limit, budget) {
			return true
		}
	}
	return *budget <= 0
}

// ---------------------------------------------------------------------------
// Filter logic (pure, unit-testable).
// ---------------------------------------------------------------------------

// buildFilterMatcher returns a predicate that implements the filter semantics
// documented in the API spec. All fields are AND-ed; empty fields are ignored.
// Regex compilation failures are surfaced to the caller.
func buildFilterMatcher(f A11yFilter) (func(*A11yNode) bool, error) {
	var nameRe *regexp.Regexp
	nameMatch := f.NameMatch
	if nameMatch == "" {
		nameMatch = "substring"
	}
	if f.Name != "" && nameMatch == "regex" {
		re, err := regexp.Compile(f.Name)
		if err != nil {
			return nil, fmt.Errorf("invalid regex for name filter: %w", err)
		}
		nameRe = re
	}
	if f.Name != "" && nameMatch != "exact" && nameMatch != "substring" && nameMatch != "regex" {
		return nil, fmt.Errorf("unknown nameMatch mode %q: want exact|substring|regex", nameMatch)
	}

	role := strings.ToLower(f.Role)
	states := append([]string(nil), f.States...)

	return func(n *A11yNode) bool {
		if role != "" && strings.ToLower(n.Role) != role {
			return false
		}
		if f.Name != "" {
			switch nameMatch {
			case "exact":
				if n.Name != f.Name {
					return false
				}
			case "substring":
				if !strings.Contains(n.Name, f.Name) {
					return false
				}
			case "regex":
				if nameRe == nil || !nameRe.MatchString(n.Name) {
					return false
				}
			}
		}
		for _, want := range states {
			if !containsStr(n.States, want) {
				return false
			}
		}
		return true
	}, nil
}

// ---------------------------------------------------------------------------
// Actions.
// ---------------------------------------------------------------------------

func (c *ComputerUse) focusAccessibilityNode(id string) error {
	sender, path, err := parseNodeID(id)
	if err != nil {
		return err
	}
	conn, err := c.connectA11y()
	if err != nil {
		return err
	}
	obj := conn.Object(sender, path)

	var ok bool
	if err := obj.Call(ifaceComponent+".GrabFocus", 0).Store(&ok); err != nil {
		return classifyDbusError(err)
	}
	if !ok {
		// AT-SPI reports false when the node refuses focus — e.g. disabled or
		// non-focusable. Surface as ErrActionNotSupported since the agent
		// can't meaningfully retry without picking a different node.
		return ErrActionNotSupported
	}
	return nil
}

func (c *ComputerUse) invokeAccessibilityNode(id string, action string) error {
	sender, path, err := parseNodeID(id)
	if err != nil {
		return err
	}
	conn, err := c.connectA11y()
	if err != nil {
		return err
	}
	obj := conn.Object(sender, path)

	var actions []struct {
		Name, Description, KeyBinding string
	}
	if err := obj.Call(ifaceAction+".GetActions", 0).Store(&actions); err != nil {
		return classifyDbusError(err)
	}

	idx := -1
	if action == "" {
		if len(actions) == 0 {
			return ErrActionNotSupported
		}
		idx = 0
	} else {
		for i, a := range actions {
			if strings.EqualFold(a.Name, action) {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		return ErrActionNotSupported
	}

	var ok bool
	if err := obj.Call(ifaceAction+".DoAction", 0, int32(idx)).Store(&ok); err != nil {
		return classifyDbusError(err)
	}
	if !ok {
		return fmt.Errorf("DoAction(%d) returned false", idx)
	}
	return nil
}

func (c *ComputerUse) setAccessibilityNodeValue(id string, value string) error {
	sender, path, err := parseNodeID(id)
	if err != nil {
		return err
	}
	conn, err := c.connectA11y()
	if err != nil {
		return err
	}
	obj := conn.Object(sender, path)

	// 1) Prefer EditableText: the common case for text entries / terminals /
	//    search fields. SetTextContents returns bool; most impls return false
	//    yet still update, so we treat a successful call as success.
	var ok bool
	etErr := obj.Call(ifaceEditableText+".SetTextContents", 0, value).Store(&ok)
	if etErr == nil {
		return nil
	}
	etTranslated := classifyDbusError(etErr)
	if etTranslated == ErrNodeNotFound {
		return etTranslated
	}
	if etTranslated != ErrActionNotSupported {
		// Real failure from a node that does implement EditableText — don't
		// silently fall through to Value with a different semantic.
		return etTranslated
	}

	// 2) Fallback to Value.CurrentValue for sliders / spin buttons.
	f, parseErr := strconv.ParseFloat(value, 64)
	if parseErr != nil {
		return fmt.Errorf("%w: node has no EditableText and %q is not numeric: %v",
			ErrActionNotSupported, value, parseErr)
	}
	if err := obj.SetProperty(ifaceValue+".CurrentValue", f); err != nil {
		return classifyDbusError(err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Interface-facing wrappers.
//
// These adapt the lowercase plugin-internal methods (positional args,
// plugin-owned A11yNode/A11yBounds) to the shape the daemon's IComputerUse
// interface expects (request structs, wire types from
// github.com/daytonaio/daemon/pkg/toolbox/computeruse). Error translation to
// HTTP status codes happens in the handler layer — wrappers propagate
// sentinel errors unchanged so the handler can use errors.Is.
// ---------------------------------------------------------------------------

// parseWireScope validates a scope string coming over the wire. The empty
// string is treated as the default ("focused"). Returns ErrInvalidScope
// wrapped with a descriptive message on unknown scopes so the handler can map
// to 400 Bad Request.
func parseWireScope(s string) (A11yScope, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "focused":
		return A11yScopeFocused, nil
	case "pid":
		return A11yScopePID, nil
	case "all":
		return A11yScopeAll, nil
	default:
		return "", fmt.Errorf("%w: got %q, expected focused|pid|all", ErrInvalidScope, s)
	}
}

// toWireBounds converts an internal A11yBounds to the daemon wire shape.
func toWireBounds(b A11yBounds) computeruse.AccessibilityBounds {
	return computeruse.AccessibilityBounds{
		X:      b.X,
		Y:      b.Y,
		Width:  b.Width,
		Height: b.Height,
	}
}

// toWireNode recursively converts an internal *A11yNode tree into the wire
// shape. Nil input yields nil so callers can distinguish "no node" from
// "empty node". Children are converted to a pointer-slice to match the wire
// type definition (pointer-slice is a swag workaround for self-referencing
// types: []AccessibilityNode confuses swag's @name resolution for recursive
// fields).
func toWireNode(n *A11yNode) *computeruse.AccessibilityNode {
	if n == nil {
		return nil
	}
	out := &computeruse.AccessibilityNode{
		ID:          n.ID,
		Role:        n.Role,
		Name:        n.Name,
		Description: n.Description,
		Bounds:      toWireBounds(n.Bounds),
		States:      n.States,
		Actions:     n.Actions,
	}
	if len(n.Children) > 0 {
		out.Children = make([]*computeruse.AccessibilityNode, 0, len(n.Children))
		for _, child := range n.Children {
			if wc := toWireNode(child); wc != nil {
				out.Children = append(out.Children, wc)
			}
		}
	}
	return out
}

// toWireNodes converts a flat slice of internal nodes (children fields
// discarded by findAccessibilityNodes) into a value-slice of wire nodes.
func toWireNodes(ns []*A11yNode) []computeruse.AccessibilityNode {
	out := make([]computeruse.AccessibilityNode, 0, len(ns))
	for _, n := range ns {
		if wn := toWireNode(n); wn != nil {
			out = append(out, *wn)
		}
	}
	return out
}

// GetAccessibilityTree is the IComputerUse-facing entry point for
// /computeruse/a11y/tree. The req.MaxDepth convention matches the HTTP
// contract: a zero MaxDepth sent with Scope="" is treated as "focused, root
// only"; callers that want unbounded descent should send -1.
func (c *ComputerUse) GetAccessibilityTree(req *computeruse.GetAccessibilityTreeRequest) (*computeruse.AccessibilityTreeResponse, error) {
	if req == nil {
		req = &computeruse.GetAccessibilityTreeRequest{}
	}
	scope, err := parseWireScope(req.Scope)
	if err != nil {
		return nil, err
	}
	node, truncated, err := c.getAccessibilityTree(scope, req.PID, req.MaxDepth)
	if err != nil {
		return nil, err
	}
	resp := &computeruse.AccessibilityTreeResponse{Truncated: truncated}
	if wire := toWireNode(node); wire != nil {
		resp.Root = *wire
	}
	return resp, nil
}

// FindAccessibilityNodes is the IComputerUse-facing entry point for
// /computeruse/a11y/find.
func (c *ComputerUse) FindAccessibilityNodes(req *computeruse.FindAccessibilityNodesRequest) (*computeruse.AccessibilityNodesResponse, error) {
	if req == nil {
		req = &computeruse.FindAccessibilityNodesRequest{}
	}
	scope, err := parseWireScope(req.Scope)
	if err != nil {
		return nil, err
	}
	filter := A11yFilter{
		Role:      req.Role,
		Name:      req.Name,
		NameMatch: req.NameMatch,
		States:    req.States,
	}
	matches, truncated, err := c.findAccessibilityNodes(scope, req.PID, filter, req.Limit)
	if err != nil {
		return nil, err
	}
	return &computeruse.AccessibilityNodesResponse{
		Matches:   toWireNodes(matches),
		Truncated: truncated,
	}, nil
}

// FocusAccessibilityNode is the IComputerUse-facing entry point for
// /computeruse/a11y/node/focus.
func (c *ComputerUse) FocusAccessibilityNode(req *computeruse.AccessibilityNodeRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := c.focusAccessibilityNode(req.ID); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

// InvokeAccessibilityNode is the IComputerUse-facing entry point for
// /computeruse/a11y/node/invoke.
func (c *ComputerUse) InvokeAccessibilityNode(req *computeruse.AccessibilityInvokeRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := c.invokeAccessibilityNode(req.ID, req.Action); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

// SetAccessibilityNodeValue is the IComputerUse-facing entry point for
// /computeruse/a11y/node/value.
func (c *ComputerUse) SetAccessibilityNodeValue(req *computeruse.AccessibilitySetValueRequest) (*computeruse.Empty, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if err := c.setAccessibilityNodeValue(req.ID, req.Value); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}
