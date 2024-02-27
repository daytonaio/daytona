package plugin

import (
	"fmt"
	"net/http"
	"path"

	agent_service_manager "github.com/daytonaio/daytona/pkg/agent_service/manager"
	"github.com/daytonaio/daytona/pkg/plugin_manager"
	provider_manager "github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/plugin/dto"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/gin-gonic/gin"
)

// InstallProviderPlugin godoc
//
//	@Tags			plugin
//	@Summary		Install a provider plugin
//	@Description	Install a provider plugin
//	@Accept			json
//	@Param			plugin	body	InstallPluginRequest	true	"Plugin to install"
//	@Success		200
//	@Router			/plugin/provider/install [post]
//
//	@id				InstallProviderPlugin
func InstallProviderPlugin(ctx *gin.Context) {
	var req dto.InstallPluginRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	downloadPath := path.Join(c.PluginsDir, "providers", req.Name, req.Name)

	err = plugin_manager.DownloadPlugin(req.DownloadUrls, downloadPath)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to download plugin: %s", err.Error()))
		return
	}

	err = provider_manager.RegisterProvider(downloadPath, c.ServerDownloadUrl, frpc.GetServerUrl(c), frpc.GetApiUrl(c))
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register provider: %s", err.Error()))
		return
	}

	ctx.Status(200)
}

// InstallAgentServicePlugin godoc
//
//	@Tags			plugin
//	@Summary		Install an agent service plugin
//	@Description	Install an agent service plugin
//	@Accept			json
//	@Param			plugin	body	InstallPluginRequest	true	"Plugin to install"
//	@Success		200
//	@Router			/plugin/agent-service/install [post]
//
//	@id				InstallAgentServicePlugin
func InstallAgentServicePlugin(ctx *gin.Context) {
	var req dto.InstallPluginRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	c, err := config.GetConfig()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get config: %s", err.Error()))
		return
	}

	downloadPath := path.Join(c.PluginsDir, "agent_services", req.Name, req.Name)

	err = plugin_manager.DownloadPlugin(req.DownloadUrls, downloadPath)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to download plugin: %s", err.Error()))
		return
	}

	err = agent_service_manager.RegisterAgentService(downloadPath)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to register agent service: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
