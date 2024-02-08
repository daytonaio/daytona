package project_agent_manager

import (
	"errors"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugin/project_agent"
	"github.com/daytonaio/daytona/plugin/project_agent/grpc/proto"
	"github.com/daytonaio/daytona/plugin/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

var projectAgentClients map[string]*plugin.Client = make(map[string]*plugin.Client)

var projectAgentHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROJECT_AGENT_PLUGIN",
	MagicCookieValue: "daytona_project_agent",
}

func GetProjectAgent(name string) (*ProjectAgent, error) {
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

	projectAgent, ok := raw.(ProjectAgent)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	return &projectAgent, nil
}

func GetProjectAgents() map[string]ProjectAgent {
	projectAgents := make(map[string]ProjectAgent)
	for name := range projectAgentClients {
		provisioner, err := GetProjectAgent(name)
		if err != nil {
			log.Printf("Error getting provisioner %s: %s", name, err)
			continue
		}

		projectAgents[name] = *provisioner
	}

	return projectAgents
}

func RegisterProjectAgent(pluginPath string) error {
	pluginName := path.Base(pluginPath)
	pluginBasePath := path.Dir(pluginPath)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: &utils.DebugLogWriter{},
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProjectAgentPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  projectAgentHandshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(pluginPath),
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	// TODO: create a cleanup or delete function that will kill the client
	// defer client.Kill()

	projectAgentClients[pluginName] = client

	log.Printf("Project Agent %s registered", pluginName)

	log.Infof("Provisioner %s registered", pluginName)

	projectAgent, err := GetProjectAgent(pluginName)
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	err = (*projectAgent).Initialize(&proto.InitializeProjectAgentRequest{
		BasePath: pluginBasePath,
	})
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	log.Infof("Provisioner %s initialized", pluginName)

	return nil
}
