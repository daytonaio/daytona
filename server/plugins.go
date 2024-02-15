package server

import (
	"errors"
	"os"
	"path"

	"github.com/daytonaio/daytona/common/grpc/proto/types"
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	"github.com/daytonaio/daytona/plugins/plugin_manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/frpc"
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

	defaultProvisionerPlugins := plugin_manager.GetDefaultPlugins(manifest.ProvisionerPlugins)
	defaultAgentServicePlugins := plugin_manager.GetDefaultPlugins(manifest.AgentServicePlugins)

	log.Info("Downloading default provisioner plugins")
	for pluginName, plugin := range defaultProvisionerPlugins {
		log.Info("Downloading " + pluginName)
		downloadPath := path.Join(c.PluginsDir, "provisioners", pluginName, pluginName)
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

func registerProvisioners(c *types.ServerConfig) error {
	log.Info("Registering provisioners")

	provisionerPluginsPath := path.Join(c.PluginsDir, "provisioners")

	files, err := os.ReadDir(provisionerPluginsPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("No provisioners found")
			return nil
		}
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pluginPath, err := getPluginPath(path.Join(provisionerPluginsPath, file.Name()))
			if err != nil {
				log.Error(err)
				continue
			}

			err = provisioner_manager.RegisterProvisioner(pluginPath, c.ServerDownloadUrl, frpc.GetServerUrl(c))
			if err != nil {
				log.Error(err)
				continue
			}
		}
	}

	log.Info("Provisioners registered")

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
