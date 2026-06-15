//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package winsession resolves the interactive console session on Windows so
// daemon-spawned processes (computer-use plugin, ffmpeg screen capture) can
// run on the logged-on user's desktop instead of session 0.
package winsession

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

const consoleSessionPollInterval = 500 * time.Millisecond

// ErrNoActiveConsoleSession is returned when no interactive user is logged on
// within the caller-supplied timeout. In Daytona's Windows sandbox image this
// should never fire after AutoLogon completes.
var ErrNoActiveConsoleSession = errors.New("no active console session available; ensure a user is logged on (AutoLogon)")

// ActiveConsoleUserToken polls WTSGetActiveConsoleSessionId until a non-sentinel
// session id appears (or timeout elapses), then queries and duplicates the
// user's token to a primary token suitable for exec.Cmd.SysProcAttr.Token.
// Caller owns the returned handle and MUST ensure it lives until
// exec.Cmd.Start() returns; do NOT Close it immediately after attaching it to
// SysProcAttr, since Windows duplicates the handle during CreateProcessAsUser.
func ActiveConsoleUserToken(timeout time.Duration) (windows.Token, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error
	for {
		sid := windows.WTSGetActiveConsoleSessionId()
		if sid != 0xFFFFFFFF {
			var raw windows.Token
			if err := windows.WTSQueryUserToken(sid, &raw); err != nil {
				lastErr = fmt.Errorf("WTSQueryUserToken(sid=%d): %w", sid, err)
			} else {
				var primary windows.Token
				err := windows.DuplicateTokenEx(
					raw,
					windows.MAXIMUM_ALLOWED,
					nil,
					windows.SecurityImpersonation,
					windows.TokenPrimary,
					&primary,
				)
				raw.Close()
				if err == nil {
					return primary, nil
				}
				lastErr = fmt.Errorf("DuplicateTokenEx(sid=%d): %w", sid, err)
			}
		}
		if time.Now().After(deadline) {
			// Distinguish "no session ever appeared" from "token acquisition
			// kept failing" (e.g. ERROR_PRIVILEGE_NOT_HELD when the daemon is
			// not running as SYSTEM); the sentinel stays errors.Is-able.
			if lastErr != nil {
				return 0, fmt.Errorf("%w (last attempt: %v)", ErrNoActiveConsoleSession, lastErr)
			}
			return 0, ErrNoActiveConsoleSession
		}
		time.Sleep(consoleSessionPollInterval)
	}
}
