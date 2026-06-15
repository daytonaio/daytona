//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package toolbox

import (
	"fmt"
	"time"

	"github.com/daytonaio/daemon/pkg/winsession"
)

const userHomeDirSessionTimeout = 5 * time.Second

func getUserHomeDir() (string, error) {
	token, err := winsession.ActiveConsoleUserToken(userHomeDirSessionTimeout)
	if err != nil {
		return "", fmt.Errorf("failed to resolve console user for home directory: %w", err)
	}
	defer token.Close()

	dir, err := token.GetUserProfileDirectory()
	if err != nil {
		return "", fmt.Errorf("failed to get user profile directory: %w", err)
	}
	return dir, nil
}
