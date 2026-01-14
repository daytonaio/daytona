// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

//go:build windows

package computeruse

import (
	"syscall"
	"time"
	"unsafe"
)

var (
	user32                   = syscall.NewLazyDLL("user32.dll")
	procGetCursorPos         = user32.NewProc("GetCursorPos")
	procSetCursorPos         = user32.NewProc("SetCursorPos")
	procSendInput            = user32.NewProc("SendInput")
	procGetSystemMetrics     = user32.NewProc("GetSystemMetrics")
	procEnumWindows          = user32.NewProc("EnumWindows")
	procGetWindowTextW       = user32.NewProc("GetWindowTextW")
	procGetWindowTextLengthW = user32.NewProc("GetWindowTextLengthW")
	procIsWindowVisible      = user32.NewProc("IsWindowVisible")
)

// POINT structure for GetCursorPos
type point struct {
	X int32
	Y int32
}

// INPUT structure for SendInput
type input struct {
	Type uint32
	Mi   mouseInput
}

type mouseInput struct {
	Dx          int32
	Dy          int32
	MouseData   uint32
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

type keybdInput struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
	_           [8]byte // padding to match union size
}

// Constants for mouse input
const (
	INPUT_MOUSE    = 0
	INPUT_KEYBOARD = 1

	MOUSEEVENTF_MOVE       = 0x0001
	MOUSEEVENTF_LEFTDOWN   = 0x0002
	MOUSEEVENTF_LEFTUP     = 0x0004
	MOUSEEVENTF_RIGHTDOWN  = 0x0008
	MOUSEEVENTF_RIGHTUP    = 0x0010
	MOUSEEVENTF_MIDDLEDOWN = 0x0020
	MOUSEEVENTF_MIDDLEUP   = 0x0040
	MOUSEEVENTF_WHEEL      = 0x0800
	MOUSEEVENTF_ABSOLUTE   = 0x8000

	KEYEVENTF_KEYUP   = 0x0002
	KEYEVENTF_UNICODE = 0x0004

	WHEEL_DELTA = 120

	SM_CXSCREEN = 0
	SM_CYSCREEN = 1
)

// getMousePosition returns the current mouse cursor position using Windows API
func getMousePosition() (int, int) {
	var pt point
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	return int(pt.X), int(pt.Y)
}

// setMousePosition sets the mouse cursor position using Windows API
func setMousePosition(x, y int) error {
	ret, _, err := procSetCursorPos.Call(uintptr(x), uintptr(y))
	if ret == 0 {
		return err
	}
	return nil
}

// sendMouseInput sends a mouse input event
func sendMouseInput(flags uint32, dx, dy int32, mouseData uint32) error {
	var inp input
	inp.Type = INPUT_MOUSE
	inp.Mi.DwFlags = flags
	inp.Mi.Dx = dx
	inp.Mi.Dy = dy
	inp.Mi.MouseData = mouseData

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		unsafe.Sizeof(inp),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// mouseClick performs a mouse click
func mouseClick(button string, double bool) error {
	var downFlag, upFlag uint32

	switch button {
	case "right":
		downFlag = MOUSEEVENTF_RIGHTDOWN
		upFlag = MOUSEEVENTF_RIGHTUP
	case "middle":
		downFlag = MOUSEEVENTF_MIDDLEDOWN
		upFlag = MOUSEEVENTF_MIDDLEUP
	default: // left
		downFlag = MOUSEEVENTF_LEFTDOWN
		upFlag = MOUSEEVENTF_LEFTUP
	}

	clicks := 1
	if double {
		clicks = 2
	}

	for i := 0; i < clicks; i++ {
		if err := sendMouseInput(downFlag, 0, 0, 0); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
		if err := sendMouseInput(upFlag, 0, 0, 0); err != nil {
			return err
		}
		if i < clicks-1 {
			time.Sleep(50 * time.Millisecond)
		}
	}

	return nil
}

// mouseDown presses a mouse button down
func mouseDown(button string) error {
	var flag uint32
	switch button {
	case "right":
		flag = MOUSEEVENTF_RIGHTDOWN
	case "middle":
		flag = MOUSEEVENTF_MIDDLEDOWN
	default:
		flag = MOUSEEVENTF_LEFTDOWN
	}
	return sendMouseInput(flag, 0, 0, 0)
}

// mouseUp releases a mouse button
func mouseUp(button string) error {
	var flag uint32
	switch button {
	case "right":
		flag = MOUSEEVENTF_RIGHTUP
	case "middle":
		flag = MOUSEEVENTF_MIDDLEUP
	default:
		flag = MOUSEEVENTF_LEFTUP
	}
	return sendMouseInput(flag, 0, 0, 0)
}

// mouseScroll scrolls the mouse wheel
func mouseScroll(amount int, direction string) error {
	var scrollAmount int32
	if direction == "up" {
		scrollAmount = int32(amount * WHEEL_DELTA)
	} else {
		scrollAmount = int32(-amount * WHEEL_DELTA)
	}
	return sendMouseInput(MOUSEEVENTF_WHEEL, 0, 0, uint32(scrollAmount))
}

// Virtual key codes
var vkCodes = map[string]uint16{
	"backspace": 0x08, "tab": 0x09, "enter": 0x0D, "return": 0x0D,
	"shift": 0x10, "ctrl": 0x11, "control": 0x11, "alt": 0x12, "menu": 0x12,
	"pause": 0x13, "capslock": 0x14, "caps": 0x14,
	"escape": 0x1B, "esc": 0x1B, "space": 0x20,
	"pageup": 0x21, "pagedown": 0x22, "end": 0x23, "home": 0x24,
	"left": 0x25, "up": 0x26, "right": 0x27, "down": 0x28,
	"printscreen": 0x2C, "insert": 0x2D, "delete": 0x2E,
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
	"a": 0x41, "b": 0x42, "c": 0x43, "d": 0x44, "e": 0x45,
	"f": 0x46, "g": 0x47, "h": 0x48, "i": 0x49, "j": 0x4A,
	"k": 0x4B, "l": 0x4C, "m": 0x4D, "n": 0x4E, "o": 0x4F,
	"p": 0x50, "q": 0x51, "r": 0x52, "s": 0x53, "t": 0x54,
	"u": 0x55, "v": 0x56, "w": 0x57, "x": 0x58, "y": 0x59, "z": 0x5A,
	"win": 0x5B, "cmd": 0x5B, "super": 0x5B, "lwin": 0x5B, "rwin": 0x5C,
	"numpad0": 0x60, "numpad1": 0x61, "numpad2": 0x62, "numpad3": 0x63,
	"numpad4": 0x64, "numpad5": 0x65, "numpad6": 0x66, "numpad7": 0x67,
	"numpad8": 0x68, "numpad9": 0x69,
	"multiply": 0x6A, "add": 0x6B, "subtract": 0x6D, "decimal": 0x6E, "divide": 0x6F,
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73, "f5": 0x74,
	"f6": 0x75, "f7": 0x76, "f8": 0x77, "f9": 0x78, "f10": 0x79,
	"f11": 0x7A, "f12": 0x7B,
	"numlock": 0x90, "scrolllock": 0x91,
	"lshift": 0xA0, "rshift": 0xA1, "lctrl": 0xA2, "rctrl": 0xA3,
	"lalt": 0xA4, "ralt": 0xA5,
	";": 0xBA, "=": 0xBB, ",": 0xBC, "-": 0xBD, ".": 0xBE, "/": 0xBF,
	"`": 0xC0, "[": 0xDB, "\\": 0xDC, "]": 0xDD, "'": 0xDE,
}

// getVKCode returns the virtual key code for a key name
func getVKCode(key string) uint16 {
	if vk, ok := vkCodes[key]; ok {
		return vk
	}
	// If single character, use its ASCII code
	if len(key) == 1 {
		c := key[0]
		if c >= 'a' && c <= 'z' {
			return uint16(c - 'a' + 'A')
		}
		return uint16(c)
	}
	return 0
}

// sendKeyInput sends a keyboard input event
func sendKeyInput(vk uint16, flags uint32) error {
	var inp struct {
		Type uint32
		Ki   keybdInput
	}
	inp.Type = INPUT_KEYBOARD
	inp.Ki.WVk = vk
	inp.Ki.DwFlags = flags

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&inp)),
		unsafe.Sizeof(inp),
	)
	if ret == 0 {
		return err
	}
	return nil
}

// keyTap presses and releases a key
func keyTap(key string, modifiers []string) error {
	// Press modifiers
	for _, mod := range modifiers {
		vk := getVKCode(mod)
		if vk != 0 {
			if err := sendKeyInput(vk, 0); err != nil {
				return err
			}
		}
	}

	// Press and release the main key
	vk := getVKCode(key)
	if vk != 0 {
		if err := sendKeyInput(vk, 0); err != nil {
			return err
		}
		time.Sleep(10 * time.Millisecond)
		if err := sendKeyInput(vk, KEYEVENTF_KEYUP); err != nil {
			return err
		}
	}

	// Release modifiers in reverse order
	for i := len(modifiers) - 1; i >= 0; i-- {
		vk := getVKCode(modifiers[i])
		if vk != 0 {
			if err := sendKeyInput(vk, KEYEVENTF_KEYUP); err != nil {
				return err
			}
		}
	}

	return nil
}

// typeString types a string using Unicode input
func typeString(text string, delay int) error {
	for _, char := range text {
		// Use Unicode input for each character
		var inp struct {
			Type uint32
			Ki   keybdInput
		}
		inp.Type = INPUT_KEYBOARD
		inp.Ki.WScan = uint16(char)
		inp.Ki.DwFlags = KEYEVENTF_UNICODE

		// Key down
		ret, _, err := procSendInput.Call(
			1,
			uintptr(unsafe.Pointer(&inp)),
			unsafe.Sizeof(inp),
		)
		if ret == 0 {
			return err
		}

		// Key up
		inp.Ki.DwFlags = KEYEVENTF_UNICODE | KEYEVENTF_KEYUP
		ret, _, err = procSendInput.Call(
			1,
			uintptr(unsafe.Pointer(&inp)),
			unsafe.Sizeof(inp),
		)
		if ret == 0 {
			return err
		}

		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	return nil
}

// getScreenSize returns the screen width and height
func getScreenSize() (int, int) {
	width, _, _ := procGetSystemMetrics.Call(uintptr(SM_CXSCREEN))
	height, _, _ := procGetSystemMetrics.Call(uintptr(SM_CYSCREEN))
	return int(width), int(height)
}

// Window enumeration callback type
type enumWindowsProc func(hwnd syscall.Handle, lParam uintptr) uintptr

// windowInfo holds information about a window
type windowInfo struct {
	Handle  syscall.Handle
	Title   string
	Visible bool
}

var windowList []windowInfo

//go:uintptrescapes
func enumWindowsCallback(hwnd syscall.Handle, lParam uintptr) uintptr {
	// Check if window is visible
	ret, _, _ := procIsWindowVisible.Call(uintptr(hwnd))
	if ret == 0 {
		return 1 // Continue enumeration
	}

	// Get window title length
	length, _, _ := procGetWindowTextLengthW.Call(uintptr(hwnd))
	if length == 0 {
		return 1 // Continue enumeration
	}

	// Get window title
	buf := make([]uint16, length+1)
	procGetWindowTextW.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&buf[0])), uintptr(length+1))
	title := syscall.UTF16ToString(buf)

	if title != "" {
		windowList = append(windowList, windowInfo{
			Handle:  hwnd,
			Title:   title,
			Visible: true,
		})
	}

	return 1 // Continue enumeration
}

// getWindowsList returns a list of visible windows
func getWindowsList() []windowInfo {
	windowList = nil // Reset the list
	procEnumWindows.Call(syscall.NewCallback(enumWindowsCallback), 0)
	return windowList
}
