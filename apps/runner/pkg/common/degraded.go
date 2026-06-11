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
// signature in a canonical syscall/runtime format, matched case-insensitively:
//
//   - the strerror text "too many open files" — covers Go/libc
//     "open /x: too many open files", Node "EMFILE: too many open files",
//     strace "EMFILE (Too many open files)", and ENFILE's
//     "too many open files in system";
//   - an explicit errno 24 spelling: "[Errno 24]", "errno 24", "errno=24",
//     "errno: 24";
//   - a dynamic-loader failure co-occurring with errno 24 (see
//     matchLoaderFdExhaustion).
//
// Bare EMFILE/ENFILE tokens WITHOUT one of those contexts deliberately do
// not classify: every real syscall/runtime format carries the strerror
// phrase or the errno number, while filename echoes in user-controlled text
// ("/tmp/EMFILE.txt: no such file") never do.
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

	if matchFdCanonical(lower) {
		return true
	}

	return matchLoaderFdExhaustion(lower)
}

// matchFdCanonical reports whether the lowercased text carries a canonical
// fd-exhaustion form: the strerror phrase or an explicit errno 24.
func matchFdCanonical(lower string) bool {
	return strings.Contains(lower, "too many open files") || fdErrno24Re.MatchString(lower)
}

// fdErrno24Re matches explicit errno-24 spellings ("[errno 24]", "errno 24",
// "errno=24", "errno: 24") in lowercased text. The trailing guard keeps
// other errno values (e.g. 240) from matching.
var fdErrno24Re = regexp.MustCompile(`errno[\s:=]*24([^0-9]|$)`)

// matchLoaderFdExhaustion detects dynamic-loader failures caused by fd
// exhaustion: "error while loading shared libraries" co-occurring with
// errno 24 — ld.so prints "Error 24" instead of the strerror text — or a
// canonical fd-exhaustion form. Takes lowercased text.
func matchLoaderFdExhaustion(lower string) bool {
	if !strings.Contains(lower, "error while loading shared libraries") {
		return false
	}
	return strings.Contains(lower, "error 24") || matchFdCanonical(lower)
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
		if resp.ExitCode != 0 && matchLoaderFdExhaustion(strings.ToLower(resp.Result)) {
			return FdExhaustionReason(resp.Result), true
		}
	}

	return "", false
}
