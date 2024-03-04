package provider

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/provider/manager"
	"github.com/daytonaio/daytona/pkg/server/api/controllers/provider/dto"
	"github.com/gin-gonic/gin"
)

// GetTargetManifest godoc
//
//	@Tags			provider
//	@Summary		Get provider target manifest
//	@Description	Get provider target manifest
//	@Param			provider	path	string	true	"Provider name"
//	@Success		200
//	@Success		200	{object}	ProviderTargetManifest
//	@Router			/provider/{provider}/target-manifest [get]
//
//	@id				GetTargetManifest
func GetTargetManifest(ctx *gin.Context) {
	providerName := ctx.Param("provider")

	p, err := manager.GetProvider(providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("provider not found: %s", err.Error()))
		return
	}

	manifest, err := (*p).GetTargetManifest()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get provider manifest: %s", err.Error()))
		return
	}

	ctx.JSON(200, manifest)
}

// SetTarget godoc
//
//	@Tags			provider
//	@Summary		Set a provider target
//	@Description	Set a provider target
//	@Param			target		body	TargetDTO	true	"Provider target to set"
//	@Param			provider	path	string		true	"Provider name"
//	@Success		201
//	@Router			/provider/{provider}/target [put]
//
//	@id				SetTarget
func SetTarget(ctx *gin.Context) {
	providerName := ctx.Param("provider")
	var req dto.TargetDTO
	err := ctx.BindJSON(&req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	p, err := manager.GetProvider(providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("provider not found: %s", err.Error()))
		return
	}

	_, err = (*p).SetTarget(provider.ProviderTarget{
		Name:    req.Name,
		Options: req.Options,
	})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to set provider target: %s", err.Error()))
		return
	}

	ctx.Status(201)
}

// RemoveTarget godoc
//
//	@Tags			provider
//	@Summary		Set a provider target
//	@Description	Set a provider target
//	@Param			provider	path	string	true	"Provider name"
//	@Param			target		path	string	true	"Target name"
//	@Success		204
//	@Router			/provider/{provider}/{target} [delete]
//
//	@id				RemoveTarget
func RemoveTarget(ctx *gin.Context) {
	providerName := ctx.Param("provider")
	targetName := ctx.Param("target")

	p, err := manager.GetProvider(providerName)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, fmt.Errorf("provider not found: %s", err.Error()))
		return
	}

	_, err = (*p).RemoveTarget(targetName)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove provider target: %s", err.Error()))
		return
	}

	ctx.Status(204)
}
