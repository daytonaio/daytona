// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspacetemplate

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var WorkspaceTemplateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Manage workspace templates",
	Aliases: []string{"templates", "workspace-template", "workspace-templates", "wt"},
	GroupID: util.TARGET_GROUP,
}

func init() {
	WorkspaceTemplateCmd.AddCommand(workspaceTemplateListCmd)
	WorkspaceTemplateCmd.AddCommand(workspaceTemplateInfoCmd)
	WorkspaceTemplateCmd.AddCommand(workspaceAddCmd)
	WorkspaceTemplateCmd.AddCommand(workspaceTemplateUpdateCmd)
	WorkspaceTemplateCmd.AddCommand(workspaceTemplateSetDefaultCmd)
	WorkspaceTemplateCmd.AddCommand(workspaceTemplateDeleteCmd)
}
