package plugin

import (
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// UninstallProvisionerPlugin godoc
//
//	@Tags			plugin
//	@Summary		Uninstall a provisioner plugin
//	@Description	Uninstall a provisioner plugin
//	@Accept			json
//	@Param			provisioner	path	string	true	"Provisioner to uninstall"
//	@Success		200
//	@Router			/plugin/provisioner/{provisioner}/uninstall [post]
//
//	@id				UninstallProvisionerPlugin
func UninstallProvisionerPlugin(ctx *gin.Context) {
	provisioner := ctx.Param("provisioner")

	err := provisioner_manager.UninstallProvisioner(provisioner)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}

// UninstallAgentServicePlugin godoc
//
//	@Tags			plugin
//	@Summary		Uninstall an agent service plugin
//	@Description	Uninstall an agent service plugin
//	@Accept			json
//	@Param			agent-service	path	string	true	"Agent Service to uninstall"
//	@Success		200
//	@Router			/plugin/agent-service/uninstall [post]
//
//	@id				UninstallAgentServicePlugin
func UninstallAgentServicePlugin(ctx *gin.Context) {
	agentService := ctx.Param("agent-service")

	err := agent_service_manager.UninstallAgentService(agentService)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}
