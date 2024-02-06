package main

import (
	"github.com/daytonaio/daytona/plugin"
	. "github.com/daytonaio/daytona/plugins/provisioner/docker-provisioner/plugin"
)

func GetProvisionerPlugin(basePath string) plugin.ProvisionerPlugin {
	return &DockerProvisionerPlugin{
		BasePath: basePath,
	}
}
