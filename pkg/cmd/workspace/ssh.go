// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	workspace_util "github.com/daytonaio/daytona/pkg/cmd/workspace/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/views"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	sshOptions []string
	edit       bool
)

var SshCmd = &cobra.Command{
	Use:     "ssh [WORKSPACE] [PROJECT] [CMD...]",
	Short:   "SSH into a project using the terminal",
	Args:    cobra.ArbitraryArgs,
	GroupID: util.WORKSPACE_GROUP,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.GetConfig()
		if err != nil {
			return err
		}

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			return err
		}

		ctx := context.Background()
		var workspace *apiclient.WorkspaceDTO
		var projectName string
		var providerConfigId *string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(workspaceList) == 0 {
				views_util.NotifyEmptyWorkspaceList(true)
				return nil
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "SSH Into")
			if workspace == nil {
				return nil
			}
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0], true)
			if err != nil {
				return err
			}
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspace.Id, &activeProfile)
			if err != nil {
				return err
			}
			if selectedProject == nil {
				return nil
			}
			projectName = selectedProject.Name
			providerConfigId = selectedProject.GitProviderConfigId
		}

		if len(args) >= 2 {
			projectName = args[1]
			for _, project := range workspace.Projects {
				if project.Name == projectName {
					providerConfigId = project.GitProviderConfigId
					break
				}
			}
		}

		if edit {
			err := editSSHConfig(activeProfile, workspace, projectName)
			if err != nil {
				return err
			}
			return nil
		}

		if !workspace_util.IsProjectRunning(workspace, projectName) {
			wsRunningStatus, err := AutoStartWorkspace(workspace.Name, projectName)
			if err != nil {
				return err
			}
			if !wsRunningStatus {
				return nil
			}
		}

		sshArgs := []string{}
		if len(args) > 2 {
			sshArgs = append(sshArgs, args[2:]...)
		}

		gpgKey, err := GetGitProviderGpgKey(apiClient, ctx, providerConfigId)
		if err != nil {
			log.Warn(err)
		}

		return ide.OpenTerminalSsh(activeProfile, workspace.Id, projectName, gpgKey, sshOptions, sshArgs...)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return getProjectNameCompletions(cmd, args, toComplete)
		}

		return getWorkspaceNameCompletions()
	},
}

func init() {
	SshCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	SshCmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit the project's SSH config")
	SshCmd.Flags().StringArrayVarP(&sshOptions, "option", "o", []string{}, "Specify SSH options in KEY=VALUE format.")
}

func editSSHConfig(activeProfile config.Profile, workspace *apiclient.WorkspaceDTO, projectName string) error {
	sshDir := filepath.Join(config.SshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")
	sshConfig, err := config.ReadSshConfig(configPath)
	if err != nil {
		return err
	}

	hostLine := fmt.Sprintf("Host %s", config.GetProjectHostname(activeProfile.Id, workspace.Id, projectName))
	regex := regexp.MustCompile(fmt.Sprintf(`%s\s*\n(?:\t.*\n?)*`, hostLine))
	matchedEntry := regex.FindString(sshConfig)
	if matchedEntry == "" {
		return fmt.Errorf("no SSH entry found for project %s", projectName)
	}

	lines := strings.Split(matchedEntry, "\n")
	if len(lines) > 0 {
		lines = lines[1:]
	}

	var proxyCommand string
	var filteredLines []string
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "ProxyCommand") {
			proxyCommand = trimmedLine
		} else {
			if trimmedLine != "" {
				filteredLines = append(filteredLines, trimmedLine)
			}
		}
	}
	modifiedContent := strings.Join(filteredLines, "\n")

	isCorrect := true
	formFields := []huh.Field{
		huh.NewText().
			Title("Edit SSH Config").
			Description(hostLine).
			CharLimit(-1).
			Value(&modifiedContent).ShowLineNumbers(true).WithHeight(10),
		huh.NewConfirm().
			Title("Is the above information correct?").
			Value(&isCorrect),
	}
	form := huh.NewForm(
		huh.NewGroup(formFields...),
	).WithTheme(views.GetCustomTheme())
	keyMap := huh.NewDefaultKeyMap()
	keyMap.Text = huh.TextKeyMap{
		NewLine: key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "new line")),
		Next:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "next")),
		Prev:    key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
	}

	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	if !isCorrect {
		return fmt.Errorf("operation canceled")
	}

	if modifiedContent == "" {
		err = config.RemoveWorkspaceSshEntries(activeProfile.Id, workspace.Id, projectName)
		if err != nil {
			return err
		}
		views.RenderInfoMessage(fmt.Sprintf("SSH configuration for %s removed successfully", projectName))

		return nil
	}

	updatedLines := strings.Split(modifiedContent, "\n")
	updatedLines = append(updatedLines, proxyCommand)
	modifiedContent = hostLine + "\n\t" + strings.Join(updatedLines, "\n\t")
	modifiedContent = strings.TrimSuffix(modifiedContent, "\t")
	if !strings.HasSuffix(modifiedContent, "\n") {
		modifiedContent += "\n"
	}

	err = config.UpdateWorkspaceSshEntry(activeProfile.Id, workspace.Id, projectName, modifiedContent)
	if err != nil {
		return err
	}

	views.RenderInfoMessage(fmt.Sprintf("SSH configuration for %s updated successfully", projectName))

	return nil
}
