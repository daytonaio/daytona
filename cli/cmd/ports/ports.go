// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd_ports

import (
	"github.com/spf13/cobra"
)

var portArg int

var PortsCmd = &cobra.Command{
	Use:   "ports",
	Short: "Manage forwarded project ports",
}

func init() {
	PortsCmd.AddCommand(portForwardCmd)
}
