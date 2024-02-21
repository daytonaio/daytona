package plugin

import (
	agent_service_manager "github.com/daytonaio/daytona/plugins/agent_service/manager"
	provisioner_manager "github.com/daytonaio/daytona/plugins/provisioner/manager"
	"github.com/daytonaio/daytona/server/api/controllers/plugin/dto"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ListProvisionerPlugins godoc
//
//	@Tags			plugin
//	@Summary		List provisioner plugins
//	@Description	List provisioner plugins
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	ProvisionerPlugin
//	@Router			/plugin/provisioner [get]
//
//	@id				ListProvisionerPlugins
func ListProvisionerPlugins(ctx *gin.Context) {
	provisioners := provisioner_manager.GetProvisioners()

	result := []dto.ProvisionerPlugin{}
	for _, provisioner := range provisioners {
		info, err := provisioner.GetInfo()
		if err != nil {
			log.Error(err)
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		result = append(result, dto.ProvisionerPlugin{
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
			log.Error(err)
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}

		result = append(result, dto.AgentServicePlugin{
			Name:    info.Name,
			Version: info.Version,
		})
	}

	ctx.JSON(200, result)
}
