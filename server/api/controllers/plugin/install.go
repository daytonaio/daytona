package plugin

import (
	"path"

	"github.com/daytonaio/daytona/plugins/plugin_manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/api/controllers/plugin/dto"
	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/frpc"
	"github.com/gin-gonic/gin"
)

// InstallProvisionerPlugin godoc
//
//	@Tags			plugin
//	@Summary		Install a provisioner plugin
//	@Description	Install a provisioner plugin
//	@Accept			json
//	@Produce		json
//	@Param			plugin	body	dto.InstallPluginRequest	true	"Plugin to install"
//	@Success		200
//	@Router			/plugin/install/provisioner [post]
//
//	@id				InstallProvisionerPlugin
func InstallProvisionerPlugin(ctx *gin.Context) {
	var req dto.InstallPluginRequest
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c, err := config.GetConfig()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	downloadPath := path.Join(c.PluginsDir, "provisioners", req.Name, req.Name)

	err = plugin_manager.DownloadPlugin(req.DownloadUrls, downloadPath)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = provisioner_manager.RegisterProvisioner(downloadPath, c.ServerDownloadUrl, frpc.GetServerUrl(c), frpc.GetApiUrl(c))
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}
