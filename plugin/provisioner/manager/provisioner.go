package provisioner_manager

import (
	"errors"
	"os"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugin/provisioner"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

var provisionerClients map[string]*plugin.Client = make(map[string]*plugin.Client)

var ProvisionerHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROVISIONER_PLUGIN",
	MagicCookieValue: "daytona_provisioner",
}

func GetProvisioner(name string) (*Provisioner, error) {
	client, ok := provisionerClients[name]
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

	provisioner, ok := raw.(Provisioner)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	return &provisioner, nil
}

func GetProvisioners() map[string]Provisioner {
	provisioners := make(map[string]Provisioner)
	for name := range provisionerClients {
		provisioner, err := GetProvisioner(name)
		if err != nil {
			log.Printf("Error getting provisioner %s: %s", name, err)
			continue
		}

		provisioners[name] = *provisioner
	}

	return provisioners
}

func RegisterProvisioner(pluginPath string) {
	pluginName := path.Base(pluginPath)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProvisionerPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  ProvisionerHandshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(pluginPath),
		Logger:           logger,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})
	// TODO: create a cleanup or delete function that will kill the client
	// defer client.Kill()

	provisionerClients[pluginName] = client

	log.Printf("Provisioner %s registered", pluginName)
}
