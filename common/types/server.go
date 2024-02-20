package types

type FRPSConfig struct {
	Domain   string `json:"domain"`
	Port     uint32 `json:"port"`
	Protocol string `json:"protocol"`
}

type ServerConfig struct {
	PluginsDir        string      `json:"pluginsDir"`
	PluginRegistryUrl string      `json:"pluginRegistryUrl"`
	Id                string      `json:"id"`
	ServerDownloadUrl string      `json:"serverDownloadUrl"`
	Frps              *FRPSConfig `json:"frps,omitempty"`
	ApiPort           uint32      `json:"apiPort"`
	HeadscalePort     uint32      `json:"headscalePort"`
}
