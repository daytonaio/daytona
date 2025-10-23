// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package interpreter

import (
	"time"
)

var (
    activeSession *InterpreterSession
)

// GetOrCreateSession returns the active session or creates a new one
func GetOrCreateSession(workDir string) (*InterpreterSession, error) {
    if activeSession != nil {
        info := activeSession.Info()
        if info.Active {
            return activeSession, nil
        }
    }

    // create new
    s := &InterpreterSession{
        info: InterpreterSessionInfo{
            ID:        "default",
            Cwd:       workDir,
            CreatedAt: time.Now(),
            Active:    false,
            Language:  "python",
        },
    }

    if err := s.start(); err != nil {
        return nil, err
    }

    activeSession = s
    return activeSession, nil
}


