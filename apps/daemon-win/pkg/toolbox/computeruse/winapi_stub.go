// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build !windows

package computeruse

// Stub implementations for non-Windows platforms
// These allow the code to compile for testing/linting on Linux/macOS

func getMousePosition() (int, int) {
	return 0, 0
}

func setMousePosition(x, y int) error {
	return nil
}

func mouseClick(button string, double bool) error {
	return nil
}

func mouseDown(button string) error {
	return nil
}

func mouseUp(button string) error {
	return nil
}

func mouseScroll(amount int, direction string) error {
	return nil
}

func keyTap(key string, modifiers []string) error {
	return nil
}

func typeString(text string, delay int) error {
	return nil
}

func getScreenSize() (int, int) {
	return 1920, 1080
}

type windowInfo struct {
	Handle  uintptr
	Title   string
	Visible bool
}

func getWindowsList() []windowInfo {
	return nil
}
