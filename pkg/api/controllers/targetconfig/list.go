// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package targetconfig

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server"
	"github.com/gin-gonic/gin"
)

// ListTargetConfigs godoc
//
//	@Tags			target-config
//	@Summary		List target configs
//	@Description	List target configs
//	@Produce		json
//	@Success		200	{array}	TargetConfig
//	@Router			/target-config [get]
//
//	@id				ListTargetConfigs
func ListTargetConfigs(ctx *gin.Context) {
	server := server.GetInstance(nil)

	targetConfigs, err := server.TargetConfigService.List(nil)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to list target configs: %w", err))
		return
	}

	for _, targetConfig := range targetConfigs {
		p, err := server.ProviderManager.GetProvider(targetConfig.ProviderInfo.Name)
		if err != nil {
			targetConfig.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		manifest, err := (*p).GetTargetConfigManifest()
		if err != nil {
			targetConfig.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		var opts map[string]interface{}
		err = json.Unmarshal([]byte(targetConfig.Options), &opts)
		if err != nil {
			targetConfig.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		for name, property := range *manifest {
			if property.InputMasked {
				delete(opts, name)
			}
		}

		updatedOptions, err := json.MarshalIndent(opts, "", "  ")
		if err != nil {
			targetConfig.Options = fmt.Sprintf("Error: %s", err.Error())
			continue
		}

		targetConfig.Options = string(updatedOptions)
	}

	ctx.JSON(200, targetConfigs)
}
