package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/gin-gonic/gin"
)

// SetTarget godoc
//
//	@Tags			target
//	@Summary		Set a target
//	@Description	Set a target
//	@Param			target	body	ProviderTarget	true	"Target to set"
//	@Success		201
//	@Router			/target [put]
//
//	@id				SetTarget
func SetTarget(ctx *gin.Context) {
	var req provider.ProviderTarget
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	err = targets.SetTarget(req)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set target: %s", err.Error()))
		return
	}

	ctx.Status(201)
}
