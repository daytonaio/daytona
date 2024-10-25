// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/jetbrains"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/daytonaio/daytona/pkg/telemetry"
	ide_views "github.com/daytonaio/daytona/pkg/views/ide"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CodeCmd = &cobra.Command{
	Use:     "code [TARGET] [PROJECT]",
	Short:   "Open a target in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var targetId string
		var projectName string
		var providerConfigId *string
		var ideId string
		var target *apiclient.TargetDTO

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		ideId = c.DefaultIdeId

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Verbose(true).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(true)
				return nil
			}

			target = selection.GetTargetFromPrompt(targetList, "Open")
			if target == nil {
				return nil
			}
			targetId = target.Id
		} else {
			target, err = apiclient_util.GetTarget(url.PathEscape(args[0]), true)
			if err != nil {
				if strings.Contains(err.Error(), targets.ErrTargetNotFound.Error()) {
					log.Debug(err)
					return errors.New("target not found. You can see all target names by running the command `daytona list`")
				}
				return err
			}
			targetId = target.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectTargetProject(targetId, &activeProfile)
			if err != nil {
				return err
			}
			if selectedProject == nil {
				return nil
			}

			projectName = selectedProject.Name
			providerConfigId = selectedProject.GitProviderConfigId
		}

		if len(args) == 2 {
			projectName = args[1]
			for _, project := range target.Projects {
				if project.Name == projectName {
					providerConfigId = project.GitProviderConfigId
					break
				}
			}
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		if !target_util.IsProjectRunning(target, projectName) {
			wsRunningStatus, err := AutoStartTarget(target.Name, projectName)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		providerMetadata := ""
		if ideId != "ssh" {
			providerMetadata, err = target_util.GetProjectProviderMetadata(target, projectName)
			if err != nil {
				return err
			}
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		yesFlag, _ := cmd.Flags().GetBool("yes")
		ideList := config.GetIdeList()
		ide_views.RenderIdeOpeningMessage(target.Name, projectName, ideId, ideList)
		return openIDE(ideId, activeProfile, targetId, projectName, providerMetadata, yesFlag, gpgKey)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 1 {
			return getProjectNameCompletions(cmd, args, toComplete)
		}

		return getTargetNameCompletions()
	},
}

func selectTargetProject(targetId string, profile *config.Profile) (*apiclient.Project, error) {
	ctx := context.Background()

	apiClient, err := apiclient_util.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	targetInfo, res, err := apiClient.TargetAPI.GetTarget(ctx, targetId).Execute()
	if err != nil {
		return nil, apiclient_util.HandleErrorResponse(res, err)
	}

	if len(targetInfo.Projects) > 1 {
		selectedProject := selection.GetProjectFromPrompt(targetInfo.Projects, "Open")
		if selectedProject == nil {
			return nil, nil
		}
		return selectedProject, nil
	} else if len(targetInfo.Projects) == 1 {
		return &targetInfo.Projects[0], nil
	}

	return nil, errors.New("no projects found in target")
}

func openIDE(ideId string, activeProfile config.Profile, targetId string, projectName string, projectProviderMetadata string, yesFlag bool, gpgKey string) error {
	telemetry.AdditionalData["ide"] = ideId

	switch ideId {
	case "vscode":
		return ide.OpenVSCode(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "code-insiders":
		return ide.OpenVSCodeInsiders(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "ssh":
		return ide.OpenTerminalSsh(activeProfile, targetId, projectName, gpgKey, nil)
	case "browser":
		return ide.OpenBrowserIDE(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "codium":
		return ide.OpenVScodium(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "codium-insiders":
		return ide.OpenVScodiumInsiders(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "cursor":
		return ide.OpenCursor(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "jupyter":
		return ide.OpenJupyterIDE(activeProfile, targetId, projectName, projectProviderMetadata, yesFlag, gpgKey)
	case "fleet":
		return ide.OpenFleet(activeProfile, targetId, projectName, gpgKey)
	case "positron":
		return ide.OpenPositron(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	case "zed":
		return ide.OpenZed(activeProfile, targetId, projectName, gpgKey)
	case "windsurf":
		return ide.OpenWindsurf(activeProfile, targetId, projectName, projectProviderMetadata, gpgKey)
	default:
		_, ok := jetbrains.GetIdes()[jetbrains.Id(ideId)]
		if ok {
			return ide.OpenJetbrainsIDE(activeProfile, ideId, targetId, projectName, gpgKey)
		}
	}

	return errors.New("invalid IDE. Please choose one by running `daytona ide`")
}

var ideFlag string

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))

	CodeCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")

}

func AutoStartTarget(targetId string, projectName string) (bool, error) {
	if !yesFlag {
		if !ide_views.RunStartTargetForm(targetId) {
			return false, nil
		}
	}

	apiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		return false, err
	}

	err = StartTarget(apiClient, targetId, projectName)
	if err != nil {
		return false, err
	}

	return true, nil
}
