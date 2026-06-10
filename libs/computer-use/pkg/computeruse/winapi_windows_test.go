//go:build windows

// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"testing"

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
		vk, ok := resolveKey(modifier)
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

			_, ok := resolveKey(chord.key)
			assert.True(t, ok, "press(%q, %v): normalized key %q must resolve", tc.key, tc.modifiers, chord.key)
			for _, modifier := range chord.modifiers {
				_, ok := resolveKey(modifier)
				assert.True(t, ok, "press(%q, %v): normalized modifier %q must resolve", tc.key, tc.modifiers, modifier)
			}
		}
	})

	t.Run("hotkey chords from the alias matrix resolve", func(t *testing.T) {
		for _, raw := range []string{"command+c", "option+tab", "win+e", "ctrl+return", "ctrl+num_plus"} {
			chord, err := normalizeKeyboardHotkey(raw)
			require.NoError(t, err, "hotkey(%q)", raw)

			_, ok := resolveKey(chord.key)
			assert.True(t, ok, "hotkey(%q): normalized key %q must resolve", raw, chord.key)
			for _, modifier := range chord.modifiers {
				_, ok := resolveKey(modifier)
				assert.True(t, ok, "hotkey(%q): normalized modifier %q must resolve", raw, modifier)
			}
		}
	})
}
