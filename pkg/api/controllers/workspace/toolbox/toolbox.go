// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package toolbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
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
	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func forwardRequestToToolbox(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	server := server.GetInstance(nil)

	w, err := server.WorkspaceService.FindWorkspace(ctx.Request.Context(), workspaceId, services.WorkspaceRetrievalParams{})
	if err != nil {
		if stores.IsWorkspaceNotFound(err) {
			ctx.AbortWithError(http.StatusNotFound, err)
			return
		}
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var client *http.Client
	var websocketDialer *websocket.Dialer

	workspaceHostname := common.GetTailscaleHostname(w.Id)
	route := strings.Replace(ctx.Request.URL.Path, fmt.Sprintf("/workspace/%s/toolbox/", workspaceId), "", 1)
	query := ctx.Request.URL.Query().Encode()

	scheme := "http"
	if ctx.Request.Header.Get("Upgrade") == "websocket" {
		scheme = "ws"
	}
	reqUrl := fmt.Sprintf("%s://%s:%d/%s?%s", scheme, workspaceHostname, config.TOOLBOX_API_PORT, route, query)
	client = server.TailscaleServer.HTTPClient()
	websocketDialer = &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			return server.TailscaleServer.Dial(ctx.Request.Context(), network, addr)
		},
	}

	if w.TargetId == "local" && w.ProviderMetadata != nil && *w.ProviderMetadata != "" {
		var metadata map[string]interface{}
		err := json.Unmarshal([]byte(*w.ProviderMetadata), &metadata)
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
