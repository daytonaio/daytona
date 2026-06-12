//go:build windows

// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"math"
	"syscall"
	"testing"
	"unicode/utf16"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every canonical token the keyboard normalizer can emit for a named key must
// have a virtual-key mapping; otherwise keyTap would silently fall back to
// typing the token as literal text.
func TestVirtualKeyCodesCoverNormalizedTokens(t *testing.T) {
	for name, key := range keyExecutionMap {
		_, ok := virtualKeyCodes[key.token]
		assert.True(t, ok, "normalized token %q (from key name %q) has no virtual-key mapping", key.token, name)
	}
}

func TestResolveKeyNormalizedModifiers(t *testing.T) {
	expected := map[string]uint16{
		"ctrl":  0x11,
		"alt":   0x12,
		"shift": 0x10,
		"cmd":   0x5B,
	}
	for _, modifier := range supportedModifierNames {
		vk, _, ok := resolveKey(modifier)
		require.True(t, ok, "modifier %q must resolve", modifier)
		assert.Equal(t, expected[modifier], vk, "modifier %q", modifier)
	}
}

func TestResolveKeyConsumesNormalizedChords(t *testing.T) {
	t.Run("press chords from the alias matrix resolve", func(t *testing.T) {
		cases := []struct {
			key       string
			modifiers []string
		}{
			{"c", []string{"command"}},
			{"tab", []string{"option"}},
			{"e", []string{"super"}},
			{"e", []string{"meta"}},
			{"e", []string{"win"}},
			{"Return", nil},
			{"Escape", nil},
			{"num_plus", nil},
		}
		for _, tc := range cases {
			chord, err := normalizeKeyboardPress(tc.key, tc.modifiers)
			require.NoError(t, err, "press(%q, %v)", tc.key, tc.modifiers)

			_, _, ok := resolveKey(chord.key)
			assert.True(t, ok, "press(%q, %v): normalized key %q must resolve", tc.key, tc.modifiers, chord.key)
			for _, modifier := range chord.modifiers {
				_, _, ok := resolveKey(modifier)
				assert.True(t, ok, "press(%q, %v): normalized modifier %q must resolve", tc.key, tc.modifiers, modifier)
			}
		}
	})

	t.Run("hotkey chords from the alias matrix resolve", func(t *testing.T) {
		for _, raw := range []string{"command+c", "option+tab", "win+e", "ctrl+return", "ctrl+num_plus"} {
			chord, err := normalizeKeyboardHotkey(raw)
			require.NoError(t, err, "hotkey(%q)", raw)

			_, _, ok := resolveKey(chord.key)
			assert.True(t, ok, "hotkey(%q): normalized key %q must resolve", raw, chord.key)
			for _, modifier := range chord.modifiers {
				_, _, ok := resolveKey(modifier)
				assert.True(t, ok, "hotkey(%q): normalized modifier %q must resolve", raw, modifier)
			}
		}
	})
}

// Numpad Enter shares VK_RETURN with the main Enter key; the extended-key
// flag is the only thing that distinguishes the two in a SendInput event.
func TestResolveKeyExtendedFlag(t *testing.T) {
	vkEnter, ext, ok := resolveKey("enter")
	require.True(t, ok, "enter must resolve")
	assert.False(t, ext, "main enter must not be extended")

	vkNumEnter, ext, ok := resolveKey("num_enter")
	require.True(t, ok, "num_enter must resolve")
	assert.True(t, ext, "numpad enter must be extended")
	assert.Equal(t, vkEnter, vkNumEnter, "enter and num_enter share VK_RETURN")

	for token := range extendedVirtualKeys {
		_, ext, ok := resolveKey(token)
		require.True(t, ok, "extended token %q must resolve", token)
		assert.True(t, ext, "token %q must report extended", token)
	}
}

// Every extended-key token must be a canonical virtualKeyCodes token,
// otherwise the entry is dead (resolveKey would never look it up).
func TestExtendedVirtualKeysAreKnownTokens(t *testing.T) {
	for token := range extendedVirtualKeys {
		_, ok := virtualKeyCodes[token]
		assert.True(t, ok, "extended token %q has no virtual-key mapping", token)
	}
}

// inputStruct is handed to SendInput verbatim and reinterpreted through
// unsafe casts, so its layout must match the 64-bit Win32 INPUT exactly:
// 40 bytes total, union at offset 8, and 8-byte alignment (mouseInput and
// keybdInput both carry a pointer-aligned ExtraInfo field).
func TestInputStructLayout(t *testing.T) {
	var in inputStruct
	assert.Equal(t, uintptr(40), unsafe.Sizeof(in), "INPUT must be 40 bytes on 64-bit Windows")
	assert.Equal(t, uintptr(8), unsafe.Offsetof(in.U), "union must start at offset 8")
	assert.Equal(t, uintptr(8), unsafe.Alignof(in), "INPUT must be 8-byte aligned for the unsafe union casts")
}

// buttonFlags receives canonical names from normalizeMouseButton; anything
// else — including the raw spellings normalization handles — must error.
func TestButtonFlags(t *testing.T) {
	leftDown, leftUp, err := buttonFlags("left")
	require.NoError(t, err)
	assert.Equal(t, uint32(mouseEventF_LEFTDOWN), leftDown)
	assert.Equal(t, uint32(mouseEventF_LEFTUP), leftUp)

	rightDown, rightUp, err := buttonFlags("right")
	require.NoError(t, err)
	assert.Equal(t, uint32(mouseEventF_RIGHTDOWN), rightDown)
	assert.Equal(t, uint32(mouseEventF_RIGHTUP), rightUp)

	middleDown, middleUp, err := buttonFlags("middle")
	require.NoError(t, err)
	assert.Equal(t, uint32(mouseEventF_MIDDLEDOWN), middleDown)
	assert.Equal(t, uint32(mouseEventF_MIDDLEUP), middleUp)

	for _, raw := range []string{"", " left ", "Left", "center", "wheel"} {
		_, _, err := buttonFlags(raw)
		assert.Errorf(t, err, "non-canonical button %q must be rejected", raw)
	}
}

// mouseScroll must reject non-canonical directions instead of treating them
// as scroll-up; only normalizeScrollDirection output is accepted.
func TestMouseScrollRejectsNonCanonicalDirection(t *testing.T) {
	for _, raw := range []string{"", "left", "Up", " down "} {
		err := mouseScroll(1, raw)
		assert.Errorf(t, err, "non-canonical direction %q must be rejected", raw)
	}
}

// getWindowsList must not compile a new syscall callback per call: the Go
// runtime caps a process at ~2000 callbacks and never releases them, so a
// per-call syscall.NewCallback panics under sustained GetWindows polling.
func TestGetWindowsListDoesNotLeakCallbacks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping 2100-iteration callback-leak regression in -short mode")
	}
	for i := 0; i < 2100; i++ {
		if _, err := getWindowsList(); err != nil {
			t.Fatalf("getWindowsList: %v", err)
		}
	}
}

// resolveKey feeds VkKeyScanW a WCHAR, so a non-BMP rune cannot be resolved
// to a virtual key: truncating it to 16 bits would collide with an unrelated
// BMP character (U+10030 -> 0x0030 -> the '0' key). It must report ok=false
// so callers take the Unicode surrogate-pair fallback instead.
func TestResolveKeyRejectsNonBMPRunes(t *testing.T) {
	for _, name := range []string{"\U00010030", "\U0001F600"} {
		_, _, ok := resolveKey(name)
		assert.Falsef(t, ok, "non-BMP rune %q must not resolve to a virtual key", name)
	}
}

// mouseScroll must reject out-of-range amounts: each notch becomes one
// SendInput event, so an unbounded amount would synthesize an absurd input
// batch from a single request, and a negative amount has no event count.
func TestMouseScrollRejectsExcessiveAmount(t *testing.T) {
	assert.Error(t, mouseScroll(maxScrollAmount+1, scrollDirectionUp))
	assert.Error(t, mouseScroll(maxScrollAmount+1, scrollDirectionDown))
	assert.Error(t, mouseScroll(-1, scrollDirectionUp))
}

// Windows consumers read a wheel event's delta as signed 16-bit
// (WM_MOUSEWHEEL: (short)HIWORD(wParam); raw input: RAWMOUSE.usButtonData as
// a short), so an event with |delta| > 32767 is truncated mod 2^16 and can
// flip the scroll direction (120*274 = 32880 reads back as -32656). Every
// event scrollInputs generates must stay within the int16 ceiling, carry the
// direction's sign, and the batch must sum to the requested notches.
func TestScrollInputsStayWithinInt16WheelDelta(t *testing.T) {
	wheelEventDelta := func(in inputStruct) int32 {
		require.Equal(t, uint32(inputMouse), in.Type)
		mi := *(*mouseInput)(unsafe.Pointer(&in.U[0]))
		require.Equal(t, uint32(mouseEventF_WHEEL), mi.DwFlags)
		return int32(mi.MouseData)
	}

	for direction, sign := range map[string]int64{
		scrollDirectionUp:   1,
		scrollDirectionDown: -1,
	} {
		inputs, err := scrollInputs(maxScrollAmount, direction)
		require.NoError(t, err, "direction %q", direction)

		var total int64
		for i, in := range inputs {
			delta := wheelEventDelta(in)
			require.LessOrEqualf(t, delta, int32(math.MaxInt16), "direction %q event %d exceeds the int16 wheel ceiling", direction, i)
			require.GreaterOrEqualf(t, delta, int32(math.MinInt16), "direction %q event %d exceeds the int16 wheel ceiling", direction, i)
			require.Truef(t, sign*int64(delta) > 0, "direction %q event %d: delta %d must carry the direction's sign", direction, i, delta)
			total += int64(delta)
		}
		assert.Equal(t, sign*int64(wheelDelta)*int64(maxScrollAmount), total, "direction %q: batch must sum to the requested notches", direction)
	}
}

// typeString batches each rune's events into one SendInput call so a non-BMP
// rune's surrogate halves cannot be split by other synthesized input.
// appendRuneInputs builds those batches: a BMP rune yields one down/up pair,
// a non-BMP rune yields down/up pairs for its high then low surrogate.
func TestAppendRuneInputs(t *testing.T) {
	type event struct {
		scan  uint16
		flags uint32
	}
	keyEvent := func(in inputStruct) event {
		require.Equal(t, uint32(inputKeyboard), in.Type)
		ki := *(*keybdInput)(unsafe.Pointer(&in.U[0]))
		return event{ki.WScan, ki.DwFlags}
	}

	inputs := appendRuneInputs(nil, 'a')
	require.Len(t, inputs, 2, "BMP rune must produce one down/up pair")
	assert.Equal(t, event{'a', keyEventF_UNICODE}, keyEvent(inputs[0]))
	assert.Equal(t, event{'a', keyEventF_UNICODE | keyEventF_KEYUP}, keyEvent(inputs[1]))

	const grin = '\U0001F600'
	hi, lo := utf16.EncodeRune(grin)
	inputs = appendRuneInputs(inputs[:0], grin)
	require.Len(t, inputs, 4, "non-BMP rune must produce down/up pairs for both surrogate halves")
	assert.Equal(t, event{uint16(hi), keyEventF_UNICODE}, keyEvent(inputs[0]))
	assert.Equal(t, event{uint16(hi), keyEventF_UNICODE | keyEventF_KEYUP}, keyEvent(inputs[1]))
	assert.Equal(t, event{uint16(lo), keyEventF_UNICODE}, keyEvent(inputs[2]))
	assert.Equal(t, event{uint16(lo), keyEventF_UNICODE | keyEventF_KEYUP}, keyEvent(inputs[3]))
}

// SendInput blocked by UIPI (or an inaccessible desktop) returns 0 without
// setting a last error, so LazyProc.Call hands back syscall.Errno(0) whose
// message is "The operation completed successfully." — the error must not
// surface that contradiction on the documented dominant failure mode.
func TestSendInputErrorZeroErrnoNamesBlockedInput(t *testing.T) {
	err := sendInputError(0, 1, syscall.Errno(0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "0/1")
	assert.Contains(t, err.Error(), "UIPI")
	assert.NotContains(t, err.Error(), "completed successfully")

	// Genuine errnos keep their system message.
	denied := syscall.Errno(5) // ERROR_ACCESS_DENIED
	err = sendInputError(0, 2, denied)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "0/2")
	assert.Contains(t, err.Error(), denied.Error())
}
