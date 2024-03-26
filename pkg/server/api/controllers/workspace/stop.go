package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server/workspaceservice"
	"github.com/gin-gonic/gin"
)

// StopWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Stop workspace
//	@Description	Stop workspace
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Success		200
//	@Router			/workspace/{workspaceId}/stop [post]
//
//	@id				StopWorkspace
func StopWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	err := workspaceservice.StopWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop workspace %s: %s", workspaceId, err.Error()))
		return
	}

	ctx.Status(200)
}

// StopProject 			godoc
//
//	@Tags			workspace
//	@Summary		Stop project
//	@Description	Stop project
//	@Param			workspaceId	path	string	true	"Workspace ID or Name"
//	@Param			projectId	path	string	true	"Project ID"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/stop [post]
//
//	@id				StopProject
func StopProject(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	projectId := ctx.Param("projectId")

	err := workspaceservice.StopProject(workspaceId, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop project %s: %s", projectId, err.Error()))
		return
	}

	ctx.Status(200)
}
