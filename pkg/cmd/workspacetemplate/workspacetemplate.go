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
	Args:    cobra.NoArgs,
	GroupID: util.TARGET_GROUP,
	Aliases: []string{"templates", "workspace-template", "workspace-templates", "wt"},
}

func init() {
	WorkspaceTemplateCmd.AddCommand(listCmd)
	WorkspaceTemplateCmd.AddCommand(infoCmd)
	WorkspaceTemplateCmd.AddCommand(createCmd)
	WorkspaceTemplateCmd.AddCommand(updateCmd)
	WorkspaceTemplateCmd.AddCommand(setDefaultCmd)
	WorkspaceTemplateCmd.AddCommand(deleteCmd)
	WorkspaceTemplateCmd.AddCommand(exportCmd)
	WorkspaceTemplateCmd.AddCommand(importCmd)
}
