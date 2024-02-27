package agent_service

import (
	"net/rpc"

	"github.com/daytonaio/daytona/pkg/types"
	"github.com/hashicorp/go-plugin"
)

type AgentServiceInfo struct {
	Name    string
	Version string
}

type InitializeAgentServiceRequest struct {
	BasePath string
}

type AgentServiceConfig struct {
	SetupPath string
	EnvVars   map[string]string
}

type AgentService interface {
	Initialize(InitializeAgentServiceRequest) (*types.Empty, error)
	GetInfo() (AgentServiceInfo, error)
	SetConfig(config *AgentServiceConfig) (*types.Empty, error)
	ProjectPreInit(project *types.Project) (*types.Empty, error)
	ProjectPostInit(project *types.Project) (*types.Empty, error)
	ProjectPreStart(project *types.Project) (*types.Empty, error)
	ProjectPostStart(project *types.Project) (*types.Empty, error)
	ProjectPreStop(project *types.Project) (*types.Empty, error)
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
	LivenessProbe() (*types.Empty, error)
	LivenessProbeTimeout() uint32
}

type AgentServicePlugin struct {
	Impl AgentService
}

func (p *AgentServicePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &AgentServiceRPCServer{Impl: p.Impl}, nil
}

func (p *AgentServicePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &AgentServiceRPCClient{client: c}, nil
}
