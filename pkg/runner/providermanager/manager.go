// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package providermanager

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/models"
	os_util "github.com/daytonaio/daytona/pkg/os"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/shirou/gopsutil/process"
	log "github.com/sirupsen/logrus"
)

const INITIAL_SETUP_LOCK_FILE_NAME = "initial-setup.lock"

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
	DownloadProvider(ctx context.Context, downloadUrls map[os_util.OperatingSystem]string, providerName string) (string, error)
	GetProvider(name string) (*provider.Provider, error)
	GetProviders() map[string]provider.Provider
	RegisterProvider(pluginPath string, manualInstall bool) error
	TerminateProviderProcesses(providersBasePath string) error
	UninstallProvider(name string) error
	Purge() error
}

type ProviderManagerConfig struct {
	GetTargetConfigMap       func(ctx context.Context) (map[string]*models.TargetConfig, error)
	CreateTargetConfig       func(ctx context.Context, name, options string, providerInfo models.ProviderInfo) error
	CreateProviderNetworkKey func(ctx context.Context, providerName string) (string, error)
	RunnerId                 string
	RunnerName               string
	DaytonaDownloadUrl       string
	ServerUrl                string
	ApiUrl                   string
	ApiKey                   string
	WorkspaceLogsDir         string
	TargetLogsDir            string
	BaseDir                  string
	ServerPort               uint32
	ApiPort                  uint32
	Logger                   *log.Logger
}

var providerManager *providerManagerImpl

func GetProviderManager(config *ProviderManagerConfig) *providerManagerImpl {
	if config != nil && providerManager != nil {
		config.Logger.Fatal("Provider manager already initialized")
	}

	if providerManager == nil {
		if config == nil {
			log.Fatal("Provider manager not initialized")
		}

		providerManager = &providerManagerImpl{
			pluginRefs:               make(map[string]*pluginRef),
			runnerId:                 config.RunnerId,
			runnerName:               config.RunnerName,
			daytonaDownloadUrl:       config.DaytonaDownloadUrl,
			serverUrl:                config.ServerUrl,
			apiUrl:                   config.ApiUrl,
			apiKey:                   config.ApiKey,
			workspaceLogsDir:         config.WorkspaceLogsDir,
			targetLogsDir:            config.TargetLogsDir,
			getTargetConfigMap:       config.GetTargetConfigMap,
			createTargetConfig:       config.CreateTargetConfig,
			baseDir:                  config.BaseDir,
			createProviderNetworkKey: config.CreateProviderNetworkKey,
			serverPort:               config.ServerPort,
			apiPort:                  config.ApiPort,
			logger:                   config.Logger,
		}
	}

	return providerManager
}

type providerManagerImpl struct {
	runnerId                 string
	runnerName               string
	pluginRefs               map[string]*pluginRef
	getTargetConfigMap       func(ctx context.Context) (map[string]*models.TargetConfig, error)
	createTargetConfig       func(ctx context.Context, name, options string, providerInfo models.ProviderInfo) error
	createProviderNetworkKey func(ctx context.Context, providerName string) (string, error)
	daytonaDownloadUrl       string
	serverUrl                string
	serverVersion            string
	apiUrl                   string
	apiKey                   string
	serverPort               uint32
	apiPort                  uint32
	workspaceLogsDir         string
	targetLogsDir            string
	baseDir                  string
	logger                   *log.Logger
}

func (m *providerManagerImpl) GetProvider(name string) (*provider.Provider, error) {
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

func (m *providerManagerImpl) GetProviders() map[string]provider.Provider {
	providers := make(map[string]provider.Provider)
	for name := range m.pluginRefs {
		provider, err := m.GetProvider(name)
		if err != nil {
			m.logger.Printf("Error getting provider %s: %s", name, err)
			continue
		}

		providers[name] = *provider
	}

	return providers
}

func (m *providerManagerImpl) RegisterProvider(pluginPath string, manualInstall bool) error {
	ctx := context.Background()

	pluginRef, err := m.initializeProvider(pluginPath)
	if err != nil {
		return err
	}

	m.pluginRefs[pluginRef.name] = pluginRef

	lockFilePath := filepath.Join(pluginRef.path, INITIAL_SETUP_LOCK_FILE_NAME)
	_, err = os.Stat(lockFilePath)
	if os.IsNotExist(err) || manualInstall {
		p, err := m.GetProvider(pluginRef.name)
		if err != nil {
			return fmt.Errorf("failed to get provider: %w", err)
		}

		providerInfo, err := (*p).GetInfo()
		if err != nil {
			return err
		}

		existingTargetConfigs, err := m.getTargetConfigMap(ctx)
		if err != nil {
			return errors.New("failed to get target configs: " + err.Error())
		}

		presetTargetConfigs, err := (*p).GetPresetTargetConfigs()
		if err != nil {
			return errors.New("failed to get preset target configs: " + err.Error())
		}

		m.logger.Infof("Setting preset target configs for %s", pluginRef.name)
		for _, targetConfig := range *presetTargetConfigs {
			if _, ok := existingTargetConfigs[targetConfig.Name]; ok {
				m.logger.Infof("Target config %s already exists. Skipping...", targetConfig.Name)
				continue
			}

			providerInfo.RunnerId = m.runnerId
			providerInfo.RunnerName = m.runnerName

			err = m.createTargetConfig(ctx, targetConfig.Name, targetConfig.Options, providerInfo)
			if err != nil {
				m.logger.Errorf("Failed to set target config %s: %s", targetConfig.Name, err)
			} else {
				m.logger.Infof("Target config %s set", targetConfig.Name)
			}
		}
		m.logger.Infof("Preset target configs set for %s", pluginRef.name)
	}

	m.logger.Infof("Provider %s initialized", pluginRef.name)

	return nil
}

func (m *providerManagerImpl) UninstallProvider(name string) error {
	pluginRef, ok := m.pluginRefs[name]
	if !ok {
		return errors.New("provider not found")
	}
	pluginRef.client.Kill()

	lockFileExisted := false
	lockFilePath := filepath.Join(pluginRef.path, INITIAL_SETUP_LOCK_FILE_NAME)
	_, err := os.Stat(lockFilePath)
	if err == nil {
		lockFileExisted = true
	}

	err = os.RemoveAll(pluginRef.path)
	if err != nil {
		return errors.New("failed to remove provider: " + err.Error())
	}

	if lockFileExisted {
		// After clearing up the contents, remake the directory and add a lock file that
		// will be used to ensure that the provider is not reinstalled automatically
		err = os.MkdirAll(pluginRef.path, os.ModePerm)
		if err != nil {
			return err
		}

		lockFilePath := filepath.Join(pluginRef.path, INITIAL_SETUP_LOCK_FILE_NAME)

		file, err := os.Create(lockFilePath)
		if err != nil {
			return err
		}
		defer file.Close()
	}

	delete(m.pluginRefs, name)

	return nil
}

func (m *providerManagerImpl) TerminateProviderProcesses(providersBasePath string) error {
	process, err := process.Processes()
	if err != nil {
		return err
	}

	for _, p := range process {
		if e, err := p.Exe(); err == nil && strings.HasPrefix(e, providersBasePath) {
			err := p.Kill()
			if err != nil {
				m.logger.Errorf("Failed to kill process %d: %s", p.Pid, err)
			}
		}
	}

	return nil
}

func (m *providerManagerImpl) Purge() error {
	for name := range m.pluginRefs {
		err := m.UninstallProvider(name)
		if err != nil {
			return err
		}
	}

	plugin.CleanupClients()

	return os.RemoveAll(m.baseDir)
}

func (m *providerManagerImpl) initializeProvider(pluginPath string) (*pluginRef, error) {
	ctx := context.Background()

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
	pluginMap[pluginName] = &provider.ProviderPlugin{}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: ProviderHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(pluginPath),
		Logger:          logger,
		Managed:         true,
	})

	m.logger.Infof("Provider %s registered", pluginName)

	p, err := m.dispenseProvider(client, pluginName)
	if err != nil {
		return nil, errors.New("failed to initialize provider: " + err.Error())
	}

	networkKey, err := m.createProviderNetworkKey(ctx, pluginName)
	if err != nil {
		return nil, errors.New("failed to create network key: " + err.Error())
	}

	_, err = (*p).Initialize(provider.InitializeProviderRequest{
		BasePath:           pluginBasePath,
		DaytonaDownloadUrl: m.daytonaDownloadUrl,
		DaytonaVersion:     m.serverVersion,
		ServerUrl:          m.serverUrl,
		ApiUrl:             m.apiUrl,
		ApiKey:             m.apiKey,
		WorkspaceLogsDir:   m.workspaceLogsDir,
		TargetLogsDir:      m.targetLogsDir,
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

func (m *providerManagerImpl) dispenseProvider(client *plugin.Client, name string) (*provider.Provider, error) {
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(name)
	if err != nil {
		return nil, err
	}

	provider, ok := raw.(provider.Provider)
	if !ok {
		return nil, errors.New("unexpected type from plugin")
	}

	return &provider, nil
}
