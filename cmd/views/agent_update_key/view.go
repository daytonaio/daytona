// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package agent_update_key

import (
	"log"

	"github.com/daytonaio/daytona/cmd/views"
	views_util "github.com/daytonaio/daytona/cmd/views/util"

	"github.com/charmbracelet/huh"
)

type AgentUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func InteractiveForm(agentUpdateKeyView *AgentUpdateKeyView) {
	authType := views_util.SshAuthTypePrivateKey

	privateKeyGroups := views_util.PrivateKeyForm(&authType, nil, &agentUpdateKeyView.PathToPrivateKey)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Generate or copy").
				Description("Generate a new or copy an existing key to the agent").
				Options(
					huh.NewOption("Generate", true),
					huh.NewOption("Copy", false),
				).
				Value(&agentUpdateKeyView.GenerateNewKey),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !agentUpdateKeyView.GenerateNewKey {
		form = huh.NewForm(
			privateKeyGroups...,
		).WithTheme(views.GetCustomTheme())

		err = form.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
