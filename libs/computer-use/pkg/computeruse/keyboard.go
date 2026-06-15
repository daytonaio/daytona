//go:build linux

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"time"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) TypeText(req *computeruse.KeyboardTypeRequest) (*computeruse.Empty, error) {
	actions, err := buildTypingActions(req.Text)
	if err != nil {
		return nil, err
	}

	for _, action := range actions {
		switch action.kind {
		case typingActionText:
			if req.Delay > 0 {
				robotgo.TypeStr(action.text, 0, req.Delay)
			} else {
				robotgo.TypeStr(action.text)
			}
		case typingActionEnter:
			if err := robotgo.KeyTap("enter"); err != nil {
				return nil, err
			}
			if req.Delay > 0 {
				time.Sleep(time.Duration(req.Delay) * time.Millisecond)
			}
		}
	}

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressKey(req *computeruse.KeyboardPressRequest) (*computeruse.Empty, error) {
	chord, err := normalizeKeyboardPress(req.Key, req.Modifiers)
	if err != nil {
		return nil, err
	}

	if len(chord.modifiers) > 0 {
		err = robotgo.KeyTap(chord.key, chord.modifiers)
	} else {
		err = robotgo.KeyTap(chord.key)
	}
	if err != nil {
		return nil, err
	}

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressHotkey(req *computeruse.KeyboardHotkeyRequest) (*computeruse.Empty, error) {
	chord, err := normalizeKeyboardHotkey(req.Keys)
	if err != nil {
		return nil, err
	}

	if len(chord.modifiers) > 0 {
		err = robotgo.KeyTap(chord.key, chord.modifiers)
	} else {
		err = robotgo.KeyTap(chord.key)
	}
	if err != nil {
		return nil, err
	}

	return new(computeruse.Empty), nil
}
