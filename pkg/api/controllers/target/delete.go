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

// DeleteTarget 			godoc
//
//	@Tags			target
//	@Summary		Delete target
//	@Description	Delete target
//	@Param			targetId	path	string	true	"Target ID"
//	@Param			force		query	bool	false	"Force"
//	@Success		200
//	@Router			/target/{targetId} [delete]
//
//	@id				DeleteTarget
func DeleteTarget(ctx *gin.Context) {
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
		err = server.TargetService.ForceDeleteTarget(ctx.Request.Context(), targetId)
	} else {
		err = server.TargetService.DeleteTarget(ctx.Request.Context(), targetId)
	}

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to delete target: %w", err))
		return
	}

	ctx.Status(200)
}
