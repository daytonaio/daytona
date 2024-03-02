package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/gin-gonic/gin"
)

// UninstallProvider godoc
//
//	@Tags			provider
//	@Summary		Uninstall a provider
//	@Description	Uninstall a provider
//	@Accept			json
//	@Param			provider	path	string	true	"Provider to uninstall"
//	@Success		200
//	@Router			/provider/{provider}/uninstall [post]
//
//	@id				UninstallProvider
func UninstallProvider(ctx *gin.Context) {
	provider := ctx.Param("provider")

	err := manager.UninstallProvider(provider)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to uninstall provider: %s", err.Error()))
		return
	}

	ctx.Status(200)
}
