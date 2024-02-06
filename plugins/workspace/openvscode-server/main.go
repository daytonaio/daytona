package main

import (
	. "github.com/daytonaio/daytona/plugin"
	. "github.com/daytonaio/daytona/plugins/workspace/openvscode-server/plugin"
)

func GetWorkspacePlugin(basePath string) WorkspacePlugin {
	return &OpenVSCodeServerPlugin{
		BasePath: basePath,
	}
}
