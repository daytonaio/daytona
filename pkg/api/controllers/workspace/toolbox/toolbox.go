// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/daytonaio/daytona/pkg/agent/toolbox/config"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/workspaces"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// GetProjectDir 			godoc
//
//	@Tags			workspace toolbox
//	@Summary		Get project dir
//	@Description	Get project directory
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Param			projectId	path		string	true	"Project ID"
//	@Success		200			{object}	ProjectDirResponse
//	@Router			/workspace/{workspaceId}/{projectId}/toolbox/project-dir [get]
//
//	@id				GetProjectDir
func GetProjectDir(ctx *gin.Context) {
	forwardRequestToToolbox(ctx)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func forwardRequestToToolbox(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	projectId := ctx.Param("projectId")

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.GetWorkspace(ctx.Request.Context(), workspaceId, true)
	if err != nil {
		if workspaces.IsWorkspaceNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var projectInfo *project.ProjectInfo
	found := false
	for _, p := range w.Info.Projects {
		if p.Name == projectId {
			projectInfo = p
			found = true
			break
		}
	}

	if !found {
		ctx.AbortWithError(http.StatusNotFound, errors.New("project not found"))
		return
	}

	var client *http.Client
	var websocketDialer *websocket.Dialer

	projectHostname := project.GetProjectHostname(w.Id, projectId)
	route := strings.Replace(ctx.Request.URL.Path, fmt.Sprintf("/workspace/%s/%s/toolbox/", workspaceId, projectId), "", 1)
	query := ctx.Request.URL.Query().Encode()

	scheme := "http"
	if ctx.Request.Header.Get("Upgrade") == "websocket" {
		scheme = "ws"
	}

	reqUrl := fmt.Sprintf("%s://%s:%d/%s?%s", scheme, projectHostname, config.TOOLBOX_API_PORT, route, query)
	client = server.TailscaleServer.HTTPClient()
	websocketDialer = &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return server.TailscaleServer.Dial(ctx.Request.Context(), network, addr)
		},
	}

	if w.Target == "local" {
		var metadata map[string]interface{}
		err := json.Unmarshal([]byte(projectInfo.ProviderMetadata), &metadata)
		if err == nil {
			if toolboxPortString, ok := metadata["daytona.toolbox.api.hostPort"]; ok {
				toolboxPort, err := strconv.ParseUint(toolboxPortString.(string), 10, 16)
				reqUrl = fmt.Sprintf("%s://localhost:%d/%s?%s", scheme, toolboxPort, route, query)

				if err == nil {
					if scheme == "ws" {
						websocketDialer = websocket.DefaultDialer
					} else {
						client = http.DefaultClient
					}
				}
			}
		}
	}

	copy := ctx.Copy()

	if scheme == "ws" {
		ws, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer ws.Close()

		conn, _, err := websocketDialer.DialContext(ctx, reqUrl, nil)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()

		go func() {
			io.Copy(ws.NetConn(), conn.NetConn())
		}()

		io.Copy(conn.NetConn(), ws.NetConn())
		return
	}

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
