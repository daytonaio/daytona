// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/build"
	"github.com/daytonaio/daytona/pkg/cmd/projectconfig"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/add"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	"github.com/spf13/cobra"
)

var prebuildAddCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add a prebuild configuration",
	Args:    cobra.NoArgs,
	Aliases: []string{"new", "create"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView add.PrebuildAddView
		var projectConfig *apiclient.ProjectConfig
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		gitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(gitProviders) == 0 {
			views.RenderInfoMessage("No registered Git providers have been found - please register a Git provider using 'daytona git-provider add' in order to start using prebuilds.")
			return nil
		}

		projectConfigList, res, err := apiClient.ProjectConfigAPI.ListProjectConfigs(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		projectConfig = selection.GetProjectConfigFromPrompt(projectConfigList, 0, false, true, "Prebuild")
		if projectConfig == nil {
			return errors.New("No project config selected")
		}

		if projectConfig.Name == selection.NewProjectConfigIdentifier {
			projectConfig, err = projectconfig.RunProjectConfigAddFlow(apiClient, gitProviders, ctx)
			if err != nil {
				return err
			}
			if projectConfig == nil {
				return nil
			}
		}

		prebuildAddView.ProjectConfigName = projectConfig.Name

		if projectConfig.BuildConfig == nil {
			return errors.New("The chosen project config does not have a build configuration")
		}

		chosenBranch, err := workspace_util.GetBranchFromProjectConfig(projectConfig, apiClient, 0)
		if err != nil {
			return err
		}

		if chosenBranch == nil {
			fmt.Println("Operation canceled")
			return nil
		}

		prebuildAddView.RunBuildOnAdd = runOnAddFlag
		prebuildAddView.Branch = chosenBranch.Name

		add.PrebuildCreationView(&prebuildAddView, false)

		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				return errors.New("commit interval must be a number")
			}
		}

		retention, err := strconv.Atoi(prebuildAddView.Retention)
		if err != nil {
			return errors.New("retention must be a number")

		}

		newPrebuild := apiclient.CreatePrebuildDTO{
			Branch:    &prebuildAddView.Branch,
			Retention: int32(retention),
		}

		if commitInterval != 0 {
			newPrebuild.CommitInterval = util.Pointer(int32(commitInterval))
		}

		if len(prebuildAddView.TriggerFiles) > 0 {
			newPrebuild.TriggerFiles = prebuildAddView.TriggerFiles
		}

		prebuildId, res, err := apiClient.PrebuildAPI.SetPrebuild(ctx, prebuildAddView.ProjectConfigName).Prebuild(newPrebuild).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild added successfully")

		if prebuildAddView.RunBuildOnAdd {
			buildId, err := build.CreateBuild(apiClient, projectConfig, chosenBranch.Name, &prebuildId)
			if err != nil {
				return err
			}

			views.RenderViewBuildLogsMessage(buildId)
		}

		return nil
	},
}

var runOnAddFlag bool

func init() {
	prebuildAddCmd.Flags().BoolVar(&runOnAddFlag, "run", false, "Run the prebuild once after adding it")
}
