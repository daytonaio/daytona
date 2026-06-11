// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package common

import (
	"strings"
	"testing"
)

func TestMatchFdExhaustion(t *testing.T) {
	cases := []struct {
		name string
		msg  string
		want bool
	}{
		// Real syscall/runtime formats — all carry the strerror phrase or
		// an errno 24, and must classify.
		{"go open", "open /x: too many open files", true},
		{"node emfile", "Error: EMFILE: too many open files, open '/x'", true},
		{"node legacy emfile", "Error: EMFILE, too many open files", true},
		{"strace emfile", "EMFILE (Too many open files)", true},
		{"python errno 24", "[Errno 24] Too many open files", true},
		{"accept4 emfile", "accept4: too many open files", true},
		{"fork exec emfile", "fork/exec /bin/sh: too many open files", true},
		{"enfile system-wide", "too many open files in system", true},
		{"errno space 24", "socket failed: errno 24", true},
		{"errno equals 24", "errno=24", true},
		{"errno colon 24", "errno: 24", true},
		{"loader with errno 24", "error while loading shared libraries: libc.so.6: cannot open shared object file: Error 24", true},
		// Bare EMFILE/ENFILE tokens without the phrase or errno number are
		// filename-shaped or echoed text, never a real error format.
		{"emfile in path segment", "/tmp/EMFILE.txt: no such file", false},
		{"emfile filename at start", "EMFILE.txt: permission denied", false},
		{"emfile glued to word", "openEMFILE", false},
		{"emfile command echo", "cat EMFILE", false},
		{"bare emfile code", "spawn failed with EMFILE", false},
		{"bare enfile code", "socket: ENFILE", false},
		{"enfile glued to word", "myENFILEhandler crashed", false},
		{"empty", "", false},
		{"file not found", "file not found", false},
		{"storage error", "no space left on device", false},
		{"substring trap", "systemfile missing", false},
		{"other errno", "errno 240", false},
		{"loader without errno 24", "error while loading shared libraries: libfoo.so: cannot open shared object file: No such file or directory", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := MatchFdExhaustion(tc.msg); got != tc.want {
				t.Errorf("MatchFdExhaustion(%q) = %v, want %v", tc.msg, got, tc.want)
			}
		})
	}
}

func TestFdExhaustionReason(t *testing.T) {
	if got := FdExhaustionReason("too many open files"); got != "fd-exhaustion: too many open files" {
		t.Errorf("FdExhaustionReason = %q", got)
	}

	long := strings.Repeat("x", 500)
	got := FdExhaustionReason(long)
	if len(got) != len(DegradedReasonFdExhaustionPrefix)+200 {
		t.Errorf("FdExhaustionReason did not truncate: len = %d", len(got))
	}
}

func TestClassifyToolboxFdExhaustion(t *testing.T) {
	cases := []struct {
		name        string
		daemonPath  string
		statusCode  int
		contentType string
		body        string
		wantOk      bool
	}{
		{
			name:        "exec daemon spawn failure",
			daemonPath:  "/process/execute",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"exitCode":-1,"result":"fork/exec /bin/sh: too many open files"}`,
			wantOk:      true,
		},
		{
			name:        "exec success printing fd text is not degraded",
			daemonPath:  "/process/execute",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"exitCode":0,"result":"grep: too many open files"}`,
			wantOk:      false,
		},
		{
			name:        "exec loader failure with errno 24",
			daemonPath:  "/process/execute",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"exitCode":127,"result":"error while loading shared libraries: libc.so.6: cannot open shared object file: Error 24"}`,
			wantOk:      true,
		},
		{
			name:        "exec nonzero exit without loader co-occurrence",
			daemonPath:  "/process/execute",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"exitCode":1,"result":"too many open files"}`,
			wantOk:      false,
		},
		{
			name:        "error response with fd text",
			daemonPath:  "/process/session/abc/exec",
			statusCode:  500,
			contentType: "application/json; charset=utf-8",
			body:        `{"statusCode":500,"message":"failed to execute command: fork/exec /bin/sh: too many open files"}`,
			wantOk:      true,
		},
		{
			name:        "error response without fd text",
			daemonPath:  "/process/session/abc/exec",
			statusCode:  500,
			contentType: "application/json",
			body:        `{"statusCode":500,"message":"internal server error"}`,
			wantOk:      false,
		},
		{
			name:        "error response non-json is skipped",
			daemonPath:  "/files",
			statusCode:  502,
			contentType: "text/plain",
			body:        "too many open files",
			wantOk:      false,
		},
		{
			name:        "non-exec 200 path is skipped",
			daemonPath:  "/files/info",
			statusCode:  200,
			contentType: "application/json",
			body:        `{"name":"too many open files"}`,
			wantOk:      false,
		},
		{
			name:        "exec malformed body is skipped",
			daemonPath:  "/process/execute",
			statusCode:  200,
			contentType: "application/json",
			body:        `not json`,
			wantOk:      false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reason, ok := ClassifyToolboxFdExhaustion(tc.daemonPath, tc.statusCode, tc.contentType, []byte(tc.body))
			if ok != tc.wantOk {
				t.Fatalf("ClassifyToolboxFdExhaustion(%q, %d) ok = %v, want %v", tc.daemonPath, tc.statusCode, ok, tc.wantOk)
			}
			if ok && !strings.HasPrefix(reason, DegradedReasonFdExhaustionPrefix) {
				t.Errorf("reason %q missing prefix %q", reason, DegradedReasonFdExhaustionPrefix)
			}
			if !ok && reason != "" {
				t.Errorf("miss should return empty reason, got %q", reason)
			}
		})
	}
}
