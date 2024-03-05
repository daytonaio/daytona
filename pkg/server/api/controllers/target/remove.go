package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/gin-gonic/gin"
)

// RemoveTarget godoc
//
//	@Tags			target
//	@Summary		Remove a target
//	@Description	Remove a target
//	@Param			target	path	string	true	"Target name"
//	@Success		204
//	@Router			/target/{target} [delete]
//
//	@id				RemoveTarget
func RemoveTarget(ctx *gin.Context) {
	targetName := ctx.Param("target")

	err := targets.RemoveTarget(targetName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove target: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
