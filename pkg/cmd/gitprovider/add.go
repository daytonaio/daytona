// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/spf13/cobra"
)

var GitProviderAddCmd = &cobra.Command{
	Use:     "add",
	Aliases: []string{"new", "register"},
	Short:   "Register a Git provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		existingAliases := util.ArrayMap(gitProviders, func(gp apiclient.GitProvider) string {
			return gp.Alias
		})

		for _, gp := range gitProviders {
			existingAliases = append(existingAliases, gp.Alias)
		}

		setGitProviderConfig := apiclient.SetGitProviderConfig{}
		setGitProviderConfig.BaseApiUrl = new(string)
		setGitProviderConfig.Username = new(string)
		setGitProviderConfig.Alias = new(string)

		if gitProviderFlag == "" {
			err = gitprovider_view.GitProviderCreationView(ctx, apiClient, &setGitProviderConfig, existingAliases)
			if err != nil {
				return err
			}
		} else {
			supportedProviders := config.GetSupportedGitProviders()
			for _, gp := range supportedProviders {
				if gp.Id == gitProviderFlag {
					setGitProviderConfig.ProviderId = gp.Id
					break
				}
			}
			providerId := setGitProviderConfig.ProviderId
			if providerId == "" {
				supprotedProvidersString := ""
				for _, gp := range supportedProviders {
					supprotedProvidersString += fmt.Sprintf("%s  ", gp.Id)
				}
				return fmt.Errorf("Invalid git provider %s .",gitProviderFlag)
			}
			if tokenFlag == "" {
				return fmt.Errorf("Token is required")
			}
			setGitProviderConfig.Token = tokenFlag
			if aliasFlag != "" {
				for _, alias := range existingAliases {
					if alias == aliasFlag {
						initialAliasValues := setGitProviderConfig.Alias
						if initialAliasValues == nil || *initialAliasValues != aliasFlag {
							return fmt.Errorf("Alias already exists")
						}
					}
				}
				setGitProviderConfig.Alias = &aliasFlag
			}
			if gitprovider_view.ProviderRequiresUsername(providerId) {
				if usernameFlag == "" {
					return fmt.Errorf("Username is required")
				}
				setGitProviderConfig.Username = &usernameFlag
			}
			if gitprovider_view.ProviderRequiresApiUrl(providerId) {
				if baseApiUrlFlag == "" {
					return fmt.Errorf("Base API URL is required")
				}
				setGitProviderConfig.BaseApiUrl = &baseApiUrlFlag
			}
			if signingKeyFlag != "" || signingMethodFlag != "" {
				if gitprovider_view.CommitSigningNotSupported(providerId) {
					return fmt.Errorf("Commit signing is not supported for this provider")
				}
				if signingKeyFlag == "" || signingMethodFlag == "" {
					return fmt.Errorf("Signing key and signing method are required")
				}
				isValidSigningMethod := false
				for _, method := range apiclient.AllowedSigningMethodEnumValues {
					if method == apiclient.SigningMethod(signingMethodFlag) {
						setGitProviderConfig.SigningMethod = &method
						isValidSigningMethod = true
						break
					}
				}
				if !isValidSigningMethod {
					return fmt.Errorf("Invalid signing method.")
				}
				if signingMethodFlag == "ssh" {
					if err := gitprovider_view.IsValidSSHKey(signingKeyFlag); err != nil {
						return err
					}
				}
				setGitProviderConfig.SigningKey = &signingKeyFlag
			}
		}

		if setGitProviderConfig.ProviderId == "" {
			return nil
		}

		res, err = apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(setGitProviderConfig).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Git provider has been registered")
		return nil
	},
}

var gitProviderFlag string
var aliasFlag string
var usernameFlag string
var tokenFlag string
var baseApiUrlFlag string
var signingKeyFlag string
var signingMethodFlag string

func init() {
	GitProviderAddCmd.Flags().StringVarP(&gitProviderFlag, "git-provider", "g", "", "Git provider")
	GitProviderAddCmd.Flags().StringVarP(&aliasFlag, "alias", "a", "", "Alias")
	GitProviderAddCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username")
	GitProviderAddCmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "Token")
	GitProviderAddCmd.Flags().StringVarP(&baseApiUrlFlag, "base-api-url", "b", "", "Base API URL")
	GitProviderAddCmd.Flags().StringVarP(&signingKeyFlag, "signing-key", "k", "", "Signing key")
	GitProviderAddCmd.Flags().StringVarP(&signingMethodFlag, "signing-method", "m", "", "Signing method")
	GitProviderAddCmd.MarkFlagsRequiredTogether("signing-key", "signing-method")
}
