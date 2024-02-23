// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_git_provider

import (
	"errors"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/common/api_client"
)

type GitProviderSelectView struct {
	Id       string
	Username string
	Token    string
}

func GitProviderSelectionView(gitProviderAddView *GitProviderSelectView, userGitProviders []api_client.GitProvider, isDeleting bool) {
	availableGitProviders := config.GetGitProviderList()

	var options []huh.Option[string]
	for _, availableProvider := range availableGitProviders {
		if isDeleting {
			for _, userProvider := range userGitProviders {
				if *userProvider.Id == availableProvider.Id {
					options = append(options, huh.Option[string]{Key: availableProvider.Name, Value: availableProvider.Id})
				}
			}
		} else {
			options = append(options, huh.Option[string]{Key: availableProvider.Name, Value: availableProvider.Id})
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					options...,
				).
				Value(&gitProviderAddView.Id)),
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(&gitProviderAddView.Username).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("username can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return isDeleting || gitProviderAddView.Id != "bitbucket"
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Personal access token").
				Value(&gitProviderAddView.Token).
				Password(true).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("token can not be blank")
					}
					return nil
				}),
		).WithHide(isDeleting),
	).WithTheme(views.GetCustomTheme())

	err := form.Run()
	if err != nil {
		log.Fatal(err)
	}
}
