// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeKeyboardPress(t *testing.T) {
	t.Run("normalizes common aliases and uppercase letters", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("Escape", []string{"CONTROL", "meta"})
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "escape",
			modifiers: []string{"ctrl", "cmd"},
		}, chord)
	})

	t.Run("normalizes command and option modifiers", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("c", []string{"command"})
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "c",
			modifiers: []string{"cmd"},
		}, chord)

		chord, err = normalizeKeyboardPress("tab", []string{"option"})
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "tab",
			modifiers: []string{"alt"},
		}, chord)
	})

	t.Run("normalizes cmd-class modifier aliases to cmd", func(t *testing.T) {
		for _, alias := range []string{"cmd", "command", "super", "meta", "win", "windows"} {
			chord, err := normalizeKeyboardPress("c", []string{alias})
			require.NoError(t, err, "alias %q", alias)
			assert.Equal(t, normalizedChord{
				key:       "c",
				modifiers: []string{"cmd"},
			}, chord, "alias %q", alias)
		}
	})

	t.Run("normalizes return and escape main key aliases", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("Return", nil)
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{key: "enter"}, chord)

		chord, err = normalizeKeyboardPress("esc", nil)
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{key: "escape"}, chord)
	})

	t.Run("normalizes uppercase printable letters without inferring shift", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("A", nil)
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{key: "a"}, chord)
	})

	t.Run("supports modifier only chords", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("lshift", []string{"ctrl"})
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "lshift",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("supports grammar safe numpad names", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("num_plus", nil)
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{key: "num+"}, chord)
	})

	t.Run("supports unshifted punctuation keys", func(t *testing.T) {
		chord, err := normalizeKeyboardPress("/", []string{"ctrl"})
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "/",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("rejects empty keys", func(t *testing.T) {
		_, err := normalizeKeyboardPress("", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "": empty key token`, err.Error())
	})

	t.Run("rejects shifted symbol shorthand", func(t *testing.T) {
		_, err := normalizeKeyboardPress("!", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "!"; use press("1", ["shift"]) for a shifted key or type("!") for text`, err.Error())
	})

	t.Run("rejects whitespace shorthand", func(t *testing.T) {
		_, err := normalizeKeyboardPress(" ", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key " "; use press("space") or type(" ")`, err.Error())
	})

	t.Run("rejects tab shorthand", func(t *testing.T) {
		_, err := normalizeKeyboardPress("\t", nil)
		require.Error(t, err)
		assert.Equal(t, "unsupported key \"\\t\"; use press(\"tab\") or type(\"\\t\")", err.Error())
	})

	t.Run("rejects unicode characters", func(t *testing.T) {
		_, err := normalizeKeyboardPress("é", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "é"; use type("é") for text input`, err.Error())
	})

	t.Run("rejects unsupported keys", func(t *testing.T) {
		_, err := normalizeKeyboardPress("FooBar", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "FooBar"; see supported keyboard key names in the computer-use docs`, err.Error())
	})

	t.Run("rejects removed media and brightness keys", func(t *testing.T) {
		_, err := normalizeKeyboardPress("audio_mute", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "audio_mute"; see supported keyboard key names in the computer-use docs`, err.Error())

		_, err = normalizeKeyboardPress("lights_mon_up", nil)
		require.Error(t, err)
		assert.Equal(t, `unsupported key "lights_mon_up"; see supported keyboard key names in the computer-use docs`, err.Error())
	})

	t.Run("rejects non modifier values in modifiers", func(t *testing.T) {
		_, err := normalizeKeyboardPress("d", []string{"ctrl", "c"})
		require.Error(t, err)
		assert.Equal(t, `unsupported modifier "c"; supported modifiers: ctrl, alt, shift, cmd`, err.Error())
	})

	t.Run("rejects sided modifiers in modifier position", func(t *testing.T) {
		_, err := normalizeKeyboardPress("a", []string{"lshift"})
		require.Error(t, err)
		assert.Equal(
			t,
			`unsupported modifier "lshift"; supported modifiers: ctrl, alt, shift, cmd; left/right-specific modifier keys are only supported as the main key`,
			err.Error(),
		)
	})

	t.Run("rejects duplicate modifiers", func(t *testing.T) {
		_, err := normalizeKeyboardPress("c", []string{"ctrl", "ctrl"})
		require.Error(t, err)
		assert.Equal(t, `duplicate modifier "ctrl" in press("c", ["ctrl", "ctrl"])`, err.Error())
	})

	t.Run("rejects duplicate modifiers after normalization", func(t *testing.T) {
		_, err := normalizeKeyboardPress("c", []string{"cmd", "meta"})
		require.Error(t, err)
		assert.Equal(t, `duplicate modifier usage after normalization in press("c", ["cmd", "meta"])`, err.Error())
	})

	t.Run("rejects duplicate modifier usage when key is itself a modifier", func(t *testing.T) {
		_, err := normalizeKeyboardPress("shift", []string{"shift"})
		require.Error(t, err)
		assert.Equal(t, `duplicate modifier usage after normalization in press("shift", ["shift"])`, err.Error())
	})
}

func TestNormalizeKeyboardHotkey(t *testing.T) {
	t.Run("normalizes a standard hotkey chord", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("CTRL+Escape")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "escape",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("normalizes alias modifiers in hotkeys", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("command+c")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "c",
			modifiers: []string{"cmd"},
		}, chord)

		chord, err = normalizeKeyboardHotkey("option+tab")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "tab",
			modifiers: []string{"alt"},
		}, chord)

		chord, err = normalizeKeyboardHotkey("win+e")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "e",
			modifiers: []string{"cmd"},
		}, chord)
	})

	t.Run("normalizes main key aliases in hotkeys", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("ctrl+return")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "enter",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("supports modifier only chords", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("ctrl+shift")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "shift",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("supports single token hotkeys", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("lshift")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{key: "lshift"}, chord)
	})

	t.Run("trims surrounding whitespace around separators", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey(" ctrl + c ")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "c",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("supports punctuation and numpad tokens", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("ctrl+num_plus")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "num+",
			modifiers: []string{"ctrl"},
		}, chord)

		chord, err = normalizeKeyboardHotkey("ctrl+/")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "/",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("normalizes uppercase letters without adding shift", func(t *testing.T) {
		chord, err := normalizeKeyboardHotkey("ctrl+A")
		require.NoError(t, err)
		assert.Equal(t, normalizedChord{
			key:       "a",
			modifiers: []string{"ctrl"},
		}, chord)
	})

	t.Run("rejects duplicate separators", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("ctrl++c")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "ctrl++c": empty key token`, err.Error())
	})

	t.Run("rejects empty token after trimming", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("ctrl+ ")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "ctrl+ ": empty key token after trimming`, err.Error())
	})

	t.Run("rejects duplicate modifiers", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("ctrl+ctrl+c")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "ctrl+ctrl+c": duplicate modifier "ctrl" in hotkey("ctrl+ctrl+c")`, err.Error())
	})

	t.Run("rejects duplicate modifiers after normalization", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("cmd+meta+c")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "cmd+meta+c": duplicate modifier usage after normalization in hotkey("cmd+meta+c")`, err.Error())
	})

	t.Run("rejects duplicate modifier usage between modifiers and key", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("shift+shift")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "shift+shift": duplicate modifier usage after normalization in hotkey("shift+shift")`, err.Error())
	})

	t.Run("rejects sided modifiers in modifier position", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("lshift+a")
		require.Error(t, err)
		assert.Equal(
			t,
			`invalid hotkey "lshift+a": unsupported modifier "lshift"; supported modifiers: ctrl, alt, shift, cmd; left/right-specific modifier keys are only supported as the main key`,
			err.Error(),
		)
	})

	t.Run("rejects multiple non modifier keys", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("ctrl+c+d")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "ctrl+c+d": chords may contain at most one non-modifier key`, err.Error())

		_, err = normalizeKeyboardHotkey("a+b")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "a+b": chords may contain at most one non-modifier key`, err.Error())
	})

	t.Run("rejects multi chord strings", func(t *testing.T) {
		_, err := normalizeKeyboardHotkey("ctrl+k ctrl+c")
		require.Error(t, err)
		assert.Equal(t, `invalid hotkey "ctrl+k ctrl+c": only a single chord is supported`, err.Error())
	})
}
