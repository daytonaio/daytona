package plugin

import (
	"github.com/daytonaio/daytona/agent/workspace"
)

type WorkspacePluginConfig struct {
	EnvVars   map[string]string `json:"env_vars"`
	SetupPath string            `json:"setup_path"`
}

type WorkspacePlugin interface {
	GetName() string
	GetVersion() string
	SetConfig(config WorkspacePluginConfig) error
	ProjectPreInit(project workspace.Project) error
	ProjectInit(project workspace.Project) error
	ProjectStart(project workspace.Project) error
	ProjectInfo(project workspace.Project) string
	ProjectLivenessProbe(project workspace.Project) (bool, error)
	ProjectLivenessProbeTimeout() int
}

var workspacePlugins []WorkspacePlugin = []WorkspacePlugin{}

func GetWorkspacePlugins() []WorkspacePlugin {
	return workspacePlugins
}

func RegisterWorkspacePlugin(plugin WorkspacePlugin) {
	workspacePlugins = append(workspacePlugins, plugin)
}
