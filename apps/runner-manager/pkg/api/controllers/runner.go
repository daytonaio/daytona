// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package controllers

import (
	"net/http"

	"github.com/daytonaio/runner-manager/pkg/api/dto"
	"github.com/daytonaio/runner-manager/pkg/provider"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Add 			godoc
//
//	@Tags			runner
//	@Summary		Add runner instances
//	@Description	Add runner instances
//	@Param			runner	body	dto.AddRunnerDTO	false	"Add runner"
//	@Produce		json
//	@Success		201	{object}	map[string]interface{}
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/runners/add [post]
//
//	@id				Add
func Add(ctx *gin.Context) {
	var addRunnerDto dto.AddRunnerDTO
	err := ctx.ShouldBindJSON(&addRunnerDto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default value of 1 if instances is not provided
	instances := 1
	if addRunnerDto.Instances != nil {
		instances = *addRunnerDto.Instances
	}

	log.Infof("Adding %d runner instance(s)", instances)

	// Get the provider manager
	providerManager := provider.GetInstance()
	runnerProvider, err := providerManager.GetProvider()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Add runners using the provider
	response, err := runnerProvider.AddRunners(ctx.Request.Context(), instances)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"job_id":    response.JobID,
		"pod_names": response.PodNames,
		"message":   response.Message,
		"instances": instances,
		"provider":  runnerProvider.GetProviderName(),
	})
}

// Remove 			godoc
//
//	@Tags			runner
//	@Summary		Remove runner instances
//	@Description	Remove runner instances
//	@Param			runner	body	dto.AddRunnerDTO	false	"Remove runner"
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Failure		400	{object}	map[string]string
//	@Failure		500	{object}	map[string]string
//	@Router			/runners/remove [post]
//
//	@id				Remove
func Remove(ctx *gin.Context) {
	var removeRunnerDto dto.AddRunnerDTO
	err := ctx.ShouldBindJSON(&removeRunnerDto)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set default value of 1 if instances is not provided
	instances := 1
	if removeRunnerDto.Instances != nil {
		instances = *removeRunnerDto.Instances
	}

	log.Infof("Removing %d runner instance(s)", instances)

	// Get the provider manager
	providerManager := provider.GetInstance()
	runnerProvider, err := providerManager.GetProvider()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Remove runners using the provider
	err = runnerProvider.RemoveRunners(ctx.Request.Context(), instances)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message":   "Runner instances removed",
		"instances": instances,
		"provider":  runnerProvider.GetProviderName(),
	})
}

// List 			godoc
//
//	@Tags			runner
//	@Summary		List all runner instances
//	@Description	List all runner instances
//	@Produce		json
//	@Success		200	{array}		types.RunnerInfo
//	@Failure		500	{object}	map[string]string
//	@Router			/runners [get]
//
//	@id				List
func List(ctx *gin.Context) {
	log.Info("Listing runner instances")

	// Get the provider manager
	providerManager := provider.GetInstance()
	runnerProvider, err := providerManager.GetProvider()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// List runners using the provider
	runners, err := runnerProvider.ListRunners(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, runners)
}

// Get 			godoc
//
//	@Tags			runner
//	@Summary		Get a specific runner instance
//	@Description	Get a specific runner instance
//	@Produce		json
//	@Param			runnerId	path		string	true	"Runner ID"
//	@Success		200			{object}	types.RunnerInfo
//	@Failure		404			{object}	map[string]string
//	@Failure		500			{object}	map[string]string
//	@Router			/runners/{runnerId} [get]
//
//	@id				Get
func Get(ctx *gin.Context) {
	runnerId := ctx.Param("runnerId")
	log.Infof("Getting runner instance: %s", runnerId)

	// Get the provider manager
	providerManager := provider.GetInstance()
	runnerProvider, err := providerManager.GetProvider()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get runner using the provider
	runner, err := runnerProvider.GetRunner(ctx.Request.Context(), runnerId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, runner)
}
