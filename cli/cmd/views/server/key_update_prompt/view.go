// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package key_update_prompt

import (
	"log"

	"github.com/daytonaio/daytona/cli/cmd/views"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"

	"github.com/charmbracelet/huh"
)

type ServerUpdateKeyView struct {
	GenerateNewKey   bool
	PathToPrivateKey string
}

func InteractiveForm(serverUpdateKeyView *ServerUpdateKeyView) {
	authType := views_util.SshAuthTypePrivateKey

	privateKeyGroups := views_util.PrivateKeyForm(&authType, nil, &serverUpdateKeyView.PathToPrivateKey)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[bool]().
				Title("Generate or copy").
				Description("Generate a new or copy an existing key to the server").
				Options(
					huh.NewOption("Generate", true),
					huh.NewOption("Copy", false),
				).
				Value(&serverUpdateKeyView.GenerateNewKey),
		),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !serverUpdateKeyView.GenerateNewKey {
		form = huh.NewForm(
			privateKeyGroups...,
		).WithTheme(views.GetCustomTheme())

		err = form.Run()
		if err != nil {
			log.Fatal(err)
		}
	}
}
