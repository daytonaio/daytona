package provisioner_manager

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	. "github.com/daytonaio/daytona/plugin/provisioner"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
)

var Provisioners []Provisioner = []Provisioner{}

var ProvisionerHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROVISIONER_PLUGIN",
	MagicCookieValue: "daytona_provisioner",
}

func GetProvisioner(name string) (*Provisioner, error) {
	//	todo
	return nil, errors.New("not implemented")
}

func GetProvisioners() []Provisioner {
	return Provisioners
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
		HandshakeConfig: ProvisionerHandshakeConfig,
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

	greeter := raw.(Provisioner)
	fmt.Println(greeter.GetName())
}
