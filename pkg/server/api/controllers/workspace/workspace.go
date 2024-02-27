package workspace

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWorkspaceInfo 			godoc
//
//	@Tags			workspace
//	@Summary		Get workspace info
//	@Description	Get workspace info
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID"
//	@Success		200			{object}	WorkspaceInfo
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				GetWorkspaceInfo
func GetWorkspaceInfo(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	w, err := db.FindWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}

	log.Debug(w)

	workspaceInfo, err := provisioner.GetWorkspaceInfo(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace info: %s", err.Error()))
		return
	}

	ctx.JSON(200, workspaceInfo)
}

// ListWorkspaces 			godoc
//
//	@Tags			workspace
//	@Summary		List workspaces info
//	@Description	List workspaces info
//	@Produce		json
//	@Success		200	{array}	Workspace
//	@Router			/workspace [get]
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {
	workspaces, err := db.ListWorkspaces()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to list workspaces"))
		return
	}

	workspaceInfos := []*types.WorkspaceInfo{}

	for _, workspace := range workspaces {
		workspaceInfo, err := provisioner.GetWorkspaceInfo(workspace)
		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to get workspace info for %s", workspace.Name))
			return
		}

		workspaceInfos = append(workspaceInfos, workspaceInfo)
	}

	ctx.JSON(200, workspaceInfos)
}

// RemoveWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Remove workspace
//	@Description	Remove workspace
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Success		200
//	@Router			/workspace/{workspaceId} [delete]
//
//	@id				RemoveWorkspace
func RemoveWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	w, err := db.FindWorkspace(workspaceId)
	if err != nil {
		ctx.AbortWithError(http.StatusNotFound, errors.New("workspace not found"))
		return
	}

	log.Debug(w)

	err = provisioner.DestroyWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to destroy workspace: %s", err.Error()))
		return
	}

	err = db.DeleteWorkspace(w)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to delete workspace: %s", err.Error()))
		return
	}

	ctx.Status(200)
}

func getProject(workspace *types.Workspace, projectName string) (*types.Project, error) {
	for _, project := range workspace.Projects {
		if project.Name == projectName {
			return project, nil
		}
	}
	return nil, errors.New("project not found")
}
