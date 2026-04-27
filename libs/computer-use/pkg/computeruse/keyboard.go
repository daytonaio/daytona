// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/go-vgo/robotgo"
)

type typingActionType int

const (
	typingActionText typingActionType = iota
	typingActionEnter
)

type typingAction struct {
	kind typingActionType
	text string
}

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

func buildTypingActions(text string) ([]typingAction, error) {
	text = strings.ReplaceAll(text, "\r\n", "\n")

	actions := make([]typingAction, 0)
	var currentText strings.Builder

	flushText := func() {
		if currentText.Len() == 0 {
			return
		}
		actions = append(actions, typingAction{
			kind: typingActionText,
			text: currentText.String(),
		})
		currentText.Reset()
	}

	for _, r := range text {
		switch r {
		case '\n', '\r':
			flushText()
			actions = append(actions, typingAction{kind: typingActionEnter})
		case '\t':
			return nil, fmt.Errorf(
				"keyboard.type does not translate '\\t' to Tab; use keyboard.press(\"tab\") for Tab key events",
			)
		case '\u2028', '\u2029':
			return nil, fmt.Errorf("unsupported separator character in keyboard.type: U+%04X", r)
		default:
			if unicode.IsControl(r) {
				return nil, fmt.Errorf("unsupported control character in keyboard.type: U+%04X", r)
			}
			currentText.WriteRune(r)
		}
	}

	flushText()
	return actions, nil
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
