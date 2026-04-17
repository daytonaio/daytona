// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Daemon-side HTTP handlers for the AT-SPI accessibility API. The plugin's
// interface-facing wrappers (libs/computer-use/pkg/computeruse/accessibility.go)
// return raw sentinel errors; this file maps them to HTTP status codes. Only
// the a11y handlers use this typed-error translation — existing computer-use
// handlers keep their inline behaviour.

package computeruse

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Sentinel error messages mirrored from the plugin's
// libs/computer-use/pkg/computeruse/accessibility.go. They are matched here by
// substring rather than errors.Is because the plugin package imports the
// daemon's wire types (for the IComputerUse interface shape), so a direct
// import of the plugin package from the daemon would create a cycle.
//
// Errors from the plugin reach this file via two paths:
//   - In-process (same binary compiled for tests): the original sentinel
//     object is preserved, so errors.Is would also work.
//   - go-plugin RPC: net/rpc flattens errors to plain strings (error value is
//     *rpc.ServerError with Error() == original message), which defeats
//     errors.Is but keeps substring matching reliable.
//
// Substring matching is robust here because the plugin always wraps its
// sentinels with fmt.Errorf("%w: ...context", ErrX, ...), so the canonical
// message text is always a substring of the final error.
const (
	a11yMsgUnavailable        = "accessibility bus not reachable"
	a11yMsgNoAccessibleRoot   = "no accessible root for focused window"
	a11yMsgNodeNotFound       = "accessibility node not found"
	a11yMsgActionNotSupported = "action not supported by node"
	a11yMsgInvalidScope       = "invalid accessibility scope"
)

// writeA11yError maps the plugin sentinel errors to HTTP responses. Unknown
// errors fall through to 500. The translation is kept here (not baked into
// the plugin) so the plugin package stays transport-agnostic.
//
// We intentionally write responses directly via c.JSON instead of emitting
// typed common_errors through c.Error + the shared middleware — matching the
// inline style used by every other WrapXHandler in this package. Direct JSON
// keeps the diff minimal; retrofitting the rest of the package is a separate
// cleanup.
func writeA11yError(c *gin.Context, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, a11yMsgUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": msg,
			"code":  "A11Y_UNAVAILABLE",
		})
	case strings.Contains(msg, a11yMsgNodeNotFound),
		strings.Contains(msg, a11yMsgNoAccessibleRoot):
		c.JSON(http.StatusNotFound, gin.H{"error": msg})
	case strings.Contains(msg, a11yMsgActionNotSupported),
		strings.Contains(msg, a11yMsgInvalidScope):
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
	}
}

// GetAccessibilityTree godoc
//
//	@Summary		Get accessibility tree
//	@Description	Fetch the AT-SPI accessibility tree for the focused application, a specific PID, or all registered applications.
//	@Tags			computer-use
//	@Produce		json
//	@Param			scope		query		string	false	"Scope: focused | pid | all (default: focused)"
//	@Param			pid			query		int		false	"Process ID when scope=pid"
//	@Param			maxDepth	query		int		false	"Max tree depth (-1 unbounded, 0 root only; default -1)"
//	@Success		200			{object}	AccessibilityTreeResponse
//	@Failure		400			{object}	map[string]string
//	@Failure		404			{object}	map[string]string
//	@Failure		503			{object}	map[string]string
//	@Router			/computeruse/a11y/tree [get]
//
//	@id				GetAccessibilityTree
func WrapGetAccessibilityTreeHandler(fn func(*GetAccessibilityTreeRequest) (*AccessibilityTreeResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Default to unbounded descent when the client omits maxDepth. We
		// detect the omission by binding against a sentinel pointer value
		// rather than relying on form binding zero-semantics (0 is a valid
		// maxDepth meaning "root only").
		req := &GetAccessibilityTreeRequest{MaxDepth: -1}
		// Bind scope / pid from query first; then explicitly re-parse
		// maxDepth so the "-1 default when absent" semantic survives binding.
		if err := c.ShouldBindQuery(req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if c.Query("maxDepth") == "" {
			req.MaxDepth = -1
		}
		if req.Scope == "" {
			req.Scope = "focused"
		}

		response, err := fn(req)
		if err != nil {
			writeA11yError(c, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// FindAccessibilityNodes godoc
//
//	@Summary		Find accessibility nodes
//	@Description	Search the AT-SPI tree for nodes matching a role/name/state filter and return a flat list.
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		FindAccessibilityNodesRequest	true	"Find request"
//	@Success		200		{object}	AccessibilityNodesResponse
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		503		{object}	map[string]string
//	@Router			/computeruse/a11y/find [post]
//
//	@id				FindAccessibilityNodes
func WrapFindAccessibilityNodesHandler(fn func(*FindAccessibilityNodesRequest) (*AccessibilityNodesResponse, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req FindAccessibilityNodesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := fn(&req)
		if err != nil {
			writeA11yError(c, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// FocusAccessibilityNode godoc
//
//	@Summary		Focus an accessibility node
//	@Description	Move keyboard focus to the AT-SPI node identified by id (bus-name:object-path).
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AccessibilityNodeRequest	true	"Node focus request"
//	@Success		200		{object}	Empty
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		503		{object}	map[string]string
//	@Router			/computeruse/a11y/node/focus [post]
//
//	@id				FocusAccessibilityNode
func WrapFocusAccessibilityNodeHandler(fn func(*AccessibilityNodeRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AccessibilityNodeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := fn(&req)
		if err != nil {
			writeA11yError(c, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// InvokeAccessibilityNode godoc
//
//	@Summary		Invoke an action on an accessibility node
//	@Description	Call an AT-SPI Action on the node. Leave action empty to invoke the node's primary (first) action.
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AccessibilityInvokeRequest	true	"Invoke request"
//	@Success		200		{object}	Empty
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		503		{object}	map[string]string
//	@Router			/computeruse/a11y/node/invoke [post]
//
//	@id				InvokeAccessibilityNode
func WrapInvokeAccessibilityNodeHandler(fn func(*AccessibilityInvokeRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AccessibilityInvokeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := fn(&req)
		if err != nil {
			writeA11yError(c, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}

// SetAccessibilityNodeValue godoc
//
//	@Summary		Set the value of an accessibility node
//	@Description	Write the given value to the node via EditableText.SetTextContents or, for numeric controls, Value.CurrentValue.
//	@Tags			computer-use
//	@Accept			json
//	@Produce		json
//	@Param			request	body		AccessibilitySetValueRequest	true	"Set value request"
//	@Success		200		{object}	Empty
//	@Failure		400		{object}	map[string]string
//	@Failure		404		{object}	map[string]string
//	@Failure		503		{object}	map[string]string
//	@Router			/computeruse/a11y/node/value [post]
//
//	@id				SetAccessibilityNodeValue
func WrapSetAccessibilityNodeValueHandler(fn func(*AccessibilitySetValueRequest) (*Empty, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AccessibilitySetValueRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response, err := fn(&req)
		if err != nil {
			writeA11yError(c, err)
			return
		}
		c.JSON(http.StatusOK, response)
	}
}
