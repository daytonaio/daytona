package target

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/targets"
	"github.com/gin-gonic/gin"
)

// ListTargets godoc
//
//	@Tags			target
//	@Summary		List targets
//	@Description	List targets
//	@Produce		json
//	@Success		200	{array}	ProviderTarget
//	@Router			/target [get]
//
//	@id				ListTargets
func ListTargets(ctx *gin.Context) {
	targets, err := targets.GetTargets()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get targets: %s", err.Error()))
		return
	}

	result := []provider.ProviderTarget{}
	for _, target := range targets {
		result = append(result, target)
	}

	ctx.JSON(200, result)
}
