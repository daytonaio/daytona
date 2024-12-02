// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
)

var commonGitProviderIds = []string{"github", "gitlab", "bitbucket"}

func GitProviderCreationView(ctx context.Context, apiClient *apiclient.APIClient, gitProviderAddView *apiclient.SetGitProviderConfig, existingAliases []string, flags map[string]string) error {
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

	var selectedSigningMethod string
	var signingKey string

	if gitProviderAddView.SigningMethod != nil {
		selectedSigningMethod = string(*gitProviderAddView.SigningMethod)
		gitProviderAddView.SigningKey = nil
	}

	aliasFlag := flags["alias"]
	tokenFlag := flags["token"]
	baseApiUrlFlag := flags["base-api-url"]
	usernameFlag := flags["username"]
	signingMethodFlag := flags["signing-method"]
	signingKeyFlag := flags["signing-key"]

	if usernameFlag != "" {
		if ProviderRequiresUsername(gitProviderAddView.ProviderId) {
			gitProviderAddView.Username = &usernameFlag
		} else {
			return fmt.Errorf("username is not required for '%s' provider", gitProviderAddView.ProviderId)
		}
	}

	if baseApiUrlFlag != "" {
		if ProviderRequiresApiUrl(gitProviderAddView.ProviderId) {
			gitProviderAddView.BaseApiUrl = &baseApiUrlFlag
		} else {
			return fmt.Errorf("base API URL is not required for '%s' provider", gitProviderAddView.ProviderId)
		}
	}

	if signingMethodFlag != "" || signingKeyFlag != "" {
		err := ValidateSigningMethodAndKey(signingMethodFlag, signingKeyFlag, gitProviderAddView.ProviderId)
		if err != nil {
			return err
		}
		signingMethod := apiclient.SigningMethod(signingMethodFlag)
		gitProviderAddView.SigningMethod = &signingMethod
		gitProviderAddView.SigningKey = &signingKeyFlag
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
			return usernameFlag != "" || !ProviderRequiresUsername(gitProviderAddView.ProviderId)
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
			return baseApiUrlFlag != "" || !ProviderRequiresApiUrl(gitProviderAddView.ProviderId)
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
		).WithHeight(5).WithHideFunc(func() bool {
			return tokenFlag != ""
		}),
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
		).WithHeight(6).WithHideFunc(func() bool {
			return aliasFlag != ""
		}),

		huh.NewGroup(huh.NewSelect[string]().
			Title("Commit Signing Method").
			DescriptionFunc(func() string {
				return getGitProviderSigningHelpMessage(gitProviderAddView.ProviderId)
			}, nil).
			Options(
				huh.Option[string]{Key: "None", Value: "none"},
				huh.Option[string]{Key: "SSH", Value: "ssh"},
				huh.Option[string]{Key: "GPG", Value: "gpg"},
			).
			Value(&selectedSigningMethod).WithHeight(6),
		).WithHeight(8).WithHideFunc(func() bool {
			return signingMethodFlag != "" || CommitSigningNotSupported(gitProviderAddView.ProviderId)
		}),
		huh.NewGroup(
			huh.NewInput().
				Title("Signing Key").
				Value(&signingKey).
				DescriptionFunc(func() string {
					return getSigningKeyDescription(selectedSigningMethod)
				}, nil).
				Validate(func(str string) error {
					if selectedSigningMethod != "none" && str == "" {
						return errors.New("signing key cannot be blank when a signing method is selected")
					}

					if selectedSigningMethod == "ssh" {
						if err := IsValidSSHKey(str); err != nil {
							return err
						}
					}
					return nil
				}),
		).WithHeight(5).WithHideFunc(func() bool {
			return signingKeyFlag != "" || selectedSigningMethod == "none"
		}),
	).WithTheme(views.GetCustomTheme())

	views.RenderInfoMessage(getGitProviderHelpMessage(gitProviderAddView.ProviderId))
	err := userDataForm.Run()
	if err != nil {
		return err
	}

	if selectedSigningMethod != "none" {
		gitProviderAddView.SigningMethod = (*apiclient.SigningMethod)(&selectedSigningMethod)
		gitProviderAddView.SigningKey = &signingKey
	} else {
		gitProviderAddView.SigningKey = nil
		gitProviderAddView.SigningMethod = nil
	}

	return nil

}
func IsValidSSHKey(key string) error {
	sshKeyPattern := regexp.MustCompile(`^(ssh-(rsa|ed25519|dss|ecdsa-sha2-nistp(256|384|521)))\s+[A-Za-z0-9+/=]+(\s+.+)?$`)
	if !sshKeyPattern.MatchString(key) {
		return errors.New("invalid SSH key: must start with valid SSH key type (e.g., ssh-rsa, ssh-ed25519)")
	}

	return nil
}

func ProviderRequiresUsername(gitProviderId string) bool {
	return gitProviderId == "bitbucket" || gitProviderId == "bitbucket-server" || gitProviderId == "aws-codecommit"
}

func ProviderRequiresApiUrl(gitProviderId string) bool {
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

func CommitSigningNotSupported(gitProviderId string) bool {
	return gitProviderId == "gitness" || gitProviderId == "bitbucket" || gitProviderId == "bitbucket-server"
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

func getSigningKeyDescription(signingMethod string) string {
	switch signingMethod {
	case "gpg":
		return "Provide your GPG key ID (e.g., 30F2B65B9246B6CA) for signing commits."
	case "ssh":
		return "Provide your public SSH key (e.g., ssh-ed25519 AAAAC3...<rest of key>) for secure signing."
	default:
		return ""
	}
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

func getGitProviderSigningHelpMessage(gitProviderId string) string {
	signingDocsLink := config.GetDocsLinkForCommitSigning(gitProviderId)

	if signingDocsLink != "" {
		return signingDocsLink
	}
	return ""
}

func CheckIfAliasExists(alias string, initialAlias *string, existingAliases []string) error {
	for _, existingAlias := range existingAliases {
		if alias == existingAlias {
			if initialAlias == nil || *initialAlias != alias {
				return fmt.Errorf("alias '%s' is already in use", alias)
			}
		}
	}
	return nil
}

func ValidateSigningMethodAndKey(signingMethod string, signingKey, providerId string) error {
	if CommitSigningNotSupported(providerId) {
		return fmt.Errorf("commit signing is not supported for '%s' provider", providerId)
	} else {
		if signingMethod == "" || signingKey == "" {
			return fmt.Errorf("both signing method and key must be provided")
		}
		isValidSigningMethod := false
		for _, signingMethod := range apiclient.AllowedSigningMethodEnumValues {
			if signingMethod == apiclient.SigningMethod(signingMethod) {
				isValidSigningMethod = true
				break
			}
		}
		if !isValidSigningMethod {
			return fmt.Errorf("invalid signing method '%s'", signingMethod)
		}
		if signingMethod == "ssh" {
			if err := IsValidSSHKey(signingKey); err != nil {
				return err
			}
		}
	}
	return nil
}
