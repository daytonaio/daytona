// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

var commonGitProviderIds = []string{"github", "gitlab", "bitbucket"}

func GitProviderCreationView(ctx context.Context, apiClient *apiclient.APIClient, gitProviderAddView *apiclient.SetGitProviderConfig, existingAliases []string) error {
	supportedProviders := config.GetSupportedGitProviders()

	var gitProviderOptions []huh.Option[string]
	var otherGitProviderOptions []huh.Option[string]
	for _, supportedProvider := range supportedProviders {
		if slices.Contains(commonGitProviderIds, supportedProvider.Id) {
			gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
		} else {
			otherGitProviderOptions = append(otherGitProviderOptions, huh.Option[string]{Key: supportedProvider.Name, Value: supportedProvider.Id})
		}
	}

	if len(otherGitProviderOptions) > 0 {
		gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: "Other", Value: "other"})
	}

	initialAlias := gitProviderAddView.Alias

	if gitProviderAddView.ProviderId == "" {
		gitProviderForm := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose a Git provider").
					Description("Note: for updating an existing Git provider use 'daytona git-provider update'").
					Options(
						gitProviderOptions...,
					).
					Value(&gitProviderAddView.ProviderId)).WithHeight(8),
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Choose a Git provider").
					Options(
						otherGitProviderOptions...,
					).
					Value(&gitProviderAddView.ProviderId)).WithHeight(12).WithHideFunc(func() bool {
				return gitProviderAddView.ProviderId != "other"
			}),
		).WithTheme(views.GetCustomTheme())

		err := gitProviderForm.Run()
		if err != nil {
			return err
		}
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
		).WithHeight(5).WithHideFunc(func() bool {
			return !providerRequiresUsername(gitProviderAddView.ProviderId)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Self-managed API URL").
				Value(gitProviderAddView.BaseApiUrl).
				Description(getApiUrlDescription(gitProviderAddView.ProviderId)).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("URL can not be blank")
					}
					return nil
				}),
		).WithHeight(6).WithHideFunc(func() bool {
			return !providerRequiresApiUrl(gitProviderAddView.ProviderId)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Personal access token").
				Value(&gitProviderAddView.Token).
				EchoMode(huh.EchoModePassword).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("token can not be blank")
					}
					return nil
				}),
		).WithHeight(5),
		huh.NewGroup(
			huh.NewInput().
				Title("Alias").
				Description("Will default to username if left empty").
				Value(gitProviderAddView.Alias).
				Validate(func(str string) error {
					for _, alias := range existingAliases {
						if alias == str {
							if initialAlias == nil || *initialAlias != str {
								return errors.New("alias is already in use")
							}
						}
					}
					return nil
				}),
		).WithHeight(6),
	).WithTheme(views.GetCustomTheme())

	views.RenderInfoMessage(getGitProviderHelpMessage(gitProviderAddView.ProviderId))
	err := userDataForm.Run()
	if err != nil {
		return err
	}

	return nil
}

func providerRequiresUsername(gitProviderId string) bool {
	return gitProviderId == "bitbucket" || gitProviderId == "bitbucket-server" || gitProviderId == "aws-codecommit"
}

func providerRequiresApiUrl(gitProviderId string) bool {
	providersRequiringApiUrl := []string{
		"gitness",
		"github-enterprise-server",
		"gitlab-self-managed",
		"gitea",
		"bitbucket-server",
		"azure-devops",
		"aws-codecommit",
		"gogs",
	}
	return slices.Contains(providersRequiringApiUrl, gitProviderId)
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
	} else if gitProviderId == "aws-codecommit" {
		return "For example: https://ap-south-1.console.aws.amazon.com"
	} else if gitProviderId == "gogs" {
		return "For example: https://gogs-host.com"
	}
	return ""
}

func getGitProviderHelpMessage(gitProviderId string) string {
	message := fmt.Sprintf("%s\n%s",
		lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("More information on:"),
		config.GetDocsLinkFromGitProvider(gitProviderId))

	requiredScopes := config.GetRequiredScopesFromGitProviderId(gitProviderId)
	if requiredScopes != "" {
		message = fmt.Sprintf("%s\n\n%s%s",
			message,
			lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("Required scopes: "),
			requiredScopes)
	}
	prebuildScopes := config.GetPrebuildScopesFromGitProviderId(gitProviderId)
	if prebuildScopes != "" {
		message = fmt.Sprintf("%s\n%s%s",
			message,
			lipgloss.NewStyle().Foreground(views.Green).Bold(true).Render("Prebuild scopes: "),
			prebuildScopes)
	}

	return message
}
