// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspace

import (
	"context"
	"os/exec"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/cmd/output"
	"github.com/daytonaio/daytona/pkg/views/workspace/info"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var InfoCmd = &cobra.Command{
	Use:     "info [WORKSPACE]",
	Short:   "Show workspace info",
	Aliases: []string{"view"},
	Args:    cobra.RangeArgs(0, 1),
	GroupID: util.WORKSPACE_GROUP,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		apiClient, err := apiclient_util.GetApiClient(nil)
		if err != nil {
			log.Fatal(err)
		}

		var workspace *apiclient.WorkspaceDTO

		if len(args) == 0 {
			workspaceList, res, err := apiClient.WorkspaceAPI.ListWorkspaces(ctx).Verbose(true).Execute()
			if err != nil {
				log.Fatal(apiclient_util.HandleErrorResponse(res, err))
			}

			workspace = selection.GetWorkspaceFromPrompt(workspaceList, "View")
		} else {
			workspace, err = apiclient_util.GetWorkspace(args[0])
			if err != nil {
				log.Fatal(err)
			}
		}

		if workspace == nil {
			return
		}

		if output.FormatFlag != "" {
			output.Output = workspace
			return
		}

		info.Render(workspace, "", false)

		displayUnpushedCommits()
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		return getWorkspaceNameCompletions()
	},
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func fetchLatestChanges() error {
	cmd := exec.Command("git", "fetch")
	err := cmd.Run()
	return err
}

func getUnpushedCommits(branch string) (int, error) {
	cmd := exec.Command("git", "rev-list", branch+"..origin/"+branch, "--count")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(output)))
}

func displayUnpushedCommits() {
	branch, err := getCurrentBranch()
	if err != nil {
		log.Warn("Failed to get current branch: ", err)
		return
	}

	err = fetchLatestChanges()
	if err != nil {
		log.Warn("Failed to fetch latest changes: ", err)
		return
	}

	unpushedCommits, err := getUnpushedCommits(branch)
	if err != nil {
		log.Warn("Failed to get unpushed commits: ", err)
		return
	}

	if unpushedCommits > 0 {
		log.Infof("Your branch is ahead of origin/%s by %d commit(s).", branch, unpushedCommits)
	} else {
		log.Info("Your branch is up-to-date with origin/", branch)
	}
}
