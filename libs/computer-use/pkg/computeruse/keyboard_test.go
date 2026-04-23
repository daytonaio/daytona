// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package computeruse

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTypingActions(t *testing.T) {
	t.Run("groups printable text and normalizes CRLF", func(t *testing.T) {
		actions, err := buildTypingActions("line one\nline two\r\nline three")
		require.NoError(t, err)
		assert.Equal(t, []typingAction{
			{kind: typingActionText, text: "line one"},
			{kind: typingActionEnter},
			{kind: typingActionText, text: "line two"},
			{kind: typingActionEnter},
			{kind: typingActionText, text: "line three"},
		}, actions)
	})

	t.Run("rejects unsupported control characters", func(t *testing.T) {
		_, err := buildTypingActions("hello\vworld")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported control character")
		assert.Contains(t, err.Error(), "U+000B")
	})

	t.Run("rejects unicode line separators", func(t *testing.T) {
		_, err := buildTypingActions("hello\u2028world")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported separator character")
		assert.Contains(t, err.Error(), "U+2028")
	})

	t.Run("rejects tab characters", func(t *testing.T) {
		_, err := buildTypingActions("hello\tworld")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not translate '\\t' to Tab")
		assert.Contains(t, err.Error(), `keyboard.press("tab")`)
	})

	t.Run("returns no actions for unsupported input", func(t *testing.T) {
		actions, err := buildTypingActions("hello\fworld")
		require.Error(t, err)
		assert.Nil(t, actions)
	})

	t.Run("returns no actions for empty input", func(t *testing.T) {
		actions, err := buildTypingActions("")
		require.NoError(t, err)
		assert.Empty(t, actions)
	})
}
