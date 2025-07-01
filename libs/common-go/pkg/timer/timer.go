// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package timer

import (
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"
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

// Timer returns a function that prints (on trace level) the name of the calling
// function and the elapsed time between the call to Timer and
// the call to the returned function. The returned function is
// intended to be used in a defer statement:
//
//	defer Timer()()
func Timer() func() {
	name := callerName(1)
	start := time.Now()
	return func() {
		log.Tracef("%s took %v", name, time.Since(start))
	}
}
