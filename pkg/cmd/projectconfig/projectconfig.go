package projectconfig

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var ProjectConfigCmd = &cobra.Command{
  Use:     "project-config",
  Short:   "Manage project configs",
  Aliases: []string{"pc"},
  GroupID: util.WORKSPACE_GROUP,
}

func init() {
  // Ensure all command variables are initialized before adding
  if projectConfigListCmd == nil || projectConfigInfoCmd == nil ||
      projectConfigAddCmd == nil || projectConfigUpdateCmd == nil ||
      projectConfigSetDefaultCmd == nil || projectConfigDeleteCmd == nil ||
      importCmd == nil || exportCmd == nil {
      panic("One or more required commands are not initialized")
  }

  ProjectConfigCmd.AddCommand(projectConfigListCmd)
  ProjectConfigCmd.AddCommand(projectConfigInfoCmd)
  ProjectConfigCmd.AddCommand(projectConfigAddCmd)
  ProjectConfigCmd.AddCommand(projectConfigUpdateCmd)
  ProjectConfigCmd.AddCommand(projectConfigSetDefaultCmd)
  ProjectConfigCmd.AddCommand(projectConfigDeleteCmd)
  ProjectConfigCmd.AddCommand(importCmd)
  ProjectConfigCmd.AddCommand(exportCmd)
}