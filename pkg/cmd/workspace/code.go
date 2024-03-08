// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/internal/util/apiclient/server"
	"github.com/daytonaio/daytona/pkg/ports"
	view_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"

	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CodeCmd = &cobra.Command{
	Use:     "code [WORKSPACE] [PROJECT]",
	Short:   "Open a workspace in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspaceId string
		var projectName string
		var ideId string

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ideId = c.DefaultIdeId

		apiClient, err := server.GetApiClient(&activeProfile)
		if err != nil {
			log.Fatal(err)
		}

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Execute()
			if err != nil {
				log.Fatal(apiclient.HandleErrorResponse(res, err))
			}

			workspace := selection.GetWorkspaceFromPrompt(workspaceList, "open")
			if workspace == nil {
				return
			}
			workspaceId = *workspace.Id
		} else {
			workspace, err := server.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
			workspaceId = *workspace.Id
		}

		if len(args) == 0 || len(args) == 1 {
			selectedProject, err := selectWorkspaceProject(workspaceId, &activeProfile)
			if err != nil {
				log.Fatal(err)
			}
			if selectedProject == nil {
				return
			}
			projectName = *selectedProject
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		view_util.RenderInfoMessage(fmt.Sprintf("Opening the workspace project '%s' in your preferred IDE.", projectName))

		openIDE(ideId, activeProfile, workspaceId, projectName)
	},
}

func selectWorkspaceProject(workspaceId string, profile *config.Profile) (*string, error) {
	ctx := context.Background()

	apiClient, err := server.GetApiClient(profile)
	if err != nil {
		return nil, err
	}

	wsInfo, res, err := apiClient.WorkspaceAPI.GetWorkspace(ctx, workspaceId).Execute()
	if err != nil {
		return nil, apiclient.HandleErrorResponse(res, err)
	}

	if len(wsInfo.Projects) > 1 {
		selectedProject := selection.GetProjectFromPrompt(wsInfo.Projects, "open")
		if selectedProject == nil {
			return nil, nil
		}
		return selectedProject.Name, nil
	} else if len(wsInfo.Projects) == 1 {
		return wsInfo.Projects[0].Name, nil
	}

	return nil, errors.New("no projects found in workspace")
}

func openIDE(ideId string, activeProfile config.Profile, workspaceId string, projectName string) {
	if ideId == "browser" {
		err := openBrowserIDE(activeProfile, workspaceId, projectName)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	openVSCode(activeProfile, workspaceId, projectName)
}

func openVSCode(activeProfile config.Profile, workspaceId string, projectName string) {
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		log.Fatal(err)
	}

	checkAndAlertVSCodeInstalled()

	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, path.Join("/workspaces", projectName))

	var vscCommand *exec.Cmd = exec.Command("code", "--folder-uri", commandArgument)

	err = vscCommand.Run()

	if err != nil {
		log.Fatal(err.Error())
	}
}

func openBrowserIDE(activeProfile config.Profile, workspaceId string, projectName string) error {
	// Download and start IDE
	err := config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceId, projectName)
	if err != nil {
		return err
	}

	view_util.RenderInfoMessageBold("Downloading OpenVSCode Server...")
	projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceId, projectName)

	installServerCommand := exec.Command("ssh", projectHostname, "curl -fsSL https://download.daytona.io/daytona/get-openvscode-server.sh | sh")
	installServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
	installServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

	err = installServerCommand.Run()
	if err != nil {
		return err
	}

	view_util.RenderInfoMessageBold("Starting OpenVSCode Server...")

	go func() {
		startServerCommand := exec.CommandContext(context.Background(), "ssh", projectHostname, startVSCodeServerCommand)
		startServerCommand.Stdout = io.Writer(&util.DebugLogWriter{})
		startServerCommand.Stderr = io.Writer(&util.DebugLogWriter{})

		err = startServerCommand.Run()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Forward IDE port
	browserPort, errChan := ports.ForwardPort(workspaceId, projectName, 63000)
	if browserPort == nil {
		if err := <-errChan; err != nil {
			return err
		}
	}

	view_util.RenderInfoMessageBold(fmt.Sprintf("Forwarded %s IDE port to %d.\nOpening browser...", projectName, *browserPort))

	err = browser.OpenURL(fmt.Sprintf("http://localhost:%d", *browserPort))
	if err != nil {
		log.Error("Error opening URL: " + err.Error())
	}

	for {
		err := <-errChan
		if err != nil {
			// Log only in debug mode
			// Connection errors to the forwarded port should not exit the process
			log.Debug(err)
		}
	}
}

const startVSCodeServerCommand = "$HOME/vscode-server/bin/openvscode-server --start-server --port=63000 --host=0.0.0.0 --without-connection-token --disable-workspace-trust --default-folder=$DAYTONA_WS_DIR"

func checkAndAlertVSCodeInstalled() {
	if err := isVSCodeInstalled(); err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Visual Studio Code and ensure it's in your PATH. "
		infoMessage := "More information on: 'https://code.visualstudio.com/docs/editor/command-line#_launching-from-command-line'"

		log.Error(redBold + errorMessage + reset + infoMessage)

		return
	}
}

func isVSCodeInstalled() error {
	_, err := exec.LookPath("code")
	return err
}

var ideFlag string

func init() {
	CodeCmd.Flags().StringVarP(&ideFlag, "ide", "i", "", "Specify the IDE ('vscode' or 'browser')")
}
