// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

type GitProviderView struct {
	Id                 string
	Name               string
	Username           string
	BaseApiUrl         string
	Token              string
	TokenScopeIdentity string
	TokenScope         string
	TokenScopeType     string
}

// Identity*
// Scope+
// 		org
// 		org/repo
// 		global

var commonGitProviderIds = []string{"github", "gitlab", "bitbucket"}

func GitProviderSelectionView(gitProviderAddView *apiclient.SetGitProviderConfig, userGitProviders []apiclient.GitProvider, isDeleting bool) {
	supportedProviders := config.GetSupportedGitProviders()
	var gitProviderOptions []huh.Option[string]
	var otherGitProviderOptions []huh.Option[string]
	for _, supportedProvider := range supportedProviders {
		if isDeleting {
			for _, userProvider := range userGitProviders {
				if strings.Split(userProvider.Id, "_")[0] == supportedProvider.Id {
					gitProviderOptions = append(gitProviderOptions, huh.Option[string]{Key: fmt.Sprintf("%s    %s", supportedProvider.Name, userProvider.TokenIdentity), Value: userProvider.Id})
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
				Value(&gitProviderAddView.Id)),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose a Git provider").
				Options(
					otherGitProviderOptions...,
				).
				Value(&gitProviderAddView.Id)).WithHideFunc(func() bool {
			return gitProviderAddView.Id != "other"
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
			return isDeleting || !providerRequiresUsername(gitProviderAddView.Id)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Self-managed API URL").
				Value(gitProviderAddView.BaseApiUrl).
				Description(getApiUrlDescription(gitProviderAddView.Id)).
				Validate(func(str string) error {
					if str == "" {
						return errors.New("URL can not be blank")
					}
					return nil
				}),
		).WithHideFunc(func() bool {
			return isDeleting || !providerRequiresApiUrl(gitProviderAddView.Id)
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
		huh.NewGroup(
			huh.NewInput().
				Title("Token Identity ").
				Value(gitProviderAddView.TokenIdentity).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("token identity can not be blank")
					}
					return nil
				}),
		).WithHide(isDeleting),
		huh.NewGroup(
			huh.NewSelect[apiclient.GitproviderTokenScopeType]().
				Title("Select token scope type").
				Value(gitProviderAddView.TokenScopeType).
				Options(
					huh.Option[apiclient.GitproviderTokenScopeType]{Key: "Global", Value: apiclient.TokenScopeTypeGlobal},
					huh.Option[apiclient.GitproviderTokenScopeType]{Key: "Organization", Value: apiclient.TokenScopeTypeOrganization},
					huh.Option[apiclient.GitproviderTokenScopeType]{Key: "Repository", Value: apiclient.TokenScopeTypeRepository},
				).
				Validate(func(gtst apiclient.GitproviderTokenScopeType) error {
					if gtst == "" {
						return errors.New("scope must be selected")
					}
					return nil
				}),
		).WithHide(isDeleting),
	).WithTheme(views.GetCustomTheme())

	err = userDataForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	tokenDataForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Enter the organization name associated with this token").
				Value(gitProviderAddView.TokenScope).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("enter the name of the org to which token has access")
					}
					return nil
				}),
		).WithHide(isDeleting || (*gitProviderAddView.TokenScopeType != apiclient.TokenScopeTypeOrganization)),
		huh.NewGroup(
			huh.NewInput().
				Title("Enter the repository path this token can access e.g., daytonaio/daytona").
				Value(gitProviderAddView.TokenScope).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("enter the path of the repe to which token has access Ex: daytonaio/daytona")
					}
					return nil
				}),
		).WithHide(isDeleting || (*gitProviderAddView.TokenScopeType != apiclient.TokenScopeTypeRepository)),
	).WithTheme(views.GetCustomTheme())

	if !isDeleting {
		views.RenderInfoMessage(getGitProviderHelpMessage(gitProviderAddView.Id))
	}

	err = tokenDataForm.Run()
	if err != nil {
		log.Fatal(err)
	}

	if *gitProviderAddView.TokenScopeType == apiclient.TokenScopeTypeGlobal {
		*gitProviderAddView.TokenScope = "global"
	}

}

func providerRequiresUsername(gitProviderId string) bool {
	return gitProviderId == "bitbucket" || gitProviderId == "bitbucket-server" || gitProviderId == "aws-codecommit"
}

func providerRequiresApiUrl(gitProviderId string) bool {
	return gitProviderId == "gitness" || gitProviderId == "github-enterprise-server" || gitProviderId == "gitlab-self-managed" || gitProviderId == "gitea" || gitProviderId == "bitbucket-server" || gitProviderId == "azure-devops" || gitProviderId == "aws-codecommit"
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
