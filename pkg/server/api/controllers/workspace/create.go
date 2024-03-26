package workspace

import (
	"fmt"
	"net/http"

	"github.com/daytonaio/daytona/pkg/server/api/controllers/workspace/dto"
	"github.com/daytonaio/daytona/pkg/server/workspaceservice"
	"github.com/gin-gonic/gin"
)

// CreateWorkspace 			godoc
//
//	@Tags			workspace
//	@Summary		Create a workspace
//	@Description	Create a workspace
//	@Param			workspace	body	CreateWorkspace	true	"Create workspace"
//	@Produce		json
//	@Success		200	{object}	Workspace
//	@Router			/workspace [post]
//
//	@id				CreateWorkspace
func CreateWorkspace(ctx *gin.Context) {
	var createWorkspaceDto dto.CreateWorkspace
	err := ctx.BindJSON(&createWorkspaceDto)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("invalid request body: %s", err.Error()))
		return
	}

	w, err := workspaceservice.CreateWorkspace(createWorkspaceDto)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to create workspace: %s", err.Error()))
		return
	}

	ctx.JSON(200, w)
}
