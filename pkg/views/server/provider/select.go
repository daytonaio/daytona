// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package provider

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views/util"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

func GetProviderFromPrompt(providers []serverapiclient.Provider, title string) *serverapiclient.Provider {
	util.RenderMainTitle(title)

	if len(providers) == 0 {
		view_util.RenderInfoMessage("No providers found")
		return nil
	}

	modelInstance := renderProvidersList(providers, true)

	m, err := tea.NewProgram(modelInstance).Run()
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	selectedProvider := m.(model).selectedProvider

	lipgloss.DefaultRenderer().Output().ClearLines(strings.Count(modelInstance.View(), "\n") + 2)

	return selectedProvider
}
