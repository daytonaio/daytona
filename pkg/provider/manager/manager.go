// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	os_util "github.com/daytonaio/daytona/pkg/os"
	. "github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

type pluginRef struct {
	client *plugin.Client
	path   string
}

var pluginRefs map[string]*pluginRef = make(map[string]*pluginRef)

var ProviderHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROVIDER_PLUGIN",
	MagicCookieValue: "daytona_provider",
}

func GetProvider(name string) (*Provider, error) {
	pluginRef, ok := pluginRefs[name]
	if !ok {
		return nil, errors.New("provider not found")
	}

	rpcClient, err := pluginRef.client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return nil, err
	}

	provider, ok := raw.(Provider)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	return &provider, nil
}

func GetProviders() map[string]Provider {
	providers := make(map[string]Provider)
	for name := range pluginRefs {
		provider, err := GetProvider(name)
		if err != nil {
			log.Printf("Error getting provider %s: %s", name, err)
			continue
		}

		providers[name] = *provider
	}

	return providers
}

func RegisterProvider(pluginPath, serverDownloadUrl, serverUrl, serverApiUrl string) error {
	pluginName := path.Base(pluginPath)
	pluginBasePath := path.Dir(pluginPath)

	err := os_util.ChmodX(pluginPath)
	if err != nil {
		return errors.New("failed to chmod plugin: " + err.Error())
	}

	logger := hclog.New(&hclog.LoggerOptions{
		Name:   pluginName,
		Output: &util.DebugLogWriter{},
		Level:  hclog.Debug,
	})

	pluginMap := map[string]plugin.Plugin{}
	pluginMap[pluginName] = &ProviderPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ProviderHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
		Managed:         true,
	})

	pluginRefs[pluginName] = &pluginRef{
		client: client,
		path:   pluginBasePath,
	}

	log.Infof("Provider %s registered", pluginName)

	p, err := GetProvider(pluginName)
	if err != nil {
		return errors.New("failed to initialize provider: " + err.Error())
	}

	_, err = (*p).Initialize(InitializeProviderRequest{
		BasePath:          pluginBasePath,
		ServerDownloadUrl: serverDownloadUrl,
		// TODO: get version from somewhere
		ServerVersion: "latest",
		ServerUrl:     serverUrl,
		ServerApiUrl:  serverApiUrl,
	})
	if err != nil {
		return errors.New("failed to initialize provider: " + err.Error())
	}

	existingTargets, err := targets.GetTargets()
	if err != nil {
		return errors.New("failed to get targets: " + err.Error())
	}

	defaultTargets, err := (*p).GetDefaultTargets()
	if err != nil {
		return errors.New("failed to get default targets: " + err.Error())
	}

	log.Info("Setting default targets")
	for _, target := range *defaultTargets {
		if _, ok := existingTargets[target.Name]; ok {
			log.Infof("Target %s already exists. Skipping...", target.Name)
			continue
		}

		err := targets.SetTarget(target)
		if err != nil {
			log.Errorf("Failed to set target %s: %s", target.Name, err)
		} else {
			log.Infof("Target %s set", target.Name)
		}
	}
	log.Info("Default targets set")

	log.Infof("Provider %s initialized", pluginName)

	return nil
}

func UninstallProvider(name string) error {
	pluginRef, ok := pluginRefs[name]
	if !ok {
		return errors.New("provider not found")
	}
	pluginRef.client.Kill()

	err := os.RemoveAll(pluginRef.path)
	if err != nil {
		return errors.New("failed to remove provider: " + err.Error())
	}

	delete(pluginRefs, name)

	return nil
}

func TerminateProviderProcesses(providersBasePath string) error {
	process, err := process.Processes()

	if err != nil {
		return err
	}

	for _, p := range process {
		if e, err := p.Exe(); err == nil && strings.HasPrefix(e, providersBasePath) {
			err := p.Kill()
			if err != nil {
				log.Errorf("Failed to kill process %d: %s", p.Pid, err)
			}
		}
	}

	return nil
}
