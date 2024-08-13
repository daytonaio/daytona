// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"log"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type GitProviderView struct {
	Id         string
	Name       string
	Username   string
	BaseApiUrl string
	Token      string
}

var commonGitProviderIds = []string{"github", "gitlab", "bitbucket"}

func GitProviderSelectionView(gitProviderAddView *apiclient.GitProvider, userGitProviders []apiclient.GitProvider, isDeleting bool) {
	supportedProviders := config.GetSupportedGitProviders()

	var gitProviderOptions []huh.Option[string]
	var otherGitProviderOptions []huh.Option[string]
	for _, supportedProvider := range supportedProviders {
		if isDeleting {
			for _, userProvider := range userGitProviders {
				if *userProvider.Id == supportedProvider.Id {
					gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
				}
			}
		} else {
			if slices.Contains(commonGitProviderIds, supportedProvider.Id) {
				gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
			} else {
				otherGitProviderOptions = append(otherGitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
			}
		}
	}

	if len(otherGitProviderOptions) > 0 {
		gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: "Other", Value: "other"})
	}

	gitProviderForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					gitProviderOptions...,
				).
				Value(gitProviderAddView.Id)).WithHeight(8),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					otherGitProviderOptions...,
				).
				Value(gitProviderAddView.Id)).WithHeight(11).WithHideFunc(func() bool {
			return *gitProviderAddView.Id != "other"
		}),
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
				Description(getApiUrlDescription(*gitProviderAddView.Id)).
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

	if !isDeleting {
		views.RenderInfoMessage(getGitProviderHelpMessage(*gitProviderAddView.Id))
	}

	err = userDataForm.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func providerRequiresUsername(gitProviderId string) bool {
	return gitProviderId == "bitbucket" || gitProviderId == "bitbucket-server"
}

func providerRequiresApiUrl(gitProviderId string) bool {
	return gitProviderId == "gitness" || gitProviderId == "github-enterprise-server" || gitProviderId == "gitlab-self-managed" || gitProviderId == "gitea" || gitProviderId == "bitbucket-server" || gitProviderId == "azure-devops"
}

func getApiUrlDescription(gitProviderId string) string {
	if gitProviderId == "gitlab-self-managed" {
		return "For example: http://gitlab-host/api/v4/"
	} else if gitProviderId == "github-enterprise-server" {
		return "For example: https://github-host"
	} else if gitProviderId == "gitea" {
		return "For example: http://gitea-host"
	} else if gitProviderId == "gitness" {
		return "For example: http://gitness-host/api/v1/"
	} else if gitProviderId == "azure-devops" {
		return "For example: https://dev.azure.com/organization"
	} else if gitProviderId == "bitbucket-server" {
		return "For example: https://bitbucket.host.com/rest"
	}
	return ""
}

func getGitProviderHelpMessage(gitProviderId string) string {
	return fmt.Sprintf("%s\n%s\n\n%s%s",
		lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("More information on:"),
		config.GetDocsLinkFromGitProvider(gitProviderId),
		lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("Required scopes: "),
		config.GetScopesFromGitProvider(gitProviderId))
}
