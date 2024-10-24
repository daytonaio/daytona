// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package target

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
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
	server := server.GetInstance(nil)

	targets, err := server.ProviderTargetService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list targets: %w", err))
		return
	}

	for _, target := range targets {
		p, err := server.ProviderManager.GetProvider(target.ProviderInfo.Name)
		if err != nil {
			target.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		manifest, err := (*p).GetTargetManifest()
		if err != nil {
			target.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		var opts map[string]interface{}
		err = json.Unmarshal([]byte(target.Options), &opts)
		if err != nil {
			target.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		for name, property := range *manifest {
			if property.InputMasked {
				delete(opts, name)
			}
		}

		updatedOptions, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			target.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		target.Options = string(updatedOptions)
	}

	ctx.JSON(200, targets)
}
