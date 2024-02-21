package workspace

import (
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	err = provisioner.StartWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
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
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	project, err := getProject(w, projectId)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	err = provisioner.StartProject(project)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}
