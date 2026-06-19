//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
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
	chord, err := normalizeKeyboardPress(req.Key, req.Modifiers)
	if err != nil {
		return nil, err
	}

	if err := keyTap(chord.key, chord.modifiers); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}

// PressHotkey presses a hotkey combination (e.g., "ctrl+c", "alt+f4")
func (c *ComputerUse) PressHotkey(req *computeruse.KeyboardHotkeyRequest) (*computeruse.Empty, error) {
	chord, err := normalizeKeyboardHotkey(req.Keys)
	if err != nil {
		return nil, err
	}

	if err := keyTap(chord.key, chord.modifiers); err != nil {
		return nil, err
	}
	return new(computeruse.Empty), nil
}
