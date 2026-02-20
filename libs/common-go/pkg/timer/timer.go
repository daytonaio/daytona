// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package timer

import (
	"log/slog"
	"runtime"
	"time"
)

// callerName returns the name of the function skip frames up the call stack.
func callerName(skip int) string {
	const unknown = "unknown"
	pcs := make([]uintptr, 1)
	n := runtime.Callers(skip+2, pcs)
	if n < 1 {
		return unknown
	}
	frame, _ := runtime.CallersFrames(pcs).Next()
	if frame.Function == "" {
		return unknown
	}
	return frame.Function
}

// Timer returns a function that logs (at debug level) the name of the calling
// function and the elapsed time between the call to Timer and
// the call to the returned function. The returned function is
// intended to be used in a defer statement:
//
//	defer Timer()()
func Timer() func() {
	name := callerName(1)
	start := time.Now()
	return func() {
		slog.Debug("function timing", "function", name, "duration", time.Since(start))
	}
}
