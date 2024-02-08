package agent_service_manager

import (
	"errors"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugins/agent_service"
	"github.com/daytonaio/daytona/plugins/agent_service/grpc/proto"
	"github.com/daytonaio/daytona/plugins/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

var projectAgentClients map[string]*plugin.Client = make(map[string]*plugin.Client)

var projectAgentHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROJECT_AGENT_PLUGIN",
	MagicCookieValue: "daytona_agent_service",
}

func GetAgentService(name string) (*AgentService, error) {
	client, ok := projectAgentClients[name]
	if !ok {
		return nil, errors.New("provisioner not found")
	}

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return nil, err
	}

	projectAgent, ok := raw.(AgentService)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	return &projectAgent, nil
}

func GetAgentServices() map[string]AgentService {
	projectAgents := make(map[string]AgentService)
	for name := range projectAgentClients {
		provisioner, err := GetAgentService(name)
		if err != nil {
			log.Printf("Error getting provisioner %s: %s", name, err)
			continue
		}

		projectAgents[name] = *provisioner
	}

	return projectAgents
}

func RegisterAgentService(pluginPath string) error {
	pluginName := path.Base(pluginPath)
	pluginBasePath := path.Dir(pluginPath)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: &utils.DebugLogWriter{},
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &AgentServicePlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  projectAgentHandshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(pluginPath),
		Logger:           logger,
		Managed:          true,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	projectAgentClients[pluginName] = client

	log.Printf("Project Agent %s registered", pluginName)

	log.Infof("Provisioner %s registered", pluginName)

	projectAgent, err := GetAgentService(pluginName)
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	err = (*projectAgent).Initialize(&proto.InitializeAgentServiceRequest{
		BasePath: pluginBasePath,
	})
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	log.Infof("Provisioner %s initialized", pluginName)

	return nil
}
