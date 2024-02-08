// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_workspace

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	cmd_ports "github.com/daytonaio/daytona/cli/cmd/ports"
	views_util "github.com/daytonaio/daytona/cli/cmd/views/util"
	select_prompt "github.com/daytonaio/daytona/cli/cmd/views/workspace_select_prompt"
	"github.com/daytonaio/daytona/cli/config"
	"github.com/daytonaio/daytona/cli/connection"
	workspace_proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/internal/util"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/browser"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var CodeCmd = &cobra.Command{
	Use:     "code [WORKSPACE_NAME] [PROJECT_NAME]",
	Short:   "Open a workspace in your preferred IDE",
	Args:    cobra.RangeArgs(0, 2),
	Aliases: []string{"open"},
	Run: func(cmd *cobra.Command, args []string) {
		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		ctx := context.Background()
		var workspaceName string
		var projectName string
		var ideId string

		activeProfile, err := c.GetActiveProfile()
		if err != nil {
			log.Fatal(err)
		}

		ideId = c.DefaultIdeId

		conn, err := connection.Get(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		client := workspace_proto.NewWorkspaceServiceClient(conn)

		if len(args) == 0 {
			workspaceList, err := client.List(ctx, &empty.Empty{})
			if err != nil {
				log.Fatal(err)
			}

			workspaceName = select_prompt.GetWorkspaceNameFromPrompt(workspaceList.Workspaces, "open")
		} else {
			workspaceName = args[0]
		}

		wsName, wsMode := os.LookupEnv("DAYTONA_WS_NAME")
		if wsMode {
			workspaceName = wsName
		}

		// Todo: make project_select_prompt view for 0 args
		if len(args) == 0 || len(args) == 1 {
			projectName, err = util.GetFirstWorkspaceProjectName(conn, workspaceName, projectName)
			if err != nil {
				log.Fatal(err)
			}
		}

		if len(args) == 2 {
			projectName = args[1]
		}

		if ideFlag != "" {
			ideId = ideFlag
		}

		if ideId == "browser" {
			err = openBrowserIDE(conn, activeProfile, workspaceName, projectName)
			if err != nil {
				log.Fatal(err)
			}
			return
		}

		err = config.EnsureSshConfigEntryAdded(activeProfile.Id, workspaceName, projectName)
		if err != nil {
			log.Fatal(err)
		}

		checkAndAlertVSCodeInstalled()

		projectHostname := config.GetProjectHostname(activeProfile.Id, workspaceName, projectName)

		log.Info("Opening " + workspaceName + "'s project " + projectName + " in Visual Studio Code.")

		commandArgument := fmt.Sprintf("vscode-remote://ssh-remote+%s/%s", projectHostname, projectName)

		var vscCommand *exec.Cmd = exec.Command("code", "--folder-uri", commandArgument)

		err = vscCommand.Run()

		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func checkAndAlertVSCodeInstalled() {
	if err := isVSCodeInstalled(); err != nil {
		redBold := "\033[1;31m" // ANSI escape code for red and bold
		reset := "\033[0m"      // ANSI escape code to reset text formatting

		errorMessage := "Please install Visual Studio Code and ensure it's in your PATH."

		log.Error(redBold + errorMessage + reset)

		log.Info("More information on: 'https://code.visualstudio.com/docs/editor/command-line#_launching-from-command-line'")
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

func openBrowserIDE(conn *grpc.ClientConn, activeProfile config.Profile, workspaceName string, projectName string) error {
	projectPortForwards, err := cmd_ports.GetProjectPortForwards(conn, workspaceName, projectName)
	if err != nil {
		return err
	}

	browserPort := new(uint32)
	*browserPort = 63000

	errChan := make(chan error)
	if _, ok := projectPortForwards.PortForwards[63000]; !ok {
		browserPort, errChan = cmd_ports.ForwardPort(conn, activeProfile, workspaceName, projectName, uint32(63000))
		if browserPort == nil {
			if err = <-errChan; err != nil {
				return err
			}
		}
	} else {
		go func() {
			errChan <- nil
		}()
	}

	views_util.RenderInfoMessageBold(fmt.Sprintf("Port %d is being used to access the codebase.\nOpening %s using the browser IDE.", *browserPort, projectName))

	url := fmt.Sprintf("http://localhost:%d", *browserPort)

	err = browser.OpenURL(url)
	if err != nil {
		log.Fatal("Error opening URL: " + err.Error())
	}

	return <-errChan
}
