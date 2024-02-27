package workspace

import (
	"errors"

	"github.com/daytonaio/daytona/common/types"
	"github.com/daytonaio/daytona/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Get workspace info
//	@Description	Get workspace info
//	@Produce		json
//	@Param			workspaceId	path		string	true	"Workspace ID"
//	@Success		200			{object}	dto.Workspace
//	@Router			/workspace/{workspaceId} [get]
//
//	@id				GetWorkspace
func GetWorkspace(ctx *gin.Context) {
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
//
//	@id				ListWorkspaces
func ListWorkspaces(ctx *gin.Context) {

	workspaces, err := db.ListWorkspaces()
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	response := []dto.Workspace{}

	for _, workspace := range workspaces {
		workspaceInfo, err := provisioner.GetWorkspaceInfo(workspace)
		if err != nil {
			log.Error(err)
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
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
