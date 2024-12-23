// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/pkg/agent/toolbox/config"
	"github.com/daytonaio/daytona/pkg/common"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/services"
	"github.com/daytonaio/daytona/pkg/stores"
	"github.com/gin-gonic/gin"
)

// GetWorkspaceDir 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get workspace dir
//	@Description	Get workspace directory
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Success		200			{object}	WorkspaceDirResponse
//	@Router			/workspace/{workspaceId}/toolbox/workspace-dir [get]
//
//	@id				GetWorkspaceDir
func GetWorkspaceDir(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

func forwardRequestToToolbox(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{})
	if err != nil {
		if stores.IsWorkspaceNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var client *http.Client

	workspaceHostname := common.GetTailscaleHostname(w.Id)
	route := strings.Replace(ctx.Request.URL.Path, fmt.Sprintf("/workspace/%s/toolbox/", workspaceId), "", 1)
	query := ctx.Request.URL.Query().Encode()

	reqUrl := fmt.Sprintf("http://%s:%d/%s?%s", workspaceHostname, config.TOOLBOX_API_PORT, route, query)
	client = server.TailscaleServer.HTTPClient()

	if w.TargetId == "local" && w.ProviderMetadata != nil && *w.ProviderMetadata != "" {
		var metadata map[string]interface{}
		err := json.Unmarshal([]byte(*w.ProviderMetadata), &metadata)
		if err == nil {
			if toolboxPortString, ok := metadata["daytona.toolbox.api.hostPort"]; ok {
				toolboxPort, err := strconv.ParseUint(toolboxPortString.(string), 10, 16)
				if err == nil {
					client = http.DefaultClient
					reqUrl = fmt.Sprintf("http://localhost:%d/%s?%s", toolboxPort, route, query)
				}
			}
		}
	}

	copy := ctx.Copy()

	newUrl, err := url.Parse(reqUrl)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	copy.Request.URL = newUrl
	copy.Request.RequestURI = ""

	resp, err := client.Do(copy.Request)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
