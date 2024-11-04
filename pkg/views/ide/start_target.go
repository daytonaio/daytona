// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package ide

import (
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/views"
)

func RunStartWorkspaceForm(workspaceName string) bool {
	confirmCheck := true

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(fmt.Sprintf("The workspace %s is stopped, would you like to start it?", workspaceName)).
				Value(&confirmCheck),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !confirmCheck {
		fmt.Println("Operation canceled.")
	}

	return confirmCheck
}
