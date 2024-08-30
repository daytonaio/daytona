// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

func GitProviderSelectionViewForAutoSelection(identity *string, userGitProviders []apiclient.GitProvider) {
	var gitProviderOptions []huh.Option[string]
	for _, gp := range userGitProviders {
		gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: gp.TokenIdentity, Value: gp.TokenIdentity})
	}

	gitProviderForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a token").
				Options(
					gitProviderOptions...,
				).
				Value(identity)),
	).WithTheme(views.GetCustomTheme())

	err := gitProviderForm.Run()
	if err != nil {
		log.Fatal(err)
	}

}
