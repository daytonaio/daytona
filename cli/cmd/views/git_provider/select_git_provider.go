// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package views_git_provider

import (
	"errors"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cli/cmd/views"
	"github.com/daytonaio/daytona/cli/config"
)

type GitProviderSelectView struct {
	Id       string
	Username string
	Token    string
}

func GitProviderSelectionView(gitProviderAddView *GitProviderSelectView, isDeleting bool) {
	providers := config.GetGitProviderList()

	var options []huh.Option[string]
	for _, provider := range providers {
		options = append(options, huh.Option[string]{Key: provider.Name, Value: provider.Id})
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
