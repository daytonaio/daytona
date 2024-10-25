// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server"
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

	w, err := server.TargetService.GetTarget(ctx.Request.Context(), targetId, verbose)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get target: %w", err))
		return
	}

	ctx.JSON(200, w)
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

	targetList, err := server.TargetService.ListTargets(ctx.Request.Context(), verbose)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list targets: %w", err))
		return
	}

	ctx.JSON(200, targetList)
}

// RemoveTarget 			godoc
//
//	@Tags			target
//	@Summary		Remove target
//	@Description	Remove target
//	@Param			targetId	path	string	true	"Target ID"
//	@Param			force		query	bool	false	"Force"
//	@Success		200
//	@Router			/target/{targetId} [delete]
//
//	@id				RemoveTarget
func RemoveTarget(ctx *gin.Context) {
	targetId := ctx.Param("targetId")
	forceQuery := ctx.Query("force")
	var err error
	force := false

	if forceQuery != "" {
		force, err = strconv.ParseBool(forceQuery)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for force flag"))
			return
		}
	}

	server := server.GetInstance(nil)

	if force {
		err = server.TargetService.ForceRemoveTarget(ctx.Request.Context(), targetId)
	} else {
		err = server.TargetService.RemoveTarget(ctx.Request.Context(), targetId)
	}

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to remove target: %w", err))
		return
	}

	ctx.Status(200)
}
