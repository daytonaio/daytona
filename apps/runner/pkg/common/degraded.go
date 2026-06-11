// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"encoding/json"
	"regexp"
	"strings"
)

// DegradedReasonFdExhaustionPrefix prefixes every degradedReason produced by
// file-descriptor exhaustion detection.
const DegradedReasonFdExhaustionPrefix = "fd-exhaustion: "

// maxObservedLen bounds the observed error text embedded in a degraded reason.
const maxObservedLen = 200

// MatchFdExhaustion reports whether msg carries a file-descriptor exhaustion
// signature: EMFILE/ENFILE errno codes (matched case-sensitively as standalone
// tokens, see fdErrnoTokenRe), the classic "too many open files" text (which
// also covers ENFILE's "too many open files in system"), or a dynamic-loader
// failure co-occurring with errno 24.
//
// NOTE: these signatures are surfacing-only. They must NEVER be added to
// recoverableErrorPatterns (recovery.go) — fd exhaustion must not trigger
// automated recovery (restart/recreate/disk-resize). Enforced by
// TestDeduceRecoveryType_FdExhaustionNotActionable.
func MatchFdExhaustion(msg string) bool {
	if msg == "" {
		return false
	}

	lower := strings.ToLower(msg)

	if strings.Contains(lower, "too many open files") {
		return true
	}
	if matchFdErrnoToken(msg) {
		return true
	}

	return matchLoaderFdExhaustion(msg, lower)
}

// matchLoaderFdExhaustion detects dynamic-loader failures caused by fd
// exhaustion: "error while loading shared libraries" co-occurring with
// errno 24 (or an explicit fd-exhaustion code/text).
func matchLoaderFdExhaustion(msg, lower string) bool {
	if !strings.Contains(lower, "error while loading shared libraries") {
		return false
	}
	return strings.Contains(lower, "error 24") ||
		strings.Contains(lower, "too many open files") ||
		matchFdErrnoToken(msg)
}

// fdErrnoTokenRe matches EMFILE/ENFILE only as standalone tokens: not glued
// to other word characters (openEMFILE) and not adjacent to a path separator
// (/tmp/EMFILE.txt), so user-controlled text echoed in errors cannot
// false-positive. Case-sensitive on purpose — errno codes are uppercase.
var fdErrnoTokenRe = regexp.MustCompile(`(^|[^/0-9A-Za-z_])(EMFILE|ENFILE)($|[^/0-9A-Za-z_])`)

// matchFdErrnoToken reports whether msg contains EMFILE or ENFILE as a
// standalone errno token (see fdErrnoTokenRe).
func matchFdErrnoToken(msg string) bool {
	return fdErrnoTokenRe.MatchString(msg)
}

// FdExhaustionReason builds the degradedReason value from the observed error
// text, truncated to a sane length.
func FdExhaustionReason(observed string) string {
	if len(observed) > maxObservedLen {
		observed = observed[:maxObservedLen]
	}
	return DegradedReasonFdExhaustionPrefix + observed
}

// toolboxExecuteResponse mirrors the daemon's ExecuteResponse
// (apps/daemon/pkg/toolbox/process/types.go).
type toolboxExecuteResponse struct {
	ExitCode int    `json:"exitCode"`
	Result   string `json:"result"`
}

// ClassifyToolboxFdExhaustion is a pure classifier for daemon responses
// observed at the toolbox proxy. It returns a degraded reason and true when
// the response carries an fd-exhaustion signature.
//
// Error responses (status >= 400) with a JSON body are matched on the raw
// body text. Successful (200) responses are only inspected for the
// /process/execute path, and only count when the daemon itself failed to
// spawn the process (exitCode == -1) or the process died in the dynamic
// loader (non-zero exit + loader co-occurrence) — a successful command merely
// printing "too many open files" must NOT mark the sandbox degraded.
func ClassifyToolboxFdExhaustion(daemonPath string, statusCode int, contentType string, body []byte) (string, bool) {
	switch {
	case statusCode >= 400:
		if !strings.Contains(contentType, "application/json") {
			return "", false
		}
		if msg := string(body); MatchFdExhaustion(msg) {
			return FdExhaustionReason(msg), true
		}
	case statusCode == 200 && daemonPath == "/process/execute":
		var resp toolboxExecuteResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return "", false
		}
		if resp.ExitCode == -1 && MatchFdExhaustion(resp.Result) {
			return FdExhaustionReason(resp.Result), true
		}
		if resp.ExitCode != 0 && matchLoaderFdExhaustion(resp.Result, strings.ToLower(resp.Result)) {
			return FdExhaustionReason(resp.Result), true
		}
	}

	return "", false
}
