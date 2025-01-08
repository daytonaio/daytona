// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0
package views

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"golang.org/x/term"
)

var getTerminalSize = term.GetSize

func TestRenderMainTitle(t *testing.T) {
	title := "Render Main Title"
	expectedOutput := lipgloss.NewStyle().Foreground(Green).Bold(true).Padding(1, 0, 1, 0).Render(title) + "\n"
	actualOutput := captureOutput(func() { RenderMainTitle(title) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderTip(t *testing.T) {
	message := "Render Tip"
	expectedOutput := lipgloss.NewStyle().Padding(0, 0, 1, 1).Render(message) + "\n"

	actualOutput := captureOutput(func() { RenderTip(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderInfoMessage(t *testing.T) {
	message := "Render Info message"
	expectedOutput := lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message) + "\n"
	actualOutput := captureOutput(func() { RenderInfoMessage(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderViewBuildLogsMessage(t *testing.T) {
	buildId := "Render View Build Logs"
	expectedOutput := lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(
		fmt.Sprintf("The build has been scheduled for running. Use `daytona build logs %s -f` to view the progress.", buildId),
	) + "\n"
	actualOutput := captureOutput(func() { RenderViewBuildLogsMessage(buildId) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderCreationInfoMessage(t *testing.T) {
	message := "Creation Info Message"
	expectedOutput := lipgloss.NewStyle().Foreground(Gray).Padding(1, 0, 1, 1).Render(message) + "\n"
	actualOutput := captureOutput(func() { RenderCreationInfoMessage(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderListLine(t *testing.T) {
	message := "Render List Line"
	expectedOutput := lipgloss.NewStyle().Padding(0, 0, 1, 1).Render(message) + "\n"
	actualOutput := captureOutput(func() { RenderListLine(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderInfoMessageBold(t *testing.T) {
	message := "Render Info Message Bold"
	expectedOutput := lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message) + "\n"
	actualOutput := captureOutput(func() { RenderInfoMessageBold(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderBorderedMessage(t *testing.T) {
	message := "Render Bordered Message"
	expectedOutput := GetBorderedMessage(message) + "\n"
	actualOutput := captureOutput(func() { RenderBorderedMessage(message) })
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestGetListFooter(t *testing.T) {
	profileName := "Get List Footer"
	padding := &Padding{1, 2, 3, 4}
	expectedOutput := lipgloss.NewStyle().Bold(true).Padding(padding.Top, padding.Right, padding.Bottom, padding.Left).Render("\n\nActive profile: " + profileName)
	actualOutput := GetListFooter(profileName, padding)
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestGetStyledMainTitle(t *testing.T) {
	content := "Get Styled Main Title"
	expectedOutput := lipgloss.NewStyle().Foreground(Dark).Background(Light).Padding(0, 1).Render(content)
	actualOutput := GetStyledMainTitle(content)
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestGetInfoMessage(t *testing.T) {
	message := "Get Info Message"
	expectedOutput := lipgloss.NewStyle().Padding(1, 0, 1, 1).Render(message)
	actualOutput := GetInfoMessage(message)
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestGetBoldedInfoMessage(t *testing.T) {
	message := "Get Bolded Info Message"
	expectedOutput := lipgloss.NewStyle().Bold(true).Padding(1, 0, 1, 1).Render(message)
	actualOutput := GetBoldedInfoMessage(message)
	assert.Equal(t, expectedOutput, actualOutput)
}

func TestRenderContainerLayout(t *testing.T) {
	originalGetSize := getTerminalSize
	defer func() { getTerminalSize = originalGetSize }()
	t.Run("terminal size error", func(t *testing.T) {
		getTerminalSize = mockGetSize(0, 0, fmt.Errorf("mock error"))
		expectedOutput := DocStyle.Render("Error: Unable to get terminal size") + "\n"

		actualOutput := captureOutput(func() { RenderContainerLayout("Test Output") })
		assert.Equal(t, expectedOutput, actualOutput)
	})

	t.Run("successful render", func(t *testing.T) {
		getTerminalSize = mockGetSize(100, 10, nil)
		expectedOutput := "Test Output"
		actualOutput := captureOutput(func() {
			RenderContainerLayout(expectedOutput)
		})

		assert.Equal(t, expectedOutput, actualOutput)
	})
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	os.Stdout = w
	f()

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stdout = stdout

	return buf.String()
}

func mockGetSize(width, height int, err error) func(fd int) (int, int, error) {
	return func(fd int) (int, int, error) {
		return width, height, err
	}
}
