// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

type normalizedKey struct {
	token          string
	modifierFamily string
}

type normalizedChord struct {
	key       string
	modifiers []string
}

var supportedModifierNames = []string{"ctrl", "alt", "shift", "cmd"}

var modifierAliasMap = map[string]string{
	"alt":     "alt",
	"cmd":     "cmd",
	"command": "cmd",
	"control": "ctrl",
	"ctrl":    "ctrl",
	"meta":    "cmd",
	"option":  "alt",
	"shift":   "shift",
	"super":   "cmd",
	"win":     "cmd",
	"windows": "cmd",
}

var mainKeyAliasMap = map[string]string{
	"cmd":           "cmd",
	"command":       "cmd",
	"control":       "ctrl",
	"ctrl":          "ctrl",
	"del":           "delete",
	"esc":           "escape",
	"left_alt":      "lalt",
	"left_cmd":      "lcmd",
	"left_command":  "lcmd",
	"left_control":  "lctrl",
	"left_ctrl":     "lctrl",
	"left_shift":    "lshift",
	"meta":          "cmd",
	"option":        "alt",
	"page_down":     "pagedown",
	"page_up":       "pageup",
	"return":        "enter",
	"right_alt":     "ralt",
	"right_cmd":     "rcmd",
	"right_command": "rcmd",
	"right_control": "rctrl",
	"right_ctrl":    "rctrl",
	"right_shift":   "rshift",
	"spacebar":      "space",
	"super":         "cmd",
	"win":           "cmd",
	"windows":       "cmd",
}

var keyExecutionMap = map[string]normalizedKey{
	"alt":          {token: "alt", modifierFamily: "alt"},
	"backspace":    {token: "backspace"},
	"capslock":     {token: "capslock"},
	"cmd":          {token: "cmd", modifierFamily: "cmd"},
	"ctrl":         {token: "ctrl", modifierFamily: "ctrl"},
	"delete":       {token: "delete"},
	"down":         {token: "down"},
	"end":          {token: "end"},
	"enter":        {token: "enter"},
	"escape":       {token: "escape"},
	"f1":           {token: "f1"},
	"f2":           {token: "f2"},
	"f3":           {token: "f3"},
	"f4":           {token: "f4"},
	"f5":           {token: "f5"},
	"f6":           {token: "f6"},
	"f7":           {token: "f7"},
	"f8":           {token: "f8"},
	"f9":           {token: "f9"},
	"f10":          {token: "f10"},
	"f11":          {token: "f11"},
	"f12":          {token: "f12"},
	"f13":          {token: "f13"},
	"f14":          {token: "f14"},
	"f15":          {token: "f15"},
	"f16":          {token: "f16"},
	"f17":          {token: "f17"},
	"f18":          {token: "f18"},
	"f19":          {token: "f19"},
	"f20":          {token: "f20"},
	"f21":          {token: "f21"},
	"f22":          {token: "f22"},
	"f23":          {token: "f23"},
	"f24":          {token: "f24"},
	"home":         {token: "home"},
	"insert":       {token: "insert"},
	"lalt":         {token: "lalt", modifierFamily: "alt"},
	"lcmd":         {token: "lcmd", modifierFamily: "cmd"},
	"lctrl":        {token: "lctrl", modifierFamily: "ctrl"},
	"left":         {token: "left"},
	"lshift":       {token: "lshift", modifierFamily: "shift"},
	"menu":         {token: "menu"},
	"num0":         {token: "num0"},
	"num1":         {token: "num1"},
	"num2":         {token: "num2"},
	"num3":         {token: "num3"},
	"num4":         {token: "num4"},
	"num5":         {token: "num5"},
	"num6":         {token: "num6"},
	"num7":         {token: "num7"},
	"num8":         {token: "num8"},
	"num9":         {token: "num9"},
	"num_asterisk": {token: "num*"},
	"num_decimal":  {token: "num."},
	"num_enter":    {token: "num_enter"},
	"num_equal":    {token: "num_equal"},
	"num_lock":     {token: "num_lock"},
	"num_minus":    {token: "num-"},
	"num_plus":     {token: "num+"},
	"num_slash":    {token: "num/"},
	"pagedown":     {token: "pagedown"},
	"pageup":       {token: "pageup"},
	"ralt":         {token: "ralt", modifierFamily: "alt"},
	"rcmd":         {token: "rcmd", modifierFamily: "cmd"},
	"rctrl":        {token: "rctrl", modifierFamily: "ctrl"},
	"right":        {token: "right"},
	"rshift":       {token: "rshift", modifierFamily: "shift"},
	"shift":        {token: "shift", modifierFamily: "shift"},
	"space":        {token: "space"},
	"tab":          {token: "tab"},
	"up":           {token: "up"},
}

var shiftedSymbolMap = map[rune]string{
	'!': "1",
	'"': "'",
	'#': "3",
	'$': "4",
	'%': "5",
	'&': "7",
	'(': "9",
	')': "0",
	'*': "8",
	'+': "=",
	':': ";",
	'<': ",",
	'>': ".",
	'?': "/",
	'@': "2",
	'^': "6",
	'_': "-",
	'{': "[",
	'|': "\\",
	'}': "]",
	'~': "`",
}

var allowedPunctuationKeys = map[rune]struct{}{
	'\'': {},
	',':  {},
	'-':  {},
	'.':  {},
	'/':  {},
	';':  {},
	'=':  {},
	'[':  {},
	'\\': {},
	']':  {},
	'`':  {},
}

var disallowedNumpadShorthand = map[string]string{
	"num*": "num_asterisk",
	"num+": "num_plus",
	"num-": "num_minus",
	"num.": "num_decimal",
	"num/": "num_slash",
}

func normalizeKeyboardPress(key string, modifiers []string) (normalizedChord, error) {
	normalizedKey, err := normalizeKeyboardKey(key)
	if err != nil {
		return normalizedChord{}, err
	}

	normalizedModifiers, err := normalizeKeyboardModifiers(modifiers, pressContext(key, modifiers))
	if err != nil {
		return normalizedChord{}, err
	}

	if err := ensureNoModifierConflicts(normalizedKey, normalizedModifiers, pressContext(key, modifiers)); err != nil {
		return normalizedChord{}, err
	}

	chord := normalizedChord{key: normalizedKey.token}
	if len(normalizedModifiers) > 0 {
		chord.modifiers = normalizedModifiers
	}
	return chord, nil
}

func normalizeKeyboardHotkey(raw string) (normalizedChord, error) {
	originalRaw := raw
	if strings.TrimSpace(raw) == "" {
		return normalizedChord{}, fmt.Errorf(`invalid hotkey %q: empty key token`, raw)
	}

	parts := strings.Split(raw, "+")
	trimmedParts := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			reason := "empty key token"
			if part != "" {
				reason = "empty key token after trimming"
			}
			return normalizedChord{}, fmt.Errorf("invalid hotkey %q: %s", raw, reason)
		}
		if len(strings.Fields(trimmed)) > 1 {
			return normalizedChord{}, fmt.Errorf(`invalid hotkey %q: only a single chord is supported`, raw)
		}
		trimmedParts = append(trimmedParts, trimmed)
	}

	keyToken := trimmedParts[len(trimmedParts)-1]
	normalizedKey, err := normalizeKeyboardKey(keyToken)
	if err != nil {
		return normalizedChord{}, fmt.Errorf("invalid hotkey %q: %w", raw, err)
	}

	normalizedModifiers := make([]string, 0, len(trimmedParts)-1)
	seenModifierRaw := make(map[string]string, len(trimmedParts))
	for _, part := range trimmedParts[:len(trimmedParts)-1] {
		modifier, err := normalizeModifierToken(part)
		if err == nil {
			if err := appendNormalizedModifier(&normalizedModifiers, seenModifierRaw, modifier, part, hotkeyContext(raw)); err != nil {
				return normalizedChord{}, fmt.Errorf("invalid hotkey %q: %w", originalRaw, err)
			}
			continue
		}

		asKey, keyErr := normalizeKeyboardKey(part)
		if keyErr == nil {
			if asKey.modifierFamily != "" {
				return normalizedChord{}, fmt.Errorf(
					`invalid hotkey %q: unsupported modifier %q; supported modifiers: %s; left/right-specific modifier keys are only supported as the main key`,
					raw,
					part,
					strings.Join(supportedModifierNames, ", "),
				)
			}
			return normalizedChord{}, fmt.Errorf(`invalid hotkey %q: chords may contain at most one non-modifier key`, raw)
		}

		return normalizedChord{}, fmt.Errorf("invalid hotkey %q: %w", raw, unsupportedModifierError(part))
	}

	if err := ensureNoModifierConflicts(normalizedKey, normalizedModifiers, hotkeyContext(raw)); err != nil {
		return normalizedChord{}, fmt.Errorf("invalid hotkey %q: %w", originalRaw, err)
	}

	chord := normalizedChord{key: normalizedKey.token}
	if len(normalizedModifiers) > 0 {
		chord.modifiers = normalizedModifiers
	}
	return chord, nil
}

func normalizeKeyboardModifiers(raw []string, context string) ([]string, error) {
	modifiers := make([]string, 0, len(raw))
	seenRaw := make(map[string]string, len(raw))
	for _, modifier := range raw {
		normalized, err := normalizeModifierToken(modifier)
		if err != nil {
			return nil, err
		}
		if err := appendNormalizedModifier(&modifiers, seenRaw, normalized, modifier, context); err != nil {
			return nil, err
		}
	}
	return modifiers, nil
}

func appendNormalizedModifier(modifiers *[]string, seenRaw map[string]string, normalized string, raw string, context string) error {
	if previous, exists := seenRaw[normalized]; exists {
		if normalizeNamedToken(previous) == normalizeNamedToken(raw) {
			return fmt.Errorf(`duplicate modifier %q in %s`, normalized, context)
		}
		return fmt.Errorf(`duplicate modifier usage after normalization in %s`, context)
	}

	seenRaw[normalized] = raw
	*modifiers = append(*modifiers, normalized)
	return nil
}

func ensureNoModifierConflicts(key normalizedKey, modifiers []string, context string) error {
	if key.modifierFamily == "" {
		return nil
	}
	if slices.Contains(modifiers, key.modifierFamily) {
		return fmt.Errorf(`duplicate modifier usage after normalization in %s`, context)
	}
	return nil
}

func normalizeKeyboardKey(raw string) (normalizedKey, error) {
	if raw == "" {
		return normalizedKey{}, unsupportedKeyEmptyError(raw)
	}

	if canonical, blocked := disallowedNumpadShorthand[strings.ToLower(raw)]; blocked {
		return normalizedKey{}, fmt.Errorf(`unsupported key %q; use %q`, raw, canonical)
	}

	if utf8.RuneCountInString(raw) == 1 {
		r, _ := utf8.DecodeRuneInString(raw)
		return normalizeSingleKeyRune(raw, r)
	}

	if strings.IndexFunc(raw, unicode.IsControl) >= 0 {
		return normalizedKey{}, unsupportedControlKeyError(raw)
	}
	if strings.IndexFunc(raw, unicode.IsSpace) >= 0 {
		return normalizedKey{}, unsupportedWhitespaceKeyError(raw)
	}
	if !isASCIIString(raw) {
		return normalizedKey{}, unsupportedUnicodeKeyError(raw)
	}

	normalizedName := normalizeNamedToken(raw)
	if alias, ok := mainKeyAliasMap[normalizedName]; ok {
		normalizedName = alias
	}

	key, ok := keyExecutionMap[normalizedName]
	if !ok {
		return normalizedKey{}, fmt.Errorf(
			`unsupported key %q; see supported keyboard key names in the computer-use docs`,
			raw,
		)
	}

	return key, nil
}

func normalizeSingleKeyRune(raw string, r rune) (normalizedKey, error) {
	switch {
	case unicode.IsSpace(r):
		return normalizedKey{}, unsupportedWhitespaceKeyError(raw)
	case unicode.IsControl(r):
		return normalizedKey{}, unsupportedControlKeyError(raw)
	case r > unicode.MaxASCII:
		return normalizedKey{}, unsupportedUnicodeKeyError(raw)
	case unicode.IsLetter(r):
		return normalizedKey{token: strings.ToLower(raw)}, nil
	case unicode.IsDigit(r):
		return normalizedKey{token: raw}, nil
	}

	if _, ok := allowedPunctuationKeys[r]; ok {
		return normalizedKey{token: raw}, nil
	}
	if shiftedBase, ok := shiftedSymbolMap[r]; ok {
		return normalizedKey{}, fmt.Errorf(
			`unsupported key %q; use press(%q, ["shift"]) for a shifted key or type(%q) for text`,
			raw,
			shiftedBase,
			raw,
		)
	}

	return normalizedKey{}, fmt.Errorf(`unsupported key %q; see supported keyboard key names in the computer-use docs`, raw)
}

func normalizeModifierToken(raw string) (string, error) {
	if raw == "" {
		return "", fmt.Errorf(`unsupported modifier %q; supported modifiers: %s`, raw, strings.Join(supportedModifierNames, ", "))
	}
	if strings.IndexFunc(raw, unicode.IsControl) >= 0 || strings.IndexFunc(raw, unicode.IsSpace) >= 0 || !isASCIIString(raw) {
		return "", unsupportedModifierError(raw)
	}

	normalizedName := normalizeNamedToken(raw)
	if alias, ok := modifierAliasMap[normalizedName]; ok {
		return alias, nil
	}
	if _, ok := keyExecutionMap[normalizedName]; ok {
		return "", fmt.Errorf(
			`unsupported modifier %q; supported modifiers: %s; left/right-specific modifier keys are only supported as the main key`,
			raw,
			strings.Join(supportedModifierNames, ", "),
		)
	}
	return "", unsupportedModifierError(raw)
}

func unsupportedModifierError(raw string) error {
	return fmt.Errorf(`unsupported modifier %q; supported modifiers: %s`, raw, strings.Join(supportedModifierNames, ", "))
}

func unsupportedKeyEmptyError(raw string) error {
	return fmt.Errorf(`unsupported key %q: empty key token`, raw)
}

func unsupportedWhitespaceKeyError(raw string) error {
	switch raw {
	case " ":
		return fmt.Errorf(`unsupported key %q; use press("space") or type(" ")`, raw)
	case "\t":
		return fmt.Errorf(`unsupported key %q; use press("tab") or type("\t")`, raw)
	case "\n", "\r":
		return fmt.Errorf(`unsupported key %q; use press("enter") or type(%q)`, raw, raw)
	default:
		return fmt.Errorf(`unsupported key %q; use a named key such as "space", "tab", or "enter"`, raw)
	}
}

func unsupportedControlKeyError(raw string) error {
	if raw == "\t" || raw == "\n" || raw == "\r" {
		return unsupportedWhitespaceKeyError(raw)
	}
	return fmt.Errorf(`unsupported key %q; use type(%q) for text input`, raw, raw)
}

func unsupportedUnicodeKeyError(raw string) error {
	return fmt.Errorf(`unsupported key %q; use type(%q) for text input`, raw, raw)
}

func normalizeNamedToken(raw string) string {
	return strings.ReplaceAll(strings.ToLower(raw), "-", "_")
}

func isASCIIString(value string) bool {
	for _, r := range value {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

func hotkeyContext(raw string) string {
	return fmt.Sprintf("hotkey(%q)", raw)
}

func pressContext(key string, modifiers []string) string {
	var builder strings.Builder
	builder.WriteString("press(")
	builder.WriteString(fmt.Sprintf("%q", key))
	builder.WriteString(", [")
	for i, modifier := range modifiers {
		if i > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%q", modifier))
	}
	builder.WriteString("])")
	return builder.String()
}
