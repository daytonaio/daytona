// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !no_gui
// +build !no_gui

package computeruse

import (
	"fmt"
	"strings"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
)

func (u *ComputerUse) TypeText(req *computeruse.TypeTextRequest) (*computeruse.Empty, error) {
	if req.Delay > 0 {
		robotgo.TypeStr(req.Text, req.Delay)
	} else {
		robotgo.TypeStr(req.Text)
	}

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressKey(req *computeruse.PressKeyRequest) (*computeruse.Empty, error) {
	if len(req.Modifiers) > 0 {
		robotgo.KeyTap(req.Key, req.Modifiers)
	} else {
		robotgo.KeyTap(req.Key)
	}

	return new(computeruse.Empty), nil
}

func (u *ComputerUse) PressHotkey(req *computeruse.PressHotkeyRequest) (*computeruse.Empty, error) {
	keys := strings.Split(req.Keys, "+")
	if len(keys) < 2 {
		return nil, fmt.Errorf("invalid hotkey format")
	}

	mainKey := keys[len(keys)-1]
	modifiers := keys[:len(keys)-1]

	robotgo.KeyTap(mainKey, modifiers)

	return new(computeruse.Empty), nil
}
