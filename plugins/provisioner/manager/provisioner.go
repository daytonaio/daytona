package provisioner_manager

import (
	"errors"
	"os"
	"os/exec"
	"path"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/plugins/provisioner"
	. "github.com/daytonaio/daytona/plugins/provisioner"
	"github.com/daytonaio/daytona/plugins/utils"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

type pluginRef struct {
	client *plugin.Client
	path   string
}

var pluginRefs map[string]*pluginRef = make(map[string]*pluginRef)

var ProvisionerHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROVISIONER_PLUGIN",
	MagicCookieValue: "daytona_provisioner",
}

func GetProvisioner(name string) (*Provisioner, error) {
	pluginRef, ok := pluginRefs[name]
	if !ok {
		return nil, errors.New("provisioner not found")
	}

	rpcClient, err := pluginRef.client.Client()
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
	for name := range pluginRefs {
		provisioner, err := GetProvisioner(name)
		if err != nil {
			log.Printf("Error getting provisioner %s: %s", name, err)
			continue
		}

		provisioners[name] = *provisioner
	}

	return provisioners
}

func RegisterProvisioner(pluginPath, serverDownloadUrl, serverUrl, serverApiUrl string) error {
	pluginName := path.Base(pluginPath)
	pluginBasePath := path.Dir(pluginPath)

	err := util.ChmodX(pluginPath)
	if err != nil {
		return errors.New("failed to chmod plugin: " + err.Error())
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: &utils.DebugLogWriter{},
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProvisionerPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ProvisionerHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
		Managed:         true,
	})

	pluginRefs[pluginName] = &pluginRef{
		client: client,
		path:   pluginBasePath,
	}

	log.Infof("Provisioner %s registered", pluginName)

	p, err := GetProvisioner(pluginName)
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	_, err = (*p).Initialize(provisioner.InitializeProvisionerRequest{
		BasePath:          pluginBasePath,
		ServerDownloadUrl: serverDownloadUrl,
		// TODO: get version from somewhere
		ServerVersion: "latest",
		ServerUrl:     serverUrl,
		ServerApiUrl:  serverApiUrl,
	})
	if err != nil {
		return errors.New("failed to initialize provisioner: " + err.Error())
	}

	log.Infof("Provisioner %s initialized", pluginName)

	return nil
}

func UninstallProvisioner(name string) error {
	pluginRef, ok := pluginRefs[name]
	if !ok {
		return errors.New("provisioner not found")
	}
	pluginRef.client.Kill()

	err := os.RemoveAll(pluginRef.path)
	if err != nil {
		return errors.New("failed to remove provisioner: " + err.Error())
	}

	delete(pluginRefs, name)

	return nil
}
