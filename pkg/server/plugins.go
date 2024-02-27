package server

import (
	"errors"
	"os"
	"path"

	agent_service_manager "github.com/daytonaio/daytona/pkg/agent_service/manager"
	"github.com/daytonaio/daytona/pkg/plugin_manager"
	provider_manager "github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/types"
	log "github.com/sirupsen/logrus"
)

func downloadDefaultPlugins() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	manifest, err := plugin_manager.GetPluginsManifest(c.PluginRegistryUrl)
	if err != nil {
		return err
	}

	defaultProviderPlugins := plugin_manager.GetDefaultPlugins(manifest.ProviderPlugins)
	defaultAgentServicePlugins := plugin_manager.GetDefaultPlugins(manifest.AgentServicePlugins)

	log.Info("Downloading default provider plugins")
	for pluginName, plugin := range defaultProviderPlugins {
		log.Info("Downloading " + pluginName)
		downloadPath := path.Join(c.PluginsDir, "providers", pluginName, pluginName)
		err = plugin_manager.DownloadPlugin(plugin.DownloadUrls, downloadPath)
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Downloading default agent service plugins")
	if len(defaultAgentServicePlugins) == 0 {
		log.Info("No default agent service plugins found")
	}

	for pluginName, plugin := range defaultAgentServicePlugins {
		log.Info("Downloading " + pluginName)
		downloadPath := path.Join(c.PluginsDir, "agent_services", pluginName, pluginName)
		err = plugin_manager.DownloadPlugin(plugin.DownloadUrls, downloadPath)
		if err != nil {
			log.Error(err)
		}
	}

	log.Info("Default plugins downloaded")

	return nil

}

func registerProviders(c *types.ServerConfig) error {
	log.Info("Registering providers")

	providerPluginsPath := path.Join(c.PluginsDir, "providers")

	files, err := os.ReadDir(providerPluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No providers found")
			return nil
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(path.Join(providerPluginsPath, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = provider_manager.RegisterProvider(pluginPath, c.ServerDownloadUrl, frpc.GetServerUrl(c), frpc.GetApiUrl(c))
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	log.Info("Providers registered")

	return nil
}

func registerAgentServices(c *types.ServerConfig) error {
	log.Info("Registering agent services")
	projectAgentPluginsPath := path.Join(c.PluginsDir, "agent_services")

	files, err := os.ReadDir(projectAgentPluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No agent services found")
			return nil
		}

		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(path.Join(projectAgentPluginsPath, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = agent_service_manager.RegisterAgentService(pluginPath)
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	log.Info("Agent services registered")

	return nil
}

func getPluginPath(dir string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if !file.IsDir() {
			return path.Join(dir, file.Name()), nil
		}
	}

	return "", errors.New("no plugin found in " + dir)
}
