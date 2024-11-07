// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package create

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/daytonaio/daytona/internal/cmd/tailscale"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	ssh_config "github.com/daytonaio/daytona/pkg/agent/ssh/config"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_common "github.com/daytonaio/daytona/pkg/cmd/workspace/common"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/views"
	logs_view "github.com/daytonaio/daytona/pkg/views/logs"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/workspace"
	log "github.com/sirupsen/logrus"
	"tailscale.com/tsnet"

	"github.com/spf13/cobra"

	"github.com/daytonaio/daytona/cmd/daytona/config"
)

var CreateCmd = &cobra.Command{
	Use:     "create [REPOSITORY_URL | WORKSPACE_CONFIG_NAME]...",
	Short:   "Create a workspace",
	GroupID: util.TARGET_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		var createWorkspaceDtos []apiclient.CreateWorkspaceDTO
		var existingWorkspaceConfigNames []string
		var targetId string
		promptUsingTUI := len(args) == 0

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			return err
		}

		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		profileData, res, err := apiClient.ProfileAPI.GetProfileData(ctx).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		target, createTargetDto, err := GetTarget(ctx, GetTargetConfigParams{
			ApiClient:         apiClient,
			ActiveProfileName: activeProfile.Name,
			TargetNameFlag:    targetNameFlag,
			PromptUsingTUI:    promptUsingTUI,
		})
		if err != nil {
			if common.IsCtrlCAbort(err) {
				return nil
			}
			return err
		}

		targetName := ""
		if target != nil {
			targetName = target.Name
		} else if createTargetDto != nil {
			targetName = createTargetDto.Name
		}

		existingWorkspaces, res, err := apiClient.WorkspaceAPI.ListWorkspaces(context.Background()).Execute()
		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}

		if promptUsingTUI {
			err = ProcessPrompting(ctx, ProcessPromptingParams{
				ApiClient:                   apiClient,
				CreateWorkspaceDtos:         &createWorkspaceDtos,
				ExistingWorkspaces:          &existingWorkspaces,
				WorkspaceConfigurationFlags: workspaceConfigurationFlags,
				MultiWorkspaceFlag:          multiWorkspaceFlag,
				BlankFlag:                   blankFlag,
				TargetName:                  targetName,
			})
			if err != nil {
				if common.IsCtrlCAbort(err) {
					return nil
				} else {
					return err
				}
			}
		} else {
			existingWorkspaceConfigNames, err = ProcessCmdArguments(ctx, ProcessCmdArgumentsParams{
				ApiClient:                   apiClient,
				RepoUrls:                    args,
				CreateWorkspaceDtos:         &createWorkspaceDtos,
				ExistingWorkspaces:          &existingWorkspaces,
				WorkspaceConfigurationFlags: workspaceConfigurationFlags,
				BlankFlag:                   blankFlag,
			})
			if err != nil {
				return err
			}
		}

		workspaceNames := []string{}
		for i := range createWorkspaceDtos {
			if profileData != nil && profileData.EnvVars != nil {
				createWorkspaceDtos[i].EnvVars = util.MergeEnvVars(profileData.EnvVars, createWorkspaceDtos[i].EnvVars)
			} else {
				createWorkspaceDtos[i].EnvVars = util.MergeEnvVars(createWorkspaceDtos[i].EnvVars)
			}
			workspaceNames = append(workspaceNames, createWorkspaceDtos[i].Name)
		}

		names := append(workspaceNames, targetName)

		logs_view.SetupLongestPrefixLength(names)

		for i, workspaceConfigName := range existingWorkspaceConfigNames {
			if workspaceConfigName == "" {
				continue
			}
			logs_view.DisplayLogEntry(logs.LogEntry{
				WorkspaceName: &createWorkspaceDtos[i].Name,
				Msg:           fmt.Sprintf("Using detected workspace config '%s'\n", workspaceConfigName),
			}, i)
		}

		requestLogEntry := logs.LogEntry{
			Msg: views.GetPrettyLogLine("Request submitted"),
		}

		if target != nil {
			requestLogEntry.TargetName = &target.Name
		} else if createTargetDto != nil {
			requestLogEntry.TargetName = &createTargetDto.Name
		}

		logs_view.DisplayLogEntry(requestLogEntry, logs_view.STATIC_INDEX)

		activeProfile, err = c.GetActiveProfile()
		if err != nil {
			return err
		}

		if target != nil {
			targetId = target.Id
		} else if createTargetDto != nil {
			targetId = createTargetDto.Id
		}

		logsContext, stopLogs := context.WithCancel(context.Background())
		defer stopLogs()

		if createTargetDto != nil {
			go apiclient_util.ReadTargetLogs(logsContext, apiclient_util.ReadLogParams{
				Id:                    targetId,
				Label:                 &createTargetDto.Name,
				ActiveProfile:         activeProfile,
				Follow:                util.Pointer(true),
				SkipPrefixLengthSetup: true,
			})

			t, res, err := apiClient.TargetAPI.CreateTarget(ctx).Target(apiclient.CreateTargetDTO{
				Id:               createTargetDto.Id,
				Name:             createTargetDto.Name,
				TargetConfigName: createTargetDto.TargetConfigName,
			}).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
			target = &apiclient.TargetDTO{
				Id:           t.Id,
				Name:         t.Name,
				Options:      t.Options,
				ProviderInfo: t.ProviderInfo,
				Default:      t.Default,
			}
		}

		var tsConn *tsnet.Server
		if !IsLocalDockerTarget(target) || activeProfile.Id != "default" {
			tsConn, err = tailscale.GetConnection(&activeProfile)
			if err != nil {
				return err
			}
		}

		for i := range createWorkspaceDtos {
			createWorkspaceDtos[i].TargetId = targetId
			go apiclient_util.ReadWorkspaceLogs(logsContext, apiclient_util.ReadLogParams{
				Id:                    createWorkspaceDtos[i].Id,
				Label:                 &createWorkspaceDtos[i].Name,
				ActiveProfile:         activeProfile,
				Index:                 util.Pointer(i),
				Follow:                util.Pointer(true),
				SkipPrefixLengthSetup: true,
			})

			_, res, err = apiClient.WorkspaceAPI.CreateWorkspace(ctx).Workspace(createWorkspaceDtos[i]).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}
		}

		if err != nil {
			return apiclient_util.HandleErrorResponse(res, err)
		}
		gpgKey, err := workspace_common.GetGitProviderGpgKey(apiClient, ctx, createWorkspaceDtos[0].GitProviderConfigId)
		if err != nil {
			log.Warn(err)
		}

		err = waitForDial(target, createWorkspaceDtos[0].Id, &activeProfile, tsConn, gpgKey)
		if err != nil {
			return err
		}

		stopLogs()

		// Make sure terminal cursor is reset
		fmt.Print("\033[?25h")

		chosenIdeId := c.DefaultIdeId
		if IdeFlag != "" {
			chosenIdeId = IdeFlag
		}

		ideList := config.GetIdeList()
		var chosenIde config.Ide

		for _, ide := range ideList {
			if ide.Id == chosenIdeId {
				chosenIde = ide
			}
		}

		fmt.Println()

		ws, err := apiclient_util.GetWorkspace(createWorkspaceDtos[0].Id, true)
		if err != nil {
			return err
		}

		info.Render(ws, chosenIde.Name, false)

		if noIdeFlag {
			views.RenderCreationInfoMessage("Run 'daytona code' when you're ready to start developing")
			return nil
		}

		views.RenderCreationInfoMessage(fmt.Sprintf("Opening the workspace in %s ...", chosenIde.Name))

		return workspace_common.OpenIDE(chosenIdeId, activeProfile, createWorkspaceDtos[0].Name, *ws.Info.ProviderMetadata, YesFlag, gpgKey)
	},
}

var YesFlag bool
var targetNameFlag string
var IdeFlag string
var noIdeFlag bool
var blankFlag bool
var multiWorkspaceFlag bool

var workspaceConfigurationFlags = workspace_common.WorkspaceConfigurationFlags{
	Builder:           new(views_util.BuildChoice),
	CustomImage:       new(string),
	CustomImageUser:   new(string),
	Branches:          new([]string),
	DevcontainerPath:  new(string),
	EnvVars:           new([]string),
	Manual:            new(bool),
	GitProviderConfig: new(string),
}

func init() {
	ideList := config.GetIdeList()
	ids := make([]string, len(ideList))
	for i, ide := range ideList {
		ids[i] = ide.Id
	}
	ideListStr := strings.Join(ids, ", ")

	CreateCmd.Flags().StringVarP(&IdeFlag, "ide", "i", "", fmt.Sprintf("Specify the IDE (%s)", ideListStr))
	CreateCmd.Flags().StringVarP(&targetNameFlag, "target", "t", "", "Specify the target (e.g. 'local')")
	CreateCmd.Flags().BoolVar(&blankFlag, "blank", false, "Create a blank workspace without using existing configurations")
	CreateCmd.Flags().BoolVarP(&noIdeFlag, "no-ide", "n", false, "Do not open the target in the IDE after target creation")
	CreateCmd.Flags().BoolVar(&multiWorkspaceFlag, "multi-workspace", false, "Target with multiple workspaces/repos")
	CreateCmd.Flags().BoolVarP(&YesFlag, "yes", "y", false, "Automatically confirm any prompts")
	CreateCmd.Flags().StringSliceVar(workspaceConfigurationFlags.Branches, "branch", []string{}, "Specify the Git branches to use in the workspaces")

	workspace_common.AddWorkspaceConfigurationFlags(CreateCmd, workspaceConfigurationFlags, true)
}

func waitForDial(target *apiclient.TargetDTO, workspaceId string, activeProfile *config.Profile, tsConn *tsnet.Server, gpgKey string) error {
	if IsLocalDockerTarget(target) && (activeProfile != nil && activeProfile.Id == "default") {
		err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, gpgKey)
		if err != nil {
			return err
		}

		workspaceHostname := config.GetWorkspaceHostname(activeProfile.Id, workspaceId)

		for {
			sshCommand := exec.Command("ssh", workspaceHostname, "daytona", "version")
			sshCommand.Stdin = nil
			sshCommand.Stdout = nil
			sshCommand.Stderr = &util.TraceLogWriter{}

			err = sshCommand.Run()
			if err == nil {
				return nil
			}

			time.Sleep(time.Second)
		}
	}

	connectChan := make(chan error)
	spinner := time.After(15 * time.Second)
	timeout := time.After(2 * time.Minute)

	go func() {
		for {
			dialConn, err := tsConn.Dial(context.Background(), "tcp", fmt.Sprintf("%s:%d", workspace.GetWorkspaceHostname(workspaceId), ssh_config.SSH_PORT))
			if err == nil {
				connectChan <- dialConn.Close()
				return
			}
			time.Sleep(time.Second)
		}
	}()

	select {
	case err := <-connectChan:
		return err
	case <-spinner:
		err := views_util.WithInlineSpinner("Connection to tailscale is taking longer than usual", func() error {
			select {
			case err := <-connectChan:
				return err
			case <-timeout:
				return errors.New("secure connection to the Daytona Server could not be established. Please check your internet connection or Tailscale availability")
			}
		})
		return err
	}
}

func IsLocalDockerTarget(target *apiclient.TargetDTO) bool {
	if target.ProviderInfo.Name != "docker-provider" {
		return false
	}

	return !strings.Contains(target.Options, "Remote Hostname")
}
