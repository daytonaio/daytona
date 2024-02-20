package server

import (
	"github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/server/config"
	"github.com/gin-gonic/gin"
)

// GetConfig 			godoc
//
//	@Tags			server
//	@Summary		Get the server configuration
//	@Description	Get the server configuration
//	@Produce		json
//	@Success		200	{object}	types.ServerConfig
//	@Router			/server/config [get]
//
//	@id				GetConfig
func GetConfig(ctx *gin.Context) {
	config, err := config.GetConfig()
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, config)
}

// SetConfig 			godoc
//
//	@Tags			server
//	@Summary		Set the server configuration
//	@Description	Set the server configuration
//	@Accept			json
//	@Produce		json
//	@Param			config	body		types.ServerConfig	true	"Server configuration"
//	@Success		200		{object}	types.ServerConfig
//	@Router			/server/config [post]
//
//	@id				SetConfig
func SetConfig(ctx *gin.Context) {
	var c types.ServerConfig
	err := ctx.BindJSON(&c)
	if err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = config.Save(&c)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, c)
}
