package plugin

import (
	"fmt"
	"net/http"

	agent_service_manager "github.com/daytonaio/daytona/pkg/agent_service/manager"
	provisioner_manager "github.com/daytonaio/daytona/pkg/provisioner/manager"
	"github.com/gin-gonic/gin"
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
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall provisioner: %s", err.Error()))
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
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall agent service: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
