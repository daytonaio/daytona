package manager

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/daytonaio/daemon/pkg/toolbox/computeruse"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log "github.com/sirupsen/logrus"
)

type pluginRef struct {
	client *plugin.Client
	impl   computeruse.IComputerUse
	path   string
}

var ComputerUseHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_COMPUTER_USE_PLUGIN",
	MagicCookieValue: "daytona_computer_use",
}

var computerUse = &pluginRef{}

func GetComputerUse(path string) (computeruse.IComputerUse, error) {
	if computerUse.impl != nil {
		return computerUse.impl, nil
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Infof("Computer use plugin not found at %s. Skipping...", path)
		return nil, nil
	}

	pluginName := filepath.Base(path)
	pluginBasePath := filepath.Dir(path)

	if runtime.GOOS == "windows" && strings.HasSuffix(path, ".exe") {
		pluginName = strings.TrimSuffix(pluginName, ".exe")
	}

	err := chmodX(path)
	if err != nil {
		return nil, errors.New("failed to chmod plugin: " + err.Error())
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name: pluginName,
		// Output: log.New().WriterLevel(log.DebugLevel),
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &computeruse.ComputerUsePlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ComputerUseHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path),
		Logger:          logger,
		Managed:         true,
	})

	log.Infof("Computer use %s registered", pluginName)

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		return nil, err
	}

	impl, ok := raw.(computeruse.IComputerUse)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	_, err = impl.Initialize()
	if err != nil {
		return nil, errors.New("failed to initialize computer use: " + err.Error())
	}

	computerUse.client = client
	computerUse.impl = impl
	computerUse.path = pluginBasePath

	return impl, nil
}

func chmodX(filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	err = f.Chmod(0755)
	if err != nil {
		return err
	}

	return nil
}
