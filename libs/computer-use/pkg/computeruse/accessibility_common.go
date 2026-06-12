// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Platform-independent accessibility semantics shared by the Linux (AT-SPI,
// accessibility.go) and Windows (UI Automation, accessibility_windows.go)
// implementations. Exactly one platform file compiles into any given build,
// so everything the daemon's wire contract depends on lives here, untagged:
// the sentinel error strings (matched by exact leading text in
// apps/daemon/pkg/toolbox/computeruse/accessibility.go after net/rpc
// flattens errors), scope parsing, walk/find limits, depth semantics, and
// the find filter matcher. Keeping a single definition makes one-sided
// drift a compile error instead of a silent HTTP status or parity break.

package computeruse

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ---------------------------------------------------------------------------
// Sentinel errors (wire-translated by the daemon layer).
// ---------------------------------------------------------------------------

var (
	ErrA11yUnavailable    = errors.New("accessibility bus not reachable")
	ErrNoAccessibleRoot   = errors.New("no accessible root for focused window")
	ErrNodeNotFound       = errors.New("accessibility node not found")
	ErrActionNotSupported = errors.New("action not supported by node")
	ErrInvalidScope       = errors.New("invalid accessibility scope")
	ErrInvalidRequest     = errors.New("invalid accessibility request")
)

// ---------------------------------------------------------------------------
// Scope.
// ---------------------------------------------------------------------------

type A11yScope string

const (
	A11yScopeFocused A11yScope = "focused"
	A11yScopePID     A11yScope = "pid"
	A11yScopeAll     A11yScope = "all"
)

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

// ---------------------------------------------------------------------------
// Walk and find limits.
// ---------------------------------------------------------------------------

const (
	// a11yWalkBudget hard-caps nodes visited during a single tree walk or
	// find. Tuneable if real workloads demand it; sized to survive a full
	// desktop dump.
	a11yWalkBudget = 20000

	findDefaultLimit = 500
	findCeilingLimit = 5000
)

// normalizeFindLimit applies the find limit defaults: non-positive limits
// fall back to the default, oversized limits are capped.
func normalizeFindLimit(limit int) int {
	if limit <= 0 {
		return findDefaultLimit
	}
	if limit > findCeilingLimit {
		return findCeilingLimit
	}
	return limit
}

// Depth semantics shared by both tree walkers: 0 visits only the current
// node, positive values bound descent, negative values are unbounded (the
// daemon defaults an absent maxDepth query parameter to -1).
func a11yDepthAllowsDescent(maxDepth int) bool {
	return maxDepth != 0
}

func a11yNextDepth(maxDepth int) int {
	if maxDepth > 0 {
		return maxDepth - 1
	}
	return maxDepth
}

// ---------------------------------------------------------------------------
// Filter semantics (pure, unit-testable).
// ---------------------------------------------------------------------------

// buildA11yMatcher returns the find filter predicate over the plain
// (role, name, states) tuple of a node, implementing the semantics
// documented in the API spec: all fields are AND-ed, empty fields are
// ignored, roles compare case-insensitively, names case-sensitively. Regex
// compilation failures and unknown nameMatch modes are surfaced to the
// caller as ErrInvalidRequest. The platform files adapt their node types to
// the tuple in one line.
func buildA11yMatcher(role, name, nameMatch string, states []string) (func(nodeRole, nodeName string, nodeStates []string) bool, error) {
	if nameMatch == "" {
		nameMatch = "substring"
	}
	if nameMatch != "exact" && nameMatch != "substring" && nameMatch != "regex" {
		return nil, fmt.Errorf("%w: unknown nameMatch mode %q, want exact|substring|regex", ErrInvalidRequest, nameMatch)
	}

	var nameRe *regexp.Regexp
	if name != "" && nameMatch == "regex" {
		re, err := regexp.Compile(name)
		if err != nil {
			return nil, fmt.Errorf("%w: invalid regex for name filter: %v", ErrInvalidRequest, err)
		}
		nameRe = re
	}

	wantRole := strings.ToLower(role)
	wantStates := append([]string(nil), states...)

	return func(nodeRole, nodeName string, nodeStates []string) bool {
		if wantRole != "" && strings.ToLower(nodeRole) != wantRole {
			return false
		}
		if name != "" {
			switch nameMatch {
			case "exact":
				if nodeName != name {
					return false
				}
			case "substring":
				if !strings.Contains(nodeName, name) {
					return false
				}
			case "regex":
				if nameRe == nil || !nameRe.MatchString(nodeName) {
					return false
				}
			}
		}
		for _, want := range wantStates {
			if !containsStr(nodeStates, want) {
				return false
			}
		}
		return true
	}, nil
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
