//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"strings"
)

// TypeText types text with optional delay between keystrokes
func (c *ComputerUse) TypeText(req *computeruse.KeyboardTypeRequest) (*computeruse.Empty, error) {
	if err := typeString(req.Text, req.Delay); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

// PressKey presses a key with optional modifiers
func (c *ComputerUse) PressKey(req *computeruse.KeyboardPressRequest) (*computeruse.Empty, error) {
	if err := keyTap(req.Key, req.Modifiers); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

// PressHotkey presses a hotkey combination (e.g., "ctrl+c", "alt+f4")
func (c *ComputerUse) PressHotkey(req *computeruse.KeyboardHotkeyRequest) (*computeruse.Empty, error) {
	keys := strings.Split(req.Keys, "+")
	if len(keys) < 2 {
		return nil, fmt.Errorf("invalid hotkey format: expected format like 'ctrl+c'")
	}

	mainKey := keys[len(keys)-1]
	modifiers := keys[:len(keys)-1]

	if err := keyTap(mainKey, modifiers); err != nil {
		return nil, err
	}

	return new(computeruse.Empty), nil
}
