package views

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestGetInitialCommandTheme(t *testing.T) {
	theme := GetInitialCommandTheme()

	//check blurred styles
	assert.True(t, theme.Blurred.FocusedButton.GetBold())
	assert.Equal(t, Green, theme.Blurred.TextInput.Prompt.GetForeground())
	assert.Equal(t, Green, theme.Blurred.TextInput.Cursor.GetForeground())
	assert.Equal(t, Green, theme.Blurred.SelectSelector.GetForeground())

	//check focused styles
	assert.Equal(t, lipgloss.Color("fff"), theme.Focused.Base.GetBorderBottomForeground())
	assert.Equal(t, Green, theme.Focused.Title.GetForeground())
	assert.True(t, theme.Focused.Title.GetBold())
	assert.True(t, theme.Focused.FocusedButton.GetBold())
	assert.Equal(t, Green, theme.Focused.FocusedButton.GetBackground())
	assert.Equal(t, Green, theme.Focused.TextInput.Prompt.GetForeground())
	assert.Equal(t, Light, theme.Focused.TextInput.Cursor.GetForeground())
	assert.Equal(t, Green, theme.Focused.SelectSelector.GetForeground())
	assert.False(t, theme.Focused.Base.GetBorderLeft())
	assert.Equal(t, Green, theme.Focused.SelectedOption.GetForeground())
}

func TestGetCustomTheeme(t *testing.T) {
	theme := GetCustomTheme()

	//check blurred styles
	assert.Equal(t, Green, theme.Blurred.FocusedButton.GetBackground())
	assert.True(t, theme.Blurred.FocusedButton.GetBold())
	assert.Equal(t, Light, theme.Blurred.TextInput.Prompt.GetForeground())
	assert.Equal(t, Light, theme.Blurred.TextInput.Cursor.GetForeground())
	assert.Equal(t, Green, theme.Blurred.SelectSelector.GetForeground())
	assert.Equal(t, Gray, theme.Blurred.Title.GetForeground())
	assert.True(t, theme.Blurred.Title.GetBold())
	assert.Equal(t, LightGray, theme.Blurred.Description.GetForeground())

	//check focused styles
	assert.Equal(t, Green, theme.Focused.Title.GetForeground())
	assert.True(t, theme.Focused.Title.GetBold())
	assert.Equal(t, LightGray, theme.Focused.Description.GetForeground())
	assert.True(t, theme.Focused.Description.GetBold())
	assert.True(t, theme.Focused.FocusedButton.GetBold())
	assert.Equal(t, Green, theme.Focused.FocusedButton.GetBackground())
	assert.Equal(t, Green, theme.Focused.TextInput.Prompt.GetForeground())
	assert.Equal(t, Light, theme.Focused.TextInput.Cursor.GetForeground())
	assert.Equal(t, Green, theme.Focused.SelectSelector.GetForeground())
	assert.Equal(t, Green, theme.Focused.SelectedOption.GetForeground())
	assert.Equal(t, Red, theme.Focused.ErrorIndicator.GetForeground())
	assert.Equal(t, Red, theme.Focused.ErrorMessage.GetForeground())
	assert.Equal(t, Green, theme.Focused.Base.GetBorderBottomForeground())
	assert.Equal(t, DefaultLayoutMarginTop, theme.Focused.Base.GetMarginTop())
	assert.Equal(t, DefaultLayoutMarginTop, theme.Blurred.Base.GetMarginTop())
}
