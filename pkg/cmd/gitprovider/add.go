package gitprovider

import (
	"context"
	"fmt"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	"github.com/spf13/cobra"
)

func NewAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add",
		Aliases: []string{"new", "register"},
		Short:   "Register a Git provider",
		RunE:    runAdd,
	}

	cmd.Flags().String("type", "", "Git provider type (e.g., github, gitlab, bitbucket)")
	cmd.Flags().String("name", "", "Alias for the git provider")
	cmd.Flags().String("token", "", "Personal Access Token for authentication")
	cmd.Flags().String("url", "", "Self-Managed API URL (for self-hosted instances)")
	cmd.Flags().String("username", "", "Username (required for some providers)")

	return cmd
}

func runAdd(cmd *cobra.Command, args []string) error {
	providerType, _ := cmd.Flags().GetString("type")
	name, _ := cmd.Flags().GetString("name")
	token, _ := cmd.Flags().GetString("token")
	url, _ := cmd.Flags().GetString("url")
	username, _ := cmd.Flags().GetString("username")

	if providerType != "" && name != "" && token != "" {
		return addGitProviderNonInteractive(cmd.Context(), providerType, name, token, url, username)
	}

	return addGitProviderInteractive(cmd.Context())
}

func addGitProviderNonInteractive(ctx context.Context, providerType, name, token, url, username string) error {
	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return err
	}

	setGitProviderConfig := apiclient.SetGitProviderConfig{
		ProviderId: providerType,
		Alias:      &name,
		Token:      token,
	}

	if url != "" {
		setGitProviderConfig.BaseApiUrl = &url
	}
	if username != "" {
		setGitProviderConfig.Username = &username
	}

	res, err := apiClient.GitProviderAPI.SetGitProvider(ctx).GitProviderConfig(setGitProviderConfig).Execute()
	if err != nil {
		return apiclient_util.HandleErrorResponse(res, err)
	}

	views.RenderInfoMessage(fmt.Sprintf("Git provider '%s' has been registered", name))
	return nil
}

func addGitProviderInteractive(ctx context.Context) error {
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

	setGitProviderConfig := apiclient.SetGitProviderConfig{}
	setGitProviderConfig.BaseApiUrl = new(string)
	setGitProviderConfig.Username = new(string)
	setGitProviderConfig.Alias = new(string)

	err = gitprovider_view.GitProviderCreationView(ctx, apiClient, &setGitProviderConfig, existingAliases)
	if err != nil {
		return err
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
}
