package agent_service

import (
	"net/rpc"

	"github.com/daytonaio/daytona/common/types"
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
	Initialize(InitializeAgentServiceRequest) error
	GetInfo() (AgentServiceInfo, error)
	SetConfig(config *AgentServiceConfig) error
	ProjectPreInit(project *types.Project) error
	ProjectPostInit(project *types.Project) error
	ProjectPreStart(project *types.Project) error
	ProjectPostStart(project *types.Project) error
	ProjectPreStop(project *types.Project) error
	GetProjectInfo(project *types.Project) (*types.ProjectInfo, error)
	LivenessProbe() error
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
