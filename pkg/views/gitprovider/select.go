// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"log"

	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/serverapiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type GitProviderSelectView struct {
	Id       string
	Username string
	Token    string
}

func GitProviderSelectionView(gitProviderAddView *GitProviderSelectView, userGitProviders []serverapiclient.GitProvider, isDeleting bool) {
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

	gitProviderForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					options...,
				).
				Value(&gitProviderAddView.Id)),
	).WithTheme(views.GetCustomTheme())

	err := gitProviderForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	userDataForm := huh.NewForm(
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

	fmt.Println("More information on:")
	fmt.Println(config.GetDocsLinkFromGitProvider(gitProviderAddView.Id))
	fmt.Println()

	err = userDataForm.Run()
	if err != nil {
		log.Fatal(err)
	}
}
