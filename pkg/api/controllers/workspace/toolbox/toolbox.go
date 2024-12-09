// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/daytonaio/daytona/pkg/agent/toolbox"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/gin-gonic/gin"
)

// GetProjectDir 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get project dir
//	@Description	Get project directory
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project Id"
//	@Success		200			{object}	ProjectDirResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/projectdir [get]
//
//	@id				GetProjectDir
func GetProjectDir(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

func forwardRequestToToolbox(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	projectId := ctx.Param("projectId")

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, false)
	if err != nil {
		if workspaces.IsWorkspaceNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	found := false
	for _, p := range w.Projects {
		if p.Name == projectId {
			found = true
			break
		}
	}

	if !found {
		ctx.AbortWithError(http.StatusNotFound, errors.New("project not found"))
		return
	}

	projectHostname := project.GetProjectHostname(w.Id, projectId)

	route := strings.Replace(ctx.Request.URL.Path, fmt.Sprintf("/workspace/%s/%s/toolbox/", workspaceId, projectId), "", 1)
	query := ctx.Request.URL.Query().Encode()

	u := fmt.Sprintf("http://%s:%d/%s?%s", projectHostname, toolbox.PORT, route, query)

	copy := ctx.Copy()

	newUrl, err := url.Parse(u)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	copy.Request.URL = newUrl
	copy.Request.RequestURI = ""

	resp, err := server.TailscaleServer.HTTPClient().Do(copy.Request)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	ctx.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
