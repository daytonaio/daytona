// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"strings"
)

// TypeText types text with optional delay between keystrokes
func (c *ComputerUse) TypeText(req *KeyboardTypeRequest) (*Empty, error) {
	if err := typeString(req.Text, req.Delay); err != nil {
		return nil, err
	}
	return new(Empty), nil
}

// PressKey presses a key with optional modifiers
func (c *ComputerUse) PressKey(req *KeyboardPressRequest) (*Empty, error) {
	if err := keyTap(req.Key, req.Modifiers); err != nil {
		return nil, err
	}
	return new(Empty), nil
}

// PressHotkey presses a hotkey combination (e.g., "ctrl+c", "alt+f4")
func (c *ComputerUse) PressHotkey(req *KeyboardHotkeyRequest) (*Empty, error) {
	keys := strings.Split(req.Keys, "+")
	if len(keys) < 2 {
		return nil, fmt.Errorf("invalid hotkey format: expected format like 'ctrl+c'")
	}

	mainKey := keys[len(keys)-1]
	modifiers := keys[:len(keys)-1]

	if err := keyTap(mainKey, modifiers); err != nil {
		return nil, err
	}

	return new(Empty), nil
}
