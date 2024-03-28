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
	view_util "github.com/daytonaio/daytona/pkg/views/util"
)

type GitProviderView struct {
	Id         string
	Name       string
	Username   string
	BaseApiUrl string
	Token      string
}

func GitProviderSelectionView(gitProviderAddView *serverapiclient.GitProvider, userGitProviders []serverapiclient.GitProvider, isDeleting bool) {
	supportedProviders := config.GetSupportedGitProviders()

	var gitProviderOptions []huh.Option[string]
	for _, supportedProvider := range supportedProviders {
		if isDeleting {
			for _, userProvider := range userGitProviders {
				if *userProvider.Id == supportedProvider.Id {
					gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
				}
			}
		} else {
			gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
		}
	}

	gitProviderForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					gitProviderOptions...,
				).
				Value(gitProviderAddView.Id)),
	).WithTheme(views.GetCustomTheme())

	err := gitProviderForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	userDataForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Username").
				Value(gitProviderAddView.Username).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("username can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return isDeleting || !providerRequiresUsername(*gitProviderAddView.Id)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Self-managed API URL").
				Value(gitProviderAddView.BaseApiUrl).
				Description("For example: http://gitlab-host/api/v4/ or https://gitea-host").
				Validate(func(str string) error {
					if str == "" {
						return errors.New("URL can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return isDeleting || !providerRequiresApiUrl(*gitProviderAddView.Id)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Personal access token").
				Value(gitProviderAddView.Token).
				Password(true).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("token can not be blank")
					}
					return nil
				}),
		).WithHide(isDeleting),
	).WithTheme(views.GetCustomTheme())

	view_util.RenderInfoMessage(fmt.Sprintf("More information on:\n%s", config.GetDocsLinkFromGitProvider(*gitProviderAddView.Id)))

	err = userDataForm.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func providerRequiresUsername(gitProviderId string) bool {
	return gitProviderId == "bitbucket"
}

func providerRequiresApiUrl(gitProviderId string) bool {
	return gitProviderId == "gitlab-self-managed" || gitProviderId == "gitea"
}
