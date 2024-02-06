package main

import (
	"github.com/daytonaio/daytona/plugin"
	. "github.com/daytonaio/daytona/plugins/provisioner/ssh-provisioner/plugin"
)

func GetProvisionerPlugin(basePath string) plugin.ProvisionerPlugin {
	return &SshProvisionerPlugin{
		BasePath: basePath,
	}
}
