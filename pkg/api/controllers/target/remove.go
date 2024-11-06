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
