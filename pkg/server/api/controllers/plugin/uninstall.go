package plugin

import (
	"fmt"
	"net/http"

	agent_service_manager "github.com/daytonaio/daytona/pkg/agent_service/manager"
	provider_manager "github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/gin-gonic/gin"
)

// UninstallProviderPlugin godoc
//
//	@Tags			plugin
//	@Summary		Uninstall a provider plugin
//	@Description	Uninstall a provider plugin
//	@Accept			json
//	@Param			provider	path	string	true	"Provider to uninstall"
//	@Success		200
//	@Router			/plugin/provider/{provider}/uninstall [post]
//
//	@id				UninstallProviderPlugin
func UninstallProviderPlugin(ctx *gin.Context) {
	provider := ctx.Param("provider")

	err := provider_manager.UninstallProvider(provider)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall provider: %s", err.Error()))
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
