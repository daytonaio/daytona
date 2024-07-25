// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package containerregistry

import (
	"github.com/daytonaio/daytona/internal/util"
	"github.com/spf13/cobra"
)

var ContainerRegistryCmd = &cobra.Command{
	Use:     "container-registry",
	Aliases: []string{"container-registries", "cr"},
	Short:   "Manage container registries",
	GroupID: util.SERVER_GROUP,
}

func init() {
	ContainerRegistryCmd.AddCommand(containerRegistryListCmd)
	ContainerRegistryCmd.AddCommand(containerRegistrySetCmd)
	ContainerRegistryCmd.AddCommand(containerRegistryDeleteCmd)
}
