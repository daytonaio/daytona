// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package manager

import (
	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-plugin"
)

type pluginRef struct {
	client *plugin.Client
	impl   computeruse.IComputerUse
	path   string
}

var ComputerUseHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_COMPUTER_USE_PLUGIN",
	MagicCookieValue: "daytona_computer_use",
}

var computerUse = &pluginRef{}

// ComputerUseError represents a computer-use plugin error with context
type ComputerUseError struct {
	Type    string // "dependency", "system", "plugin"
	Message string
	Details string
}

func (e *ComputerUseError) Error() string {
	return e.Message
}

// KillComputerUse terminates the plugin client and clears the cached impl.
// Used by the daemon's /computeruse/stop HTTP handler.
func KillComputerUse() {
	if computerUse.client != nil {
		computerUse.client.Kill()
	}
	computerUse.client = nil
	computerUse.impl = nil
	computerUse.path = ""
}
