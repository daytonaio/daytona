package workspace

import (
	"github.com/daytonaio/daytona/server/db"
	"github.com/daytonaio/daytona/server/provisioner"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// StopWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Stop workspace
//	@Description	Stop workspace
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Success		200
//	@Router			/workspace/{workspaceId}/stop [post]
//
//	@id				StopWorkspace
func StopWorkspace(ctx *gin.Context) {
	workspaceId := ctx.Param("workspaceId")

	w, err := db.FindWorkspace(workspaceId)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	err = provisioner.StopWorkspace(w)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}

// StopProject 			godoc
//
//	@Tags			workspace
//	@Summary		Stop project
//	@Description	Stop project
//	@Param			workspaceId	path	string	true	"Workspace ID"
//	@Param			projectId	path	string	true	"Project ID"
//	@Success		200
//	@Router			/workspace/{workspaceId}/{projectId}/stop [post]
//
//	@id				StopProject
func StopProject(ctx *gin.Context) {
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

	err = provisioner.StopProject(project)
	if err != nil {
		log.Error(err)
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(200)
}
