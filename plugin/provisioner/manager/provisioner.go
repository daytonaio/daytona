package provisioner_manager

import (
	"errors"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugin/provisioner"
	"github.com/daytonaio/daytona/plugin/provisioner/grpc/proto"
	"github.com/daytonaio/daytona/plugin/utils"
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

func RegisterProvisioner(pluginPath string) error {
	pluginName := path.Base(pluginPath)
	pluginBasePath := path.Dir(pluginPath)

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: &utils.DebugLogWriter{},
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProvisionerPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  ProvisionerHandshakeConfig,
		Plugins:          pluginMap,
		Cmd:              exec.Command(pluginPath),
		Logger:           logger,
		Managed:          true,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	provisionerClients[pluginName] = client

	log.Infof("Provisioner %s registered", pluginName)

	provisioner, err := GetProvisioner(pluginName)
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	err = (*provisioner).Initialize(&proto.InitializeProvisionerRequest{
		BasePath: pluginBasePath,
	})
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	log.Infof("Provisioner %s initialized", pluginName)

	return nil
}
