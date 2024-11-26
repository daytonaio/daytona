// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package gitprovider

import (
	"context"
	"fmt"
	"strings"

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
	Args:    cobra.RangeArgs(0, 1),
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

		if len(args) == 0 {
			err = gitprovider_view.GitProviderCreationView(ctx, apiClient, &setGitProviderConfig, existingAliases)
			if err != nil {
				return err
			}
		} else {
			supportedProviders := config.GetSupportedGitProviders()
			for _, gp := range supportedProviders {
				if gp.Id == args[0] {
					setGitProviderConfig.ProviderId = gp.Id
					break
				}
			}
			providerId := setGitProviderConfig.ProviderId

			if providerId == "" {
				supportedProvidersStr := ""
				for _, gp := range supportedProviders {
					supportedProvidersStr += fmt.Sprintf("%s, ", gp.Id)
				}
				return fmt.Errorf("'%s' is invalid or not a supported git provider.\nSupported providers are: %s", args[0], strings.TrimSuffix(supportedProvidersStr, ", "))
			}

			if tokenFlag == "" {
				return fmt.Errorf("token is required")
			}
			setGitProviderConfig.Token = tokenFlag

			if aliasFlag != "" {
				for _, alias := range existingAliases {
					if alias == aliasFlag {
						initialAlias := setGitProviderConfig.Alias
						if initialAlias == nil || *initialAlias != aliasFlag {
							return fmt.Errorf("alias '%s' is already in use", aliasFlag)
						}
					}
				}
				setGitProviderConfig.Alias = &aliasFlag
			}

			if gitprovider_view.ProviderRequiresUsername(providerId) {
				if usernameFlag == "" {
					return fmt.Errorf("username is required for '%s' provider", providerId)
				}
				setGitProviderConfig.Username = &usernameFlag
			}

			if gitprovider_view.ProviderRequiresApiUrl(providerId) {
				if baseApiUrlFlag == "" {
					return fmt.Errorf("base API URL is required for '%s' provider", providerId)
				}
				setGitProviderConfig.BaseApiUrl = &baseApiUrlFlag
			}

			if signingMethodFlag != "" || signingKeyFlag != "" {
				if gitprovider_view.CommitSigningNotSupported(providerId) {
					return fmt.Errorf("commit signing is not supported for '%s' provider", providerId)
				} else {
					if signingMethodFlag == "" || signingKeyFlag == "" {
						return fmt.Errorf("both signing method and key must be provided")
					}
					isValidSigningMethod := false
					for _, signingMethod := range apiclient.AllowedSigningMethodEnumValues {
						if signingMethod == apiclient.SigningMethod(signingMethodFlag) {
							setGitProviderConfig.SigningMethod = &signingMethod
							isValidSigningMethod = true
							break
						}
					}
					if !isValidSigningMethod {
						return fmt.Errorf("invalid signing method '%s'", signingMethodFlag)
					}
					if signingMethodFlag == "ssh" {
						if err := gitprovider_view.IsValidSSHKey(signingKeyFlag); err != nil {
							return err
						}
					}
					setGitProviderConfig.SigningKey = &signingKeyFlag
				}
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

var aliasFlag string
var usernameFlag string
var baseApiUrlFlag string
var tokenFlag string
var signingMethodFlag string
var signingKeyFlag string

func init() {
	GitProviderAddCmd.Flags().StringVarP(&aliasFlag, "alias", "a", "", "Alias")
	GitProviderAddCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Username")
	GitProviderAddCmd.Flags().StringVarP(&baseApiUrlFlag, "baseApiUrl", "b", "", "BaseApiUrl")
	GitProviderAddCmd.Flags().StringVarP(&tokenFlag, "token", "t", "", "Token")
	GitProviderAddCmd.Flags().StringVarP(&signingMethodFlag, "signing-method", "s", "", "SigningMethod")
	GitProviderAddCmd.Flags().StringVarP(&signingKeyFlag, "signing-key", "k", "", "SigningKey")
	GitProviderAddCmd.MarkFlagsRequiredTogether("signing-method", "signing-key")
}
