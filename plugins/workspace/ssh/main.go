package main

import (
	. "github.com/daytonaio/daytona/plugin"
	. "github.com/daytonaio/daytona/plugins/workspace/ssh/plugin"
)

func GetWorkspacePlugin(basePath string) WorkspacePlugin {
	return &SshPlugin{
		PublicKey: "implement-me",
		BasePath:  basePath,
	}
}
