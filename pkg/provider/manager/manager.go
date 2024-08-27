// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package manager

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	os_util "github.com/daytonaio/daytona/pkg/os"
	. "github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/providertargets"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

type pluginRef struct {
	client *plugin.Client
	path   string
	name   string
}

var ProviderHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "DAYTONA_PROVIDER_PLUGIN",
	MagicCookieValue: "daytona_provider",
}

type IProviderManager interface {
	DownloadProvider(downloadUrls map[os_util.OperatingSystem]string, providerName string, throwIfPresent bool) (string, error)
	GetProvider(name string) (*Provider, error)
	GetProviders() map[string]Provider
	GetProvidersManifest() (*ProvidersManifest, error)
	RegisterProvider(pluginPath string) error
	TerminateProviderProcesses(providersBasePath string) error
	UninstallProvider(name string) error
	Purge() error
}

type ProviderManagerConfig struct {
	DaytonaDownloadUrl       string
	ServerUrl                string
	ApiUrl                   string
	LogsDir                  string
	ProviderTargetService    providertargets.IProviderTargetService
	RegistryUrl              string
	BaseDir                  string
	CreateProviderNetworkKey func(providerName string) (string, error)
	ServerPort               uint32
	ApiPort                  uint32
}

func NewProviderManager(config ProviderManagerConfig) *ProviderManager {
	return &ProviderManager{
		pluginRefs:               make(map[string]*pluginRef),
		daytonaDownloadUrl:       config.DaytonaDownloadUrl,
		serverUrl:                config.ServerUrl,
		apiUrl:                   config.ApiUrl,
		logsDir:                  config.LogsDir,
		providerTargetService:    config.ProviderTargetService,
		registryUrl:              config.RegistryUrl,
		baseDir:                  config.BaseDir,
		createProviderNetworkKey: config.CreateProviderNetworkKey,
		serverPort:               config.ServerPort,
		apiPort:                  config.ApiPort,
	}
}

type ProviderManager struct {
	pluginRefs               map[string]*pluginRef
	daytonaDownloadUrl       string
	serverUrl                string
	apiUrl                   string
	serverPort               uint32
	apiPort                  uint32
	logsDir                  string
	providerTargetService    providertargets.IProviderTargetService
	registryUrl              string
	baseDir                  string
	createProviderNetworkKey func(providerName string) (string, error)
}

func (m *ProviderManager) GetProvider(name string) (*Provider, error) {
	pluginRef, ok := m.pluginRefs[name]
	if !ok {
		return nil, errors.New("provider not found")
	}

	p, err := m.dispenseProvider(pluginRef.client, name)
	if err != nil {
		// Attempt to reinitialize the provider
		pluginRef.client.Kill()
		pluginRef, err := m.initializeProvider(filepath.Join(pluginRef.path, name))
		if err != nil {
			return nil, err
		}

		m.pluginRefs[name] = pluginRef

		return m.dispenseProvider(pluginRef.client, name)
	}

	return p, nil
}

func (m *ProviderManager) GetProviders() map[string]Provider {
	providers := make(map[string]Provider)
	for name := range m.pluginRefs {
		provider, err := m.GetProvider(name)
		if err != nil {
			log.Printf("Error getting provider %s: %s", name, err)
			continue
		}

		providers[name] = *provider
	}

	return providers
}

func (m *ProviderManager) RegisterProvider(pluginPath string) error {
	pluginRef, err := m.initializeProvider(pluginPath)
	if err != nil {
		return err
	}

	m.pluginRefs[pluginRef.name] = pluginRef

	p, err := m.dispenseProvider(pluginRef.client, pluginRef.name)
	if err != nil {
		return err
	}

	existingTargets, err := m.providerTargetService.Map()
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

		err := m.providerTargetService.Save(&target)
		if err != nil {
			log.Errorf("Failed to set target %s: %s", target.Name, err)
		} else {
			log.Infof("Target %s set", target.Name)
		}
	}
	log.Info("Default targets set")

	log.Infof("Provider %s initialized", pluginRef.name)

	return nil
}

func (m *ProviderManager) UninstallProvider(name string) error {
	pluginRef, ok := m.pluginRefs[name]
	if !ok {
		return errors.New("provider not found")
	}
	pluginRef.client.Kill()

	err := os.RemoveAll(pluginRef.path)
	if err != nil {
		return errors.New("failed to remove provider: " + err.Error())
	}

	delete(m.pluginRefs, name)

	return nil
}

func (m *ProviderManager) TerminateProviderProcesses(providersBasePath string) error {
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

func (m *ProviderManager) Purge() error {
	for name := range m.pluginRefs {
		err := m.UninstallProvider(name)
		if err != nil {
			return err
		}
	}

	plugin.CleanupClients()

	return os.RemoveAll(m.baseDir)
}

func (m *ProviderManager) initializeProvider(pluginPath string) (*pluginRef, error) {
	pluginName := filepath.Base(pluginPath)
	pluginBasePath := filepath.Dir(pluginPath)

	if runtime.GOOS == "windows" && strings.HasSuffix(pluginPath, ".exe") {
		pluginName = strings.TrimSuffix(pluginName, ".exe")
	}

	err := os_util.ChmodX(pluginPath)
	if err != nil {
		return nil, errors.New("failed to chmod plugin: " + err.Error())
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

	log.Infof("Provider %s registered", pluginName)

	p, err := m.dispenseProvider(client, pluginName)
	if err != nil {
		return nil, errors.New("failed to initialize provider: " + err.Error())
	}

	networkKey, err := m.createProviderNetworkKey(pluginName)
	if err != nil {
		return nil, errors.New("failed to create network key: " + err.Error())
	}

	_, err = (*p).Initialize(InitializeProviderRequest{
		BasePath:           pluginBasePath,
		DaytonaDownloadUrl: m.daytonaDownloadUrl,
		DaytonaVersion:     internal.Version,
		ServerUrl:          m.serverUrl,
		ApiUrl:             m.apiUrl,
		LogsDir:            m.logsDir,
		NetworkKey:         networkKey,
		ServerPort:         m.serverPort,
		ApiPort:            m.apiPort,
	})
	if err != nil {
		return nil, errors.New("failed to initialize provider: " + err.Error())
	}

	return &pluginRef{
		client: client,
		path:   pluginBasePath,
		name:   pluginName,
	}, nil
}

func (m *ProviderManager) dispenseProvider(client *plugin.Client, name string) (*Provider, error) {
	rpcClient, err := client.Client()
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
