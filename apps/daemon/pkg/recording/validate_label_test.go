// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package recording

import (
	"strings"
	"testing"
)

func TestValidateLabel(t *testing.T) {
	valid := []string{
		"recording",
		"my recording 1",
		"a.b_c-d",
		"v1.2.3",
		"x",
		"trailing space ",
		" leading space",
		strings.Repeat("a", 100),
	}
	for _, label := range valid {
		if err := validateLabel(label); err != nil {
			t.Errorf("validateLabel(%q) = %v, want nil", label, err)
		}
	}

	invalid := []string{
		"",
		"   ",
		"\t",
		strings.Repeat("a", 101),
		"a/b",
		`a\b`,
		".hidden",
		" .hidden",
		"..",
		"a..b",
		"v1..final",
		"a.. b",
		"a\tb",
		"a\nb",
		"a\rb",
		"a\vb",
		"a\fb",
		"name%03d", // ffmpeg output pattern injection
		"a:b",
		"a*b",
		"a?b",
		"a\"b",
		"a<b>c",
		"a|b",
		"h\u00e9llo", // non-ASCII
		"a\x00b",
	}
	for _, label := range invalid {
		if err := validateLabel(label); err == nil {
			t.Errorf("validateLabel(%q) = nil, want error", label)
		}
	}
}
