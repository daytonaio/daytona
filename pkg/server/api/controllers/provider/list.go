package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/internal/util"
	provider_types "github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/provider/dto"
	"github.com/gin-gonic/gin"
)

// ListProviders godoc
//
//	@Tags			provider
//	@Summary		List providers
//	@Description	List providers
//	@Produce		json
//	@Success		200	{array}	dto.Provider
//	@Router			/provider [get]
//
//	@id				ListProviders
func ListProviders(ctx *gin.Context) {
	providers := manager.GetProviders()

	result := []dto.Provider{}
	for _, provider := range providers {
		info, err := provider.GetInfo()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider: %s", err.Error()))
			return
		}

		targets, err := provider.GetTargets()
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider targets: %s", err.Error()))
			return
		}

		result = append(result, dto.Provider{
			Name:    info.Name,
			Version: info.Version,
			Targets: util.ArrayMap(*targets, func(t provider_types.ProviderTarget) dto.TargetDTO {
				return dto.TargetDTO{
					Name:    t.Name,
					Options: t.Options,
				}
			}),
		})
	}

	ctx.JSON(200, result)
}
