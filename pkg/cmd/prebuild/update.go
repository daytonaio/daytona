// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package prebuild

import (
	"context"
	"errors"
	"strconv"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/build"
	"github.com/daytonaio/daytona/pkg/cmd/common"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/create"
	"github.com/daytonaio/daytona/pkg/views/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update [WORKSPACE_CONFIG] [PREBUILD_ID]",
	Short:   "Update a prebuild configuration",
	Args:    cobra.MaximumNArgs(2),
	Aliases: common.GetAliases("update"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView create.PrebuildAddView
		var prebuild *apiclient.PrebuildDTO
		var workspaceTemplateRecieved string
		var retention int
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		// Fetch the list of registered Git providers
		userGitProviders, res, err := apiClient.GitProviderAPI.ListGitProviders(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if len(userGitProviders) == 0 {
			views.RenderInfoMessage("No registered Git providers have been found - please register a Git provider using 'daytona git-provider create' in order to start using prebuilds.")
			return nil
		}

		// Determine the mode of operation: interactive or non-interactive
		if len(args) == 2 || (branchFlag != "" || retentionFlag != 0 || commitIntervalFlag != 0 || len(triggerFilesFlag) > 0) {
			// Non-interactive mode: use provided arguments and flags
			if len(args) < 2 {
				return errors.New("Both workspace template name and prebuild ID must be specified when using flags")
			}

			workspaceTemplateRecieved = args[0]
			prebuildID := args[1]

			prebuild, res, err = apiClient.PrebuildAPI.FindPrebuild(ctx, workspaceTemplateRecieved, prebuildID).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			// Set prebuild details based on flags
			if branchFlag != "" {
				prebuild.Branch = branchFlag
			}

			if retentionFlag > 0 {
				prebuild.Retention = int32(retentionFlag)
			}

			if commitIntervalFlag > 0 {
				prebuild.CommitInterval = util.Pointer(int32(commitIntervalFlag))
			}

			if len(triggerFilesFlag) > 0 {
				prebuild.TriggerFiles = triggerFilesFlag
			}
			prebuildAddView.Branch = prebuild.Branch
			prebuildAddView.Retention = strconv.Itoa(int(prebuild.Retention))
			prebuildAddView.WorkspaceTemplateName = workspaceTemplateRecieved
			prebuildAddView.TriggerFiles = prebuild.TriggerFiles
			prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
			retention = int(prebuild.Retention)
		} else {
			// Interactive mode: Prompt for details
			var prebuilds []apiclient.PrebuildDTO
			var selectedWorkspaceTemplateName string

			if len(args) == 1 {
				selectedWorkspaceTemplateName = args[0]
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuildsForWorkspaceTemplate(ctx, selectedWorkspaceTemplateName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			} else {
				prebuilds, res, err = apiClient.PrebuildAPI.ListPrebuilds(ctx).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
			}

			if len(prebuilds) == 0 {
				views_util.NotifyEmptyPrebuildList(true)
				return nil
			}

			prebuild = selection.GetPrebuildFromPrompt(prebuilds, "Update")
			if prebuild == nil {
				return nil
			}

			workspaceTemplateRecieved = prebuild.WorkspaceTemplateName
			prebuildAddView = create.PrebuildAddView{
				Branch:                prebuild.Branch,
				Retention:             strconv.Itoa(int(prebuild.Retention)),
				WorkspaceTemplateName: workspaceTemplateRecieved,
			}
			retention, err = strconv.Atoi(prebuildAddView.Retention)
			if err != nil {
				return errors.New("retention must be a number")
			}

			if prebuild.CommitInterval != nil {
				prebuildAddView.CommitInterval = strconv.Itoa(int(*prebuild.CommitInterval))
			}
			if len(prebuild.TriggerFiles) > 0 {
				prebuildAddView.TriggerFiles = prebuild.TriggerFiles
			}
			create.PrebuildCreationView(&prebuildAddView, false)
		}

		prebuildAddView.RunBuildOnAdd = runFlag

		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				return errors.New("commit interval must be a number")
			}
		}

		newPrebuild := apiclient.CreatePrebuildDTO{
			Id:        &prebuild.Id,
			Branch:    &prebuildAddView.Branch,
			Retention: int32(retention),
		}

		if commitInterval != 0 {
			newPrebuild.CommitInterval = util.Pointer(int32(commitInterval))
		}

		if len(prebuildAddView.TriggerFiles) > 0 {
			newPrebuild.TriggerFiles = prebuildAddView.TriggerFiles
		}

		prebuildId, res, err := apiClient.PrebuildAPI.SavePrebuild(ctx, prebuildAddView.WorkspaceTemplateName).Prebuild(newPrebuild).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild updated successfully")

		if prebuildAddView.RunBuildOnAdd {
			workspaceTemplate, res, err := apiClient.WorkspaceTemplateAPI.FindWorkspaceTemplate(ctx, prebuildAddView.WorkspaceTemplateName).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			buildId, err := build.CreateBuild(apiClient, workspaceTemplate, *newPrebuild.Branch, &prebuildId)
			if err != nil {
				return err
			}

			views.RenderViewBuildLogsMessage(buildId)
		}

		return nil
	},
}

var (
	branchFlag         string
	retentionFlag      int
	commitIntervalFlag int
	triggerFilesFlag   []string
	runFlag            bool
)

func init() {
	updateCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Git branch for the prebuild")
	updateCmd.Flags().IntVarP(&retentionFlag, "retention", "r", 0, "Maximum number of resulting builds stored at a time")
	updateCmd.Flags().IntVarP(&commitIntervalFlag, "commit-interval", "c", 0, "Commit interval for running a prebuild - leave blank to ignore push events")
	updateCmd.Flags().StringSliceVarP(&triggerFilesFlag, "trigger-files", "t", nil, "Full paths of files whose changes should explicitly trigger a  prebuild")
	updateCmd.Flags().BoolVar(&runFlag, "run", false, "Run the prebuild once after updating it")
}
