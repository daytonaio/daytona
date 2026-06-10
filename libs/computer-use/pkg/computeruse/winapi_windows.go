//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ---------------------------------------------------------------------------
// Win32 bindings
//
// We avoid pulling in github.com/lxn/win (which is unmaintained and quite
// large) and instead bind the handful of user32 entry points we need
// directly via golang.org/x/sys/windows. This keeps the dep surface small
// and the binary CGO-free.
// ---------------------------------------------------------------------------

var (
	user32                = windows.NewLazySystemDLL("user32.dll")
	procGetCursorPos      = user32.NewProc("GetCursorPos")
	procSetCursorPos      = user32.NewProc("SetCursorPos")
	procSendInput         = user32.NewProc("SendInput")
	procVkKeyScanW        = user32.NewProc("VkKeyScanW")
	procEnumWindows       = user32.NewProc("EnumWindows")
	procGetWindowTextW    = user32.NewProc("GetWindowTextW")
	procGetWindowTextLenW = user32.NewProc("GetWindowTextLengthW")
	procIsWindowVisible   = user32.NewProc("IsWindowVisible")
	procGetWindowRect     = user32.NewProc("GetWindowRect")
	procMapVirtualKeyW    = user32.NewProc("MapVirtualKeyW")
	procGetForegroundWnd  = user32.NewProc("GetForegroundWindow")
)

// SendInput type field.
const (
	inputMouse    = 0
	inputKeyboard = 1
)

// MOUSEINPUT.dwFlags values.
const (
	mouseEventF_MOVE       = 0x0001
	mouseEventF_LEFTDOWN   = 0x0002
	mouseEventF_LEFTUP     = 0x0004
	mouseEventF_RIGHTDOWN  = 0x0008
	mouseEventF_RIGHTUP    = 0x0010
	mouseEventF_MIDDLEDOWN = 0x0020
	mouseEventF_MIDDLEUP   = 0x0040
	mouseEventF_WHEEL      = 0x0800
	mouseEventF_HWHEEL     = 0x01000
	mouseEventF_ABSOLUTE   = 0x8000
	wheelDelta             = 120
)

// KEYBDINPUT.dwFlags values.
const (
	keyEventF_EXTENDEDKEY = 0x0001
	keyEventF_KEYUP       = 0x0002
	keyEventF_UNICODE     = 0x0004
	keyEventF_SCANCODE    = 0x0008
)

// MapVirtualKey translation types.
const mapVkVkToVsc = 0

// Virtual-key codes for the canonical key tokens emitted by
// keyboard_normalization.go. Aliases (e.g. "command", "return", "esc") are
// resolved there before reaching this map — do not add alias rows here.
// Anything not listed is parsed via VkKeyScanW (printable characters) or
// treated as a single Unicode character.
var virtualKeyCodes = map[string]uint16{
	// Modifiers
	"shift":  0x10,
	"ctrl":   0x11,
	"alt":    0x12,
	"cmd":    0x5B, // Windows / Super key
	"lshift": 0xA0,
	"rshift": 0xA1,
	"lctrl":  0xA2,
	"rctrl":  0xA3,
	"lalt":   0xA4,
	"ralt":   0xA5,
	"lcmd":   0x5B,
	"rcmd":   0x5C,

	// Whitespace / control
	"backspace": 0x08,
	"tab":       0x09,
	"enter":     0x0D,
	"escape":    0x1B,
	"space":     0x20,
	"capslock":  0x14,
	"menu":      0x5D, // VK_APPS

	// Navigation
	"pageup":   0x21,
	"pagedown": 0x22,
	"end":      0x23,
	"home":     0x24,
	"left":     0x25,
	"up":       0x26,
	"right":    0x27,
	"down":     0x28,
	"insert":   0x2D,
	"delete":   0x2E,

	// Function keys
	"f1": 0x70, "f2": 0x71, "f3": 0x72, "f4": 0x73,
	"f5": 0x74, "f6": 0x75, "f7": 0x76, "f8": 0x77,
	"f9": 0x78, "f10": 0x79, "f11": 0x7A, "f12": 0x7B,
	"f13": 0x7C, "f14": 0x7D, "f15": 0x7E, "f16": 0x7F,
	"f17": 0x80, "f18": 0x81, "f19": 0x82, "f20": 0x83,
	"f21": 0x84, "f22": 0x85, "f23": 0x86, "f24": 0x87,

	// Numpad
	"num0": 0x60, "num1": 0x61, "num2": 0x62, "num3": 0x63, "num4": 0x64,
	"num5": 0x65, "num6": 0x66, "num7": 0x67, "num8": 0x68, "num9": 0x69,
	"num*":      0x6A, // VK_MULTIPLY
	"num+":      0x6B, // VK_ADD
	"num-":      0x6D, // VK_SUBTRACT
	"num.":      0x6E, // VK_DECIMAL
	"num/":      0x6F, // VK_DIVIDE
	"num_enter": 0x0D,
	"num_equal": 0x92, // VK_OEM_NEC_EQUAL
	"num_lock":  0x90,

	// Numeric/letters - digits 0-9
	"0": 0x30, "1": 0x31, "2": 0x32, "3": 0x33, "4": 0x34,
	"5": 0x35, "6": 0x36, "7": 0x37, "8": 0x38, "9": 0x39,
}

// Canonical tokens that are "extended" keys — the E0-prefixed scan codes on
// a physical keyboard. Windows distinguishes some of these from their twins
// solely via KEYEVENTF_EXTENDEDKEY (e.g. numpad Enter vs the main Enter key,
// which share VK_RETURN), so key events for them must carry the flag.
//
// Per the Win32 "Keyboard Input Overview", the extended keys are the
// right-hand ALT and CTRL keys, the INS/DEL/HOME/END/PGUP/PGDN/arrow
// navigation cluster, NUM LOCK, the numpad divide and ENTER keys, and the
// Windows / Application keys.
var extendedVirtualKeys = map[string]bool{
	"ralt": true, "rctrl": true,
	"cmd": true, "lcmd": true, "rcmd": true, "menu": true,
	"insert": true, "delete": true,
	"home": true, "end": true, "pageup": true, "pagedown": true,
	"left": true, "up": true, "right": true, "down": true,
	"num_lock": true, "num/": true, "num_enter": true,
}

// pointStruct mirrors the Win32 POINT struct.
type pointStruct struct {
	X int32
	Y int32
}

// rectStruct mirrors the Win32 RECT struct.
type rectStruct struct {
	Left, Top, Right, Bottom int32
}

// mouseInput mirrors the Win32 MOUSEINPUT struct.
type mouseInput struct {
	Dx        int32
	Dy        int32
	MouseData uint32
	DwFlags   uint32
	Time      uint32
	ExtraInfo uintptr
}

// keybdInput mirrors the Win32 KEYBDINPUT struct.
type keybdInput struct {
	WVk       uint16
	WScan     uint16
	DwFlags   uint32
	Time      uint32
	ExtraInfo uintptr
}

// inputUnion is large enough to hold any INPUT.U variant on 64-bit Windows:
// MOUSEINPUT is the largest at 32 bytes (five 4-byte fields plus an 8-byte
// ExtraInfo pointer). It is [4]uint64 rather than [32]byte so the union —
// and therefore inputStruct — carries 8-byte alignment, which the unsafe
// casts in asMouse/asKeyboard require (both structs hold a pointer-sized
// ExtraInfo field). This mirrors the 64-bit (amd64/arm64) INPUT layout only;
// 32-bit Windows uses a 28-byte INPUT and is not supported.
type inputUnion [4]uint64

// inputStruct mirrors the Win32 INPUT tagged-union.
type inputStruct struct {
	Type uint32
	_    uint32 // padding to align the union on 64-bit
	U    inputUnion
}

// asMouse writes a mouseInput into the union.
func (i *inputStruct) asMouse(mi mouseInput) {
	i.Type = inputMouse
	*(*mouseInput)(unsafe.Pointer(&i.U[0])) = mi
}

// asKeyboard writes a keybdInput into the union.
func (i *inputStruct) asKeyboard(ki keybdInput) {
	i.Type = inputKeyboard
	*(*keybdInput)(unsafe.Pointer(&i.U[0])) = ki
}

// sendInputs sends one or more INPUT events.
func sendInputs(inputs []inputStruct) error {
	if len(inputs) == 0 {
		return nil
	}
	ret, _, err := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		unsafe.Sizeof(inputs[0]),
	)
	if int(ret) != len(inputs) {
		return fmt.Errorf("SendInput sent %d/%d events: %v", ret, len(inputs), err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Mouse
// ---------------------------------------------------------------------------

// getMousePositionChecked reads the cursor position, reporting failure
// instead of fabricating (0,0). GetCursorPos fails with ERROR_ACCESS_DENIED
// when the calling desktop is unavailable (locked session, secure desktop),
// so API handlers must surface the error rather than return zero values.
func getMousePositionChecked() (int, int, error) {
	var pt pointStruct
	ret, _, err := procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))
	if ret == 0 {
		return 0, 0, fmt.Errorf("GetCursorPos failed: %v", err)
	}
	return int(pt.X), int(pt.Y), nil
}

// getMousePosition is the Windows side of the per-OS mouse-position seam
// used by the shared screenshot pipeline in screenshot.go, where cursor
// drawing is best-effort. Mouse API handlers use getMousePositionChecked.
func getMousePosition() (int, int) {
	x, y, _ := getMousePositionChecked()
	return x, y
}

func setMousePosition(x, y int) error {
	ret, _, err := procSetCursorPos.Call(uintptr(int32(x)), uintptr(int32(y)))
	if ret == 0 {
		return fmt.Errorf("SetCursorPos failed: %v", err)
	}
	return nil
}

// buttonFlags maps a canonical button name (as produced by
// normalizeMouseButton) to (downFlag, upFlag).
func buttonFlags(button string) (uint32, uint32, error) {
	switch button {
	case "left":
		return mouseEventF_LEFTDOWN, mouseEventF_LEFTUP, nil
	case "right":
		return mouseEventF_RIGHTDOWN, mouseEventF_RIGHTUP, nil
	case "middle":
		return mouseEventF_MIDDLEDOWN, mouseEventF_MIDDLEUP, nil
	default:
		return 0, 0, fmt.Errorf("unsupported mouse button %q: expected one of left, right, middle", button)
	}
}

func mouseClick(button string, double bool) error {
	// Mirror keyTap below: a button successfully pressed must be released
	// even when a later SendInput fails (UIPI, secure desktop, locked
	// session), or it stays logically held system-wide and turns every
	// subsequent move into a drag. The release here is best-effort; the
	// success path clears held as it releases with error propagation.
	held := false
	defer func() {
		if held {
			_ = mouseUp(button)
		}
	}()
	if err := mouseDown(button); err != nil {
		return err
	}
	held = true
	time.Sleep(20 * time.Millisecond)
	if err := mouseUp(button); err != nil {
		return err
	}
	held = false
	if double {
		time.Sleep(50 * time.Millisecond)
		if err := mouseDown(button); err != nil {
			return err
		}
		held = true
		time.Sleep(20 * time.Millisecond)
		if err := mouseUp(button); err != nil {
			return err
		}
		held = false
	}
	return nil
}

func mouseDown(button string) error {
	down, _, err := buttonFlags(button)
	if err != nil {
		return err
	}
	inputs := []inputStruct{{}}
	inputs[0].asMouse(mouseInput{DwFlags: down})
	return sendInputs(inputs)
}

func mouseUp(button string) error {
	_, up, err := buttonFlags(button)
	if err != nil {
		return err
	}
	inputs := []inputStruct{{}}
	inputs[0].asMouse(mouseInput{DwFlags: up})
	return sendInputs(inputs)
}

// maxScrollAmount bounds a single scroll request. Past ~17.9M notches the
// wheel-delta product below overflows int32 and silently flips the scroll
// direction; any amount remotely near that is garbage input. Keep in sync
// with the identical bound in mouse.go (linux).
const maxScrollAmount = 10000

// mouseScroll scrolls by `amount` wheel notches in canonical `direction`
// ("up" or "down", as produced by normalizeScrollDirection).
func mouseScroll(amount int, direction string) error {
	if amount > maxScrollAmount {
		return fmt.Errorf("scroll amount %d exceeds maximum of %d", amount, maxScrollAmount)
	}
	delta := int32(wheelDelta * amount)
	switch direction {
	case scrollDirectionUp:
		// Positive delta scrolls up.
	case scrollDirectionDown:
		delta = -delta
	default:
		return fmt.Errorf("unsupported scroll direction %q: expected up or down", direction)
	}
	inputs := []inputStruct{{}}
	inputs[0].asMouse(mouseInput{
		MouseData: uint32(delta),
		DwFlags:   mouseEventF_WHEEL,
	})
	return sendInputs(inputs)
}

// ---------------------------------------------------------------------------
// Keyboard
// ---------------------------------------------------------------------------

// resolveKey maps a key name (e.g. "a", "enter", "f5") to a virtual-key code
// plus whether the key requires KEYEVENTF_EXTENDEDKEY.
//
// Returns (vk, extended, ok). When ok is false the caller should fall back
// to typing the literal characters as Unicode.
func resolveKey(name string) (uint16, bool, bool) {
	lower := strings.ToLower(name)
	if vk, ok := virtualKeyCodes[lower]; ok {
		return vk, extendedVirtualKeys[lower], true
	}
	// Single printable character — use VkKeyScanW to translate. VkKeyScanW
	// takes a WCHAR, so only BMP runes are eligible; a non-BMP rune must
	// return ok=false so the caller takes the Unicode fallback (typed as a
	// surrogate pair) instead of silently truncating to a colliding BMP
	// code unit (e.g. U+10030 -> 0x0030, the '0' key).
	runes := []rune(name)
	if len(runes) == 1 && runes[0] <= 0xFFFF {
		ret, _, _ := procVkKeyScanW.Call(uintptr(uint16(runes[0])))
		// Low byte = VK code. -1 means no translation.
		if int16(ret) != -1 {
			return uint16(ret & 0xFF), false, true
		}
	}
	return 0, false, false
}

// keyPress sends a single KEYDOWN or KEYUP event for a virtual key.
func keyPress(vk uint16, extended, up bool) error {
	flags := uint32(0)
	if extended {
		flags |= keyEventF_EXTENDEDKEY
	}
	if up {
		flags |= keyEventF_KEYUP
	}
	// Pre-compute a scan code so games / apps that look at lParam scan bits
	// see consistent values.
	scan, _, _ := procMapVirtualKeyW.Call(uintptr(vk), mapVkVkToVsc)
	inputs := []inputStruct{{}}
	inputs[0].asKeyboard(keybdInput{
		WVk:     vk,
		WScan:   uint16(scan),
		DwFlags: flags,
	})
	return sendInputs(inputs)
}

// keyTap presses `key` while holding any modifier keys, then releases all.
func keyTap(key string, modifiers []string) error {
	// Resolve modifiers.
	type resolved struct {
		vk       uint16
		extended bool
	}
	mods := make([]resolved, 0, len(modifiers))
	for _, m := range modifiers {
		vk, ext, ok := resolveKey(m)
		if !ok {
			return fmt.Errorf("unknown modifier: %q", m)
		}
		mods = append(mods, resolved{vk, ext})
	}

	// Every key successfully pressed must be released even when a later
	// SendInput fails (UIPI, secure desktop, locked session), or it stays
	// logically held system-wide and corrupts all subsequent input. Releases
	// here are best-effort; the success path below pops keys as it releases
	// them with error propagation.
	var pressed []resolved
	defer func() {
		for i := len(pressed) - 1; i >= 0; i-- {
			_ = keyPress(pressed[i].vk, pressed[i].extended, true)
		}
	}()

	// Press modifiers down.
	for _, mod := range mods {
		if err := keyPress(mod.vk, mod.extended, false); err != nil {
			return err
		}
		pressed = append(pressed, mod)
	}

	// Tap the main key.
	if vk, ext, ok := resolveKey(key); ok {
		if err := keyPress(vk, ext, false); err != nil {
			return err
		}
		pressed = append(pressed, resolved{vk, ext})
		time.Sleep(10 * time.Millisecond)
		if err := keyPress(vk, ext, true); err != nil {
			return err
		}
		pressed = pressed[:len(pressed)-1]
	} else {
		// Last-ditch fallback: type as a single unicode rune.
		if err := typeString(key, 0); err != nil {
			return err
		}
	}

	// Release modifiers in reverse order.
	for i := len(pressed) - 1; i >= 0; i-- {
		if err := keyPress(pressed[i].vk, pressed[i].extended, true); err != nil {
			return err
		}
		pressed = pressed[:i]
	}
	return nil
}

// typeString types the given text via Unicode key events. delay is the
// per-character delay in milliseconds (0 = no delay).
func typeString(text string, delay int) error {
	if text == "" {
		return nil
	}
	// Each rune's UTF-16 code units go out in a single SendInput call:
	// within one call the events cannot interleave with other synthesized
	// input, so a non-BMP rune's surrogate halves always arrive paired and
	// the delay applies once per character, not per code unit.
	inputs := make([]inputStruct, 0, 4)
	for _, r := range text {
		inputs = appendRuneInputs(inputs[:0], r)
		if err := sendInputs(inputs); err != nil {
			return err
		}
		if delay > 0 {
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}
	return nil
}

// appendRuneInputs appends the KEYEVENTF_UNICODE down/up events that type r:
// one down/up pair per UTF-16 code unit, so a non-BMP rune contributes four
// events (its high then low surrogate).
func appendRuneInputs(inputs []inputStruct, r rune) []inputStruct {
	var units [2]uint16
	for _, cu := range utf16.AppendRune(units[:0], r) {
		var down, up inputStruct
		down.asKeyboard(keybdInput{
			WScan:   cu,
			DwFlags: keyEventF_UNICODE,
		})
		up.asKeyboard(keybdInput{
			WScan:   cu,
			DwFlags: keyEventF_UNICODE | keyEventF_KEYUP,
		})
		inputs = append(inputs, down, up)
	}
	return inputs
}

// ---------------------------------------------------------------------------
// Display / windows enumeration
// ---------------------------------------------------------------------------

// windowInfo describes a top-level window enumerated by getWindowsList.
type windowInfo struct {
	HWND    uintptr
	Title   string
	Visible bool
	X       int
	Y       int
	Width   int
	Height  int
}

// EnumWindows invokes its callback synchronously on the calling thread, and
// syscall.NewCallback allocations are permanent (the runtime caps a process
// at ~2000 callbacks), so a single package-level callback feeds a
// mutex-guarded accumulator instead of compiling a fresh closure per call.
var (
	enumWindowsMu        sync.Mutex
	enumWindowsCollected []windowInfo
	enumWindowsCallback  = syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
		visibleRet, _, _ := procIsWindowVisible.Call(hwnd)
		visible := visibleRet != 0

		// Length of window title in chars.
		titleLen, _, _ := procGetWindowTextLenW.Call(hwnd)
		title := ""
		if titleLen > 0 {
			buf := make([]uint16, int(titleLen)+1)
			procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
			title = windows.UTF16ToString(buf)
		}

		// Position / size via GetWindowRect.
		var r rectStruct
		x, y, w, h := 0, 0, 0, 0
		ret, _, _ := procGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&r)))
		if ret != 0 {
			x = int(r.Left)
			y = int(r.Top)
			w = int(r.Right - r.Left)
			h = int(r.Bottom - r.Top)
		}

		// Skip windows with empty titles and that are invisible — they're
		// usually internal Win32 helpers that pollute the listing.
		if title != "" || visible {
			enumWindowsCollected = append(enumWindowsCollected, windowInfo{
				HWND:    hwnd,
				Title:   title,
				Visible: visible,
				X:       x,
				Y:       y,
				Width:   w,
				Height:  h,
			})
		}
		return 1 // continue enumeration
	})
)

// getWindowsList enumerates all top-level windows. The enumeration callback
// always returns 1, so a zero return from EnumWindows is a real failure
// (e.g. inaccessible desktop), not an aborted walk — report it instead of
// passing off an empty desktop.
func getWindowsList() ([]windowInfo, error) {
	enumWindowsMu.Lock()
	defer enumWindowsMu.Unlock()

	enumWindowsCollected = nil
	ret, _, err := procEnumWindows.Call(enumWindowsCallback, 0)
	collected := enumWindowsCollected
	enumWindowsCollected = nil
	if ret == 0 {
		return nil, fmt.Errorf("EnumWindows failed: %v", err)
	}
	return collected, nil
}

// getForegroundWindow returns the HWND of the current foreground window, or
// 0 when no window has focus (e.g. a screensaver or secure desktop is up).
func getForegroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWnd.Call()
	return hwnd
}
