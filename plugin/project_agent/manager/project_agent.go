package project_agent_manager

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugin/project_agent"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var ProjectAgents []ProjectAgent = []ProjectAgent{}

var projectAgentHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROJECT_AGENT_PLUGIN",
	MagicCookieValue: "daytona_project_agent",
}

func GetProjectAgent(name string) (*ProjectAgent, error) {
	//	todo
	return nil, errors.New("not implemented")
}

func GetProjectAgents() []ProjectAgent {
	return ProjectAgents
}

func RegisterProjectAgent(pluginPath string) {
	pluginName := path.Base(pluginPath)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProjectAgentPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: projectAgentHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
	})
	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		log.Fatal(err)
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		log.Fatal(err)
	}

	greeter := raw.(ProjectAgent)
	fmt.Println(greeter.GetName())
}
