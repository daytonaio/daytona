package plugin

import (
	"fmt"
	"net/http"

	agent_service_manager "github.com/daytonaio/daytona/pkg/agent_service/manager"
	provider_manager "github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/plugin/dto"
	"github.com/gin-gonic/gin"
)

// ListProviderPlugins godoc
//
//	@Tags			plugin
//	@Summary		List provider plugins
//	@Description	List provider plugins
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	ProviderPlugin
//	@Router			/plugin/provider [get]
//
//	@id				ListProviderPlugins
func ListProviderPlugins(ctx *gin.Context) {
	providers := provider_manager.GetProviders()

	result := []dto.ProviderPlugin{}
	for _, provider := range providers {
		info, err := provider.GetInfo()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider: %s", err.Error()))
			return
		}

		result = append(result, dto.ProviderPlugin{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	ctx.JSON(200, result)
}

// ListAgentServicePlugins godoc
//
//	@Tags			plugin
//	@Summary		List agent service plugins
//	@Description	List agent service plugins
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	AgentServicePlugin
//	@Router			/plugin/agent-service [get]
//
//	@id				ListAgentServicePlugins
func ListAgentServicePlugins(ctx *gin.Context) {
	agentServices := agent_service_manager.GetAgentServices()

	result := []dto.AgentServicePlugin{}
	for _, agentService := range agentServices {
		info, err := agentService.GetInfo()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get agent service: %s", err.Error()))
			return
		}

		result = append(result, dto.AgentServicePlugin{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	ctx.JSON(200, result)
}
