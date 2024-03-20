package workspace

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/db"
	"github.com/daytonaio/daytona/pkg/server/provisioner"
	"github.com/daytonaio/daytona/pkg/types"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Get workspace info
//	@Description	Get workspace info
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID or Name"
//	@Success		200			{object}	dto.Workspace
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				GetWorkspace
func GetWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	w, err := db.FindWorkspaceByIdOrName(workspaceId)
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

	response := dto.Workspace{
		Workspace: *w,
		Info:      workspaceInfo,
	}

	ctx.JSON(200, response)
}

// ListWorkspaces 			godoc
//
//	@Tags			workspace
//	@Summary		List workspaces
//	@Description	List workspaces
//	@Produce		json
//	@Success		200	{array}	dto.Workspace
//	@Router			/workspace [get]
//	@Param			verbose	query	bool	false	"Verbose"
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {
	workspaces, err := db.ListWorkspaces()
	verbose := ctx.Query("verbose")

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("failed to list workspaces"))
		return
	}

	response := []dto.Workspace{}

	for _, workspace := range workspaces {
		var workspaceInfo *types.WorkspaceInfo
		if verbose != "" {
			isVerbose, err := strconv.ParseBool(verbose)
			if err != nil {
				ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid value for verbose flag"))
				return
			}
	
			if isVerbose {
				workspaceInfo, err = provisioner.GetWorkspaceInfo(workspace)
				if err != nil {
					log.Error(fmt.Errorf("failed to get workspace info for %s", workspace.Name))
					// return
				}
			}
		}
			

		response = append(response, dto.Workspace{
			Workspace: *workspace,
			Info:      workspaceInfo,
		})
	}

	ctx.JSON(200, response)
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

	w, err := db.FindWorkspaceByIdOrName(workspaceId)
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
