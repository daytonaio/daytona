package workspace

import (
	"errors"

	"github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"
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
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Debug(w)

	workspaceInfo, err := provisioner.GetWorkspaceInfo(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
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
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	workspaceInfos := []*types.WorkspaceInfo{}

	for _, workspace := range workspaces {
		workspaceInfo, err := provisioner.GetWorkspaceInfo(workspace)
		if err != nil {
			log.Error(err)
			ctx.JSON(500, gin.H{"error": err.Error()})
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
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	log.Debug(w)

	err = provisioner.DestroyWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = db.DeleteWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
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
