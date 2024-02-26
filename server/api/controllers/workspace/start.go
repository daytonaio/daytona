package workspace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"
	"github.com/gin-gonic/gin"
)

// StartWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Start workspace
//	@Description	Start workspace
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Success		200
//	@Router			/workspace/{workspaceId}/start [post]
//
//	@id				StartWorkspace
func StartWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	w, err := db.FindWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}
	err = provisioner.StartWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to start workspace %s: %s", w.Name, err.Error()))
		return
	}

	ctx.Status(200)
}

// StartProject 			godoc
//
//	@Tags			workspace
//	@Summary		Start project
//	@Description	Start project
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Param			projectId	path	string	true	"Project ID"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/start [post]
//
//	@id				StartProject
func StartProject(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")
	projectId := ctx.Param("projectId")

	w, err := db.FindWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}

	project, err := getProject(w, projectId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("project not found"))
		return
	}

	err = provisioner.StartProject(project)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to start project %s: %s", project.Name, err.Error()))
		return
	}

	ctx.Status(200)
}
