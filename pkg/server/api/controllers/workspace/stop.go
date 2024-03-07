package workspace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
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

	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}
	err = provisioner.StopWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop workspace %s: %s", w.Name, err.Error()))
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

	w, err := db.FindWorkspaceByIdOrName(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}

	project, err := getProject(w, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("project not found"))
		return
	}

	err = provisioner.StopProject(project)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to stop project %s: %s", project.Name, err.Error()))
		return
	}

	ctx.Status(200)
}
