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
	"github.com/daytonaio/daytona/pkg/cmd/common"
	ws_create "github.com/daytonaio/daytona/pkg/cmd/workspace/create"
	"github.com/daytonaio/daytona/pkg/cmd/workspacetemplate"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/prebuild/create"
	"github.com/daytonaio/daytona/pkg/views/selection"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:     "create [WORKSPACE_CONFIG]",
	Short:   "Create a prebuild configuration",
	Args:    cobra.MaximumNArgs(1),
	Aliases: common.GetAliases("create"),
	RunE: func(cmd *cobra.Command, args []string) error {
		var prebuildAddView create.PrebuildAddView
		var workspaceTemplate *apiclient.WorkspaceTemplate
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
			views.RenderInfoMessage("No Git providers have been found - please create a Git provider using 'daytona git-provider create' in order to start using prebuilds.")
			return nil
		}

		// If no arguments and no flags are provided, run the interactive CLI
		if len(args) == 0 && branchFlag == "" && retentionFlag == 0 &&
			commitIntervalFlag == 0 && triggerFilesFlag == nil {
			// Interactive CLI logic

			workspaceTemplateList, res, err := apiClient.WorkspaceTemplateAPI.ListWorkspaceTemplates(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			workspaceTemplate = selection.GetWorkspaceTemplateFromPrompt(workspaceTemplateList, 0, false, true, "Prebuild")
			if workspaceTemplate == nil {
				return errors.New("No workspace template selected")
			}

			if workspaceTemplate.Name == selection.NewWorkspaceTemplateIdentifier {
				workspaceTemplate, err = workspacetemplate.RunWorkspaceTemplateAddFlow(apiClient, gitProviders, ctx)
				if err != nil {
					return err
				}
				if workspaceTemplate == nil {
					return nil
				}
			}

			prebuildAddView.WorkspaceTemplateName = workspaceTemplate.Name
			if workspaceTemplate.BuildConfig == nil {
				return errors.New("The chosen workspace template does not have a build configuration")
			}

			chosenBranch, err := ws_create.GetBranchFromWorkspaceTemplate(ctx, workspaceTemplate, apiClient, 0)
			if err != nil {
				return err
			}

			if chosenBranch == nil {
				fmt.Println("Operation canceled")
				return nil
			}
			prebuildAddView.RunBuildOnAdd = runFlag
			prebuildAddView.Branch = chosenBranch.Name
			create.PrebuildCreationView(&prebuildAddView, false)
		} else {
			// Non-interactive mode: use provided arguments and flags
			if len(args) > 0 {
				prebuildAddView.WorkspaceTemplateName = args[0]

				// Fetch the workspace template based on the provided argument
				workspaceTemplateTemp, res, err := apiClient.WorkspaceTemplateAPI.FindWorkspaceTemplate(ctx, prebuildAddView.WorkspaceTemplateName).Execute()
				if err != nil {
					return apiclient_util.HandleErrorResponse(res, err)
				}
				if workspaceTemplateTemp == nil {
					return errors.New("Invalid workspace template specified")
				}

				prebuildAddView.WorkspaceTemplateName = workspaceTemplateTemp.Name
				workspaceTemplate = workspaceTemplateTemp

			} else {
				return errors.New("Workspace template must be specified when using flags")
			}

			// Validate and handle required flags
			if branchFlag == "" {
				return errors.New("Branch flag is required when using flags")
			}
			prebuildAddView.Branch = branchFlag

			if retentionFlag <= 0 {
				return errors.New("Retention must be a positive integer")
			}
			prebuildAddView.Retention = strconv.Itoa(retentionFlag)

			if commitIntervalFlag > 0 {
				prebuildAddView.CommitInterval = strconv.Itoa(commitIntervalFlag)
			}

			prebuildAddView.TriggerFiles = triggerFilesFlag
			prebuildAddView.RunBuildOnAdd = runFlag
		}

		// Shared logic to create the prebuild configuration
		var commitInterval int
		if prebuildAddView.CommitInterval != "" {
			commitInterval, err = strconv.Atoi(prebuildAddView.CommitInterval)
			if err != nil {
				return errors.New("commit interval must be a number")
			}
		}
		var retention int

		if prebuildAddView.Retention != "" {
			retention, err = strconv.Atoi(prebuildAddView.Retention)
			if err != nil {
				return errors.New("retention must be a number")
			}
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

		prebuildId, res, err := apiClient.PrebuildAPI.SavePrebuild(ctx, prebuildAddView.WorkspaceTemplateName).Prebuild(newPrebuild).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		views.RenderInfoMessage("Prebuild added successfully")

		if prebuildAddView.RunBuildOnAdd {
			buildId, err := build.CreateBuild(apiClient, workspaceTemplate, prebuildAddView.Branch, &prebuildId)
			if err != nil {
				return err
			}

			views.RenderViewBuildLogsMessage(buildId)
		}

		return nil
	},
}

func init() {
	addCmd.Flags().BoolVar(&runFlag, "run", false, "Run the prebuild once after adding it")
	addCmd.Flags().StringVarP(&branchFlag, "branch", "b", "", "Git branch for the prebuild")
	addCmd.Flags().IntVarP(&retentionFlag, "retention", "r", 0, "Maximum number of resulting builds stored at a time")
	addCmd.Flags().IntVarP(&commitIntervalFlag, "commit-interval", "c", 0, "Commit interval for running a prebuild - leave blank to ignore push events")
	addCmd.Flags().StringSliceVarP(&triggerFilesFlag, "trigger-files", "t", nil, "Full paths of files whose changes should explicitly trigger a  prebuild")
}
