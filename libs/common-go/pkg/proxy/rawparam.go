// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package proxy

import "github.com/gin-gonic/gin"

// RawParam returns the raw (percent-encoded) value of a Gin wildcard parameter,
// preserving the original encoding from the request URL.
//
// ctx.Param() fully decodes percent-encoded characters (e.g. "%40" → "@"),
// and the decoded form is then re-encoded by Go's url package using its own
// rules (e.g. "(" → "%28"). This round-trip mutates the path that reaches the
// backend. RawParam avoids the mutation by recovering the original encoded
// suffix directly from url.URL.EscapedPath().
//
// It assumes the non-wildcard prefix of the route contains no percent-encoded
// characters. This holds for all routes in this codebase (IDs are UUIDs,
// ports are numeric).
func RawParam(ctx *gin.Context, paramName string) string {
	decodedParam := ctx.Param(paramName)
	if decodedParam == "" {
		return ""
	}
	escapedFull := ctx.Request.URL.EscapedPath()
	decodedFull := ctx.Request.URL.Path
	prefixLen := len(decodedFull) - len(decodedParam)
	if prefixLen < 0 || prefixLen > len(escapedFull) {
		// Fallback: return the decoded form (pre-existing behaviour)
		return decodedParam
	}
	return escapedFull[prefixLen:]
}
