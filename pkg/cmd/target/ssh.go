// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

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
	target_util "github.com/daytonaio/daytona/pkg/cmd/target/util"
	"github.com/daytonaio/daytona/pkg/ide"
	"github.com/daytonaio/daytona/pkg/views"
	"github.com/daytonaio/daytona/pkg/views/target/selection"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	sshOptions []string
	edit       bool
)

var SshCmd = &cobra.Command{
	Use:     "ssh [TARGET] [WORKSPACE] [CMD...]",
	Short:   "SSH into a workspace using the terminal",
	Args:    cobra.ArbitraryArgs,
	GroupID: util.TARGET_GROUP,
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
		var target *apiclient.TargetDTO
		var workspaceName string
		var providerConfigId *string

		apiClient, err := apiclient_util.GetApiClient(&activeProfile)
		if err != nil {
			return err
		}

		if len(args) == 0 {
			targetList, res, err := apiClient.TargetAPI.ListTargets(ctx).Execute()
			if err != nil {
				return apiclient_util.HandleErrorResponse(res, err)
			}

			if len(targetList) == 0 {
				views_util.NotifyEmptyTargetList(true)
				return nil
			}

			target = selection.GetTargetFromPrompt(targetList, "SSH Into")
			if target == nil {
				return nil
			}
		} else {
			target, err = apiclient_util.GetTarget(args[0], true)
			if err != nil {
				return err
			}
		}

		if len(args) == 0 || len(args) == 1 {
			selectedWorkspace, err := selectTargetWorkspace(target.Id, &activeProfile)
			if err != nil {
				return err
			}
			if selectedWorkspace == nil {
				return nil
			}
			workspaceName = selectedWorkspace.Name
			providerConfigId = selectedWorkspace.GitProviderConfigId
		}

		if len(args) >= 2 {
			workspaceName = args[1]
			for _, workspace := range target.Workspaces {
				if workspace.Name == workspaceName {
					providerConfigId = workspace.GitProviderConfigId
					break
				}
			}
		}

		if edit {
			err := editSSHConfig(activeProfile, target, workspaceName)
			if err != nil {
				return err
			}
		}

		if !target_util.IsWorkspaceRunning(target, workspaceName) {
			tgRunningStatus, err := AutoStartTarget(target.Name, workspaceName)
			if err != nil {
				return err
			}
			if !tgRunningStatus {
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

		return ide.OpenTerminalSsh(activeProfile, target.Id, workspaceName, gpgKey, sshOptions, sshArgs...)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) >= 2 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			return getWorkspaceNameCompletions(cmd, args, toComplete)
		}

		return getTargetNameCompletions()
	},
}

func init() {
	SshCmd.Flags().BoolVarP(&yesFlag, "yes", "y", false, "Automatically confirm any prompts")
	SshCmd.Flags().BoolVarP(&edit, "edit", "e", false, "Edit the workspace's SSH config")
	SshCmd.Flags().StringArrayVarP(&sshOptions, "option", "o", []string{}, "Specify SSH options in KEY=VALUE format.")
}

func editSSHConfig(activeProfile config.Profile, target *apiclient.TargetDTO, workspaceName string) error {
	sshDir := filepath.Join(config.SshHomeDir, ".ssh")
	configPath := filepath.Join(sshDir, "daytona_config")
	sshConfig, err := config.ReadSshConfig(configPath)
	if err != nil {
		return err
	}

	hostLine := fmt.Sprintf("Host %s", config.GetWorkspaceHostname(activeProfile.Id, target.Id, workspaceName))
	regex := regexp.MustCompile(fmt.Sprintf(`%s\s*\n(?:\t.*\n?)*`, hostLine))
	matchedEntry := regex.FindString(sshConfig)
	if matchedEntry == "" {
		return fmt.Errorf("no SSH entry found for workspace %s", workspaceName)
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
		err = config.RemoveTargetSshEntries(activeProfile.Id, target.Id)
		if err != nil {
			return err
		}
		views.RenderInfoMessage(fmt.Sprintf("SSH configuration for %s removed successfully", workspaceName))

		return nil
	}

	trimmedContent := strings.TrimSpace(modifiedContent)
	updatedLines := strings.Split(trimmedContent, "\n")
	updatedLines = append(updatedLines, proxyCommand)
	var updatedLinesWithoutEmpty []string
	for _, line := range updatedLines {
		if line != "" {
			updatedLinesWithoutEmpty = append(updatedLinesWithoutEmpty, line)
		}
	}
	modifiedContent = hostLine + "\n\t" + strings.Join(updatedLinesWithoutEmpty, "\n\t")
	modifiedContent = strings.TrimSuffix(modifiedContent, "\t")
	if !strings.HasSuffix(modifiedContent, "\n") {
		modifiedContent += "\n"
	}

	err = config.UpdateWorkspaceSshEntry(activeProfile.Id, target.Id, workspaceName, modifiedContent)
	if err != nil {
		return err
	}

	views.RenderInfoMessage(fmt.Sprintf("SSH configuration for %s updated successfully", workspaceName))

	return nil
}
