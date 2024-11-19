// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/api/util"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetTarget 			godoc
//
//	@Tags			target
//	@Summary		Get target info
//	@Description	Get target info
//	@Produce		json
//	@Param			targetId	path		string	true	"Target ID or Name"
//	@Param			verbose		query		bool	false	"Verbose"
//	@Success		200			{object}	TargetDTO
//	@Router			/target/{targetId} [get]
//
//	@id				GetTarget
func GetTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	verboseQuery := ctx.Query("verbose")
	verbose := false
	var err error

	if verboseQuery != "" {
		verbose, err = strconv.ParseBool(verboseQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	t, err := server.TargetService.GetTarget(ctx.Request.Context(), &stores.TargetFilter{IdOrName: &targetId}, services.TargetRetrievalParams{Verbose: verbose})
	if err != nil {
		statusCode := http.StatusInternalServerError
		if stores.IsTargetNotFound(err) || services.IsTargetDeleted(err) {
			statusCode = http.StatusNotFound
		}
		ctx.AbortWithError(statusCode, fmt.Errorf("failed to find target: %w", err))
		return
	}

	maskedOptions, err := util.GetMaskedOptions(server, t.ProviderInfo.Name, t.Options)
	if err != nil {
		t.Options = fmt.Sprintf("Error: %s", err.Error())
	} else {
		t.Options = maskedOptions
	}

	util.HideDaytonaEnvVars(&t.EnvVars)

	ctx.JSON(200, t)
}

// ListTargets 			godoc
//
//	@Tags			target
//	@Summary		List targets
//	@Description	List targets
//	@Produce		json
//	@Success		200	{array}	TargetDTO
//	@Router			/target [get]
//	@Param			verbose	query	bool	false	"Verbose"
//
//	@id				ListTargets
func ListTargets(ctx *gin.Context) {
	verboseQuery := ctx.Query("verbose")
	verbose := false
	var err error

	if verboseQuery != "" {
		verbose, err = strconv.ParseBool(verboseQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	targetList, err := server.TargetService.ListTargets(ctx.Request.Context(), nil, services.TargetRetrievalParams{Verbose: verbose})
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list targets: %w", err))
		return
	}

	for i, t := range targetList {
		maskedOptions, err := util.GetMaskedOptions(server, t.ProviderInfo.Name, t.Options)
		if err != nil {
			targetList[i].Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		targetList[i].Options = maskedOptions
		util.HideDaytonaEnvVars(&targetList[i].EnvVars)
	}

	ctx.JSON(200, targetList)
}
